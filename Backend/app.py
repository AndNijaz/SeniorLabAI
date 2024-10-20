from readability import Document
import requests
from bs4 import BeautifulSoup
import openai
import json
from dotenv import load_dotenv
import os
from datetime import datetime
from googlesearch import search
from flask import Flask, request, jsonify
import html2text
from concurrent.futures import ThreadPoolExecutor, as_completed
import logging
import traceback
import threading
import time
from pydantic import BaseModel
from flask_cors import CORS
import pprint

# Set up logging
logging.basicConfig(
    level=logging.DEBUG,
    filename="app.log",
    encoding="utf-8",
    filemode="w",
    format="{asctime} - {levelname} - {message}",
    style="{",
    datefmt="%Y-%m-%d %H:%M",
)

tokens = 0
tokensoutput = 0
load_dotenv()
openai.api_key = os.getenv("OPENAI_KEY")

def scrape_webpage(url):
    try:
        if not url.startswith(("https://", "http://")):
            url = "https://" + url
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        html_content = response.text
        text_maker = html2text.HTML2Text()
        text_maker.ignore_links = False
        text_maker.ignore_images = True
        text = text_maker.handle(html_content)
        return text
    except (requests.exceptions.Timeout, requests.exceptions.HTTPError, Exception) as e:
        logging.error(f"Error scraping URL {url}: {e}")
        return None

def extract_date(soup):
    date = None
    time_tag = soup.find("time")
    if time_tag and time_tag.get("datetime"):
        date = time_tag.get("datetime")
    elif time_tag:
        date = time_tag.get_text()
    if not date:
        meta_date = (
            soup.find("meta", {"property": "article:published_time"}) or
            soup.find("meta", {"name": "article:published_time"}) or
            soup.find("meta", {"name": "publish_date"}) or
            soup.find("meta", {"name": "pubdate"}) or
            soup.find("meta", {"property": "og:pubdate"}) or
            soup.find("meta", {"property": "og:updated_time"})
        )
        if meta_date:
            date = meta_date.get("content")
    if not date:
        date_span = (
            soup.find("span", class_="date") or
            soup.find("span", class_="publish-date") or
            soup.find("div", class_="date") or
            soup.find("div", class_="publish-date")
        )
        if date_span:
            date = date_span.get_text()
    return date

def google_search(query):
    logging.info(f"Search Query: {query}")
    links = []
    successful_scrapes = 0
    max_successful_scrapes = 3
    max_workers = 5
    search_results_limit = 50
    scrape_counter_lock = threading.Lock()

    search_results = list(search(query, num_results=search_results_limit))

    def process_result(j):
        nonlocal successful_scrapes
        try:
            with scrape_counter_lock:
                if successful_scrapes >= max_successful_scrapes:
                    return None
            headers = {
                "User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/36.0 Mobile/15E148 Safari/605.1.15"
            }
            max_retries = 3
            backoff_factor = 2
            for attempt in range(max_retries):
                try:
                    response = requests.get(j, headers=headers, timeout=10)
                    response.raise_for_status()
                    break
                except (requests.exceptions.ConnectionError, requests.exceptions.Timeout) as e:
                    logging.warning(
                        f"Attempt {attempt + 1}/{max_retries} failed for URL {j}: {e}"
                    )
                    if attempt < max_retries - 1:
                        sleep_time = backoff_factor**attempt
                        time.sleep(sleep_time)
                        continue
                    else:
                        logging.error(
                            f"Failed to retrieve URL {j} after {max_retries} attempts."
                        )
                        return None
            doc = Document(response.text)
            title = doc.title()
            body = scrape_webpage(j)
            if body is None:
                return None
            full_soup = BeautifulSoup(response.text, "html.parser")
            article_date = extract_date(full_soup)
            result = {
                "title": title,
                "body": body,
                "date": article_date,
                "link": j
            }
            with scrape_counter_lock:
                if successful_scrapes < max_successful_scrapes:
                    successful_scrapes += 1
                    return result
            return None
        except requests.exceptions.HTTPError as http_err:
            logging.error(f"HTTP error while accessing URL {j}: {http_err}")
            return None
        except Exception as err:
            logging.error(f"Error occurred while processing URL {j}: {err}")
            logging.debug(traceback.format_exc())
            return None

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = [executor.submit(process_result, j) for j in search_results]

        for future in as_completed(futures):
            result = future.result()
            if result:
                links.append(result)
            with scrape_counter_lock:
                if successful_scrapes >= max_successful_scrapes:
                    break
    return json.dumps(links, indent=4)

class Response(BaseModel):
    longresponse: str
    shortresponse: str
    title: str
tools = [
    {
        "type": "function",
        "function": {
            "name": "search_google",
            "description": "Search google regarding a certain topic",
            "strict": True,
            "parameters": {
                "type": "object",
                "properties": {
                    "search": {
                        "type": "string",
                        "description": "The search parameter that you will use in Google.",
                    }
                },
                "required": ["search"],
                "additionalProperties": False,
            },
        },
    }
]
app = Flask(__name__)
CORS(app)
def chat_completion_request(messages, tools):
    global tokens
    global tokensoutput
    response = openai.beta.chat.completions.parse(
        model="gpt-4o-mini", messages=messages, tools=tools,
        response_format=Response
    )
    tokens += int(response.usage.prompt_tokens)
    tokensoutput += int(response.usage.completion_tokens)
    
    # Convert the response to a dictionary before returning
    return response.dict()

def chat_with_tools(messages, tools):
    global tokensoutput
    global tokens

    try:
        truthvalue = False
        while True:
            response = chat_completion_request(messages, tools)
            tool_calls = response['choices'][0]['message'].get('tool_calls', [])

            if tool_calls:
                tool_call = tool_calls[0]  # Expecting a list, use the first item
                print("Tool Call:", tool_call)  # For debugging purposes

                function_call = tool_call.get('function', {})
                function_name = function_call.get('name', '')

                if function_name == "search_google":
                    truthvalue = True

                    # Use parsed_arguments directly
                    parsed_arguments = function_call.get('parsed_arguments', {})
                    prompt_to_scrape = parsed_arguments.get("search", "")
                    
                    if prompt_to_scrape:
                        scraping_result = google_search(prompt_to_scrape)
                        messages.append(
                            {
                                "role": "assistant",
                                "content": f"Scraping result: {scraping_result}",
                            }
                        )
            else:
                return {
                    "content": response['choices'][0]['message']['parsed'],
                    "internet_search": truthvalue
                }

    except Exception as e:
        logging.error(f"An error occurred: {e}")
        return {"error": "An error occurred"}

# Rest of the code remains the same



@app.route("/", methods=["POST"])
def index():
    if request.method == "POST":
        logging.info("Client connected")
        text_input = request.get_json()
        logging.debug("Original input: " + pprint.pformat(text_input))
        data = text_input.get("text", "")
        logging.debug("Parsed input: " + data)
        date = datetime.today().strftime("%Y-%m")
        
        messages = [
            {"role": "system", "content": f"Ti si inteligenti pomagac koji samo odgovara na srpskom/bosanskom jeziku. Takodjer pazi za mjesece, ne moj pisati lipanj, nego pisi juni na primjer. OVO JE OBAVEZNO NEZAVISNO STA TI JE NA ULAZU. Ako bilo gdje u tvojim podatcima pisu mjeseci kao na primjer prosinac, to trebas da prevedes na decembar i tako isto za svaki drugi mjesec. Samo napisi text, nemoj nikakvog formatiranja dodati. Takodjer nemoj ni dodavati nove linije, samo cisti tekst. Trenutni datum je {date}\
            Ako je potrebno da se ovo tacno odgovori, mozes pozvati funkciju search_google da bi nasao vise informacija. longresponse, kad izbacis koristi html, ne moras cijeli kod, samo tagove za to za sta je vezano, nova linija kad treba sa <br>, boldirani text sa <b>, itallics, i sve slicno tome. Ne mozes koristiti <script> tag"},
            {"role": "user", "content": data}
        ]
        
        result = chat_with_tools(messages, tools)
        
        if isinstance(result, dict):
            logging.info(f"Input tokens: {tokens}")
            logging.info(f"Input price: {(tokens/1000)*0.000150}$")
            logging.info(f"Output tokens: {tokensoutput}")
            logging.info(f"Output price: {(tokensoutput/1000)*0.000600}$")
            logging.info(
                f"Total price: {((tokens/1000)*0.000150)+((tokensoutput/1000)*0.000600)}$"
            )
            logging.info("Result created")
            
            price_info = {
                "input_price": f"{(tokens/1000)*0.000150}$",
                "output_price": f"{(tokensoutput/1000)*0.000600}$",
                "total_price": f"{((tokens/1000)*0.000150)+((tokensoutput/1000)*0.000600)}$",
            }
            result["price_info"] = price_info
            logging.debug("Result: " + pprint.pformat(result))
            return jsonify(result)
        else:
            return jsonify({"error": "Failed to process request."})

if __name__ == "__main__":
    app.run(debug=True)

