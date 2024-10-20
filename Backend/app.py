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
import tiktoken
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

def check_for_illegal_content(text, user_ip):
    """
    This function uses OpenAI's moderation API to check if the provided text 
    contains any illegal content. If illegal content is detected, it logs the 
    user's IP address.

    :param text: Text to be checked for illegal content
    :param user_ip: IP address of the user who sent the request
    :return: True if illegal content is detected, False otherwise
    """
    try:
        # Send a request to OpenAI's moderation API to analyze the input text
        response = openai.moderations.create(
            model="omni-moderation-latest",
            input=text,
        )
        
        # Log the raw moderation response
        logging.debug(f"Moderation API response: {response}")

        # Extract the flagged status directly from the response object
        if response.results[0].flagged:
            # If flagged, log the IP address in the illegal-activity.log file
            logging.error(f"Illegal content detected from IP: {user_ip}")
            with open("illegal-activity.log", "a") as log_file:
                log_file.write(f"{datetime.now()} - IP: {user_ip} - Content: {text}\n")
            return True
    except Exception as e:
        # Log any errors encountered during moderation checks
        logging.error(f"Error checking for illegal content: {e}")
        logging.debug(traceback.format_exc())

    return False


def trim_messages(messages, max_tokens):
    # Initialize tiktoken model. You might need to specify your OpenAI model here.
    encoder = tiktoken.encoding_for_model("gpt-4")
    
    total_tokens = 0
    trimmed_messages = []
    
    for message in reversed(messages):
        message_tokens = len(encoder.encode(json.dumps(message)))  # Count tokens for the message
        if total_tokens + message_tokens > max_tokens:
            break
        trimmed_messages.append(message)
        total_tokens += message_tokens
    
    # Reverse to maintain original message order
    return list(reversed(trimmed_messages))

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
    max_tokens = 128000  # Set buffer for model's response
    
    # Trim messages if they exceed the limit
    trimmed_messages = trim_messages(messages, max_tokens)

    response = openai.beta.chat.completions.parse(
        model="gpt-4o-mini", messages=trimmed_messages, tools=tools,
        response_format=Response
    )
    tokens += int(response.usage.prompt_tokens)
    tokensoutput += int(response.usage.completion_tokens)
    
    # Convert the response to a dictionary before returning
    return response.dict()

def chat_with_tools(messages, tools):
    global tokensoutput
    global tokens
    captured_links = []

    try:
        truthvalue = False
        while True:
            response = chat_completion_request(messages, tools)
            tool_calls = response['choices'][0]['message'].get('tool_calls', [])

            if tool_calls:
                tool_call = tool_calls[0]
                function_call = tool_call.get('function', {})
                function_name = function_call.get('name', '')

                if function_name == "search_google":
                    truthvalue = True
                    parsed_arguments = function_call.get('parsed_arguments', {})
                    prompt_to_scrape = parsed_arguments.get("search", "")
                    
                    if prompt_to_scrape:
                        scraping_result = json.loads(google_search(prompt_to_scrape))
                        # Extract and capture links from the scraping result
                        for result in scraping_result:
                            if 'link' in result:
                                captured_links.append(result['link'])
                        messages.append(
                            {
                                "role": "assistant",
                                "content": f"Scraping result: {json.dumps(scraping_result)}",
                            }
                        )
            else:
                return {
                    "content": response['choices'][0]['message']['parsed'],
                    "internet_search": truthvalue,
                    "used_links": captured_links  # Include the used links in the result
                }

    except Exception as e:
        logging.error(f"An error occurred: {e}")
        return {"error": "An error occurred"}

@app.route("/", methods=["POST"])
def index():
    if request.method == "POST":
        logging.info("Client connected")
        text_input = request.get_json()  # Retrieve JSON data from the client
        logging.debug("Original input: " + pprint.pformat(text_input))
        data = text_input.get("text", "")  # Extract the 'text' field from JSON
        logging.debug("Parsed input: " + data)

        # Retrieve the user's IP address from the request's X-Forwarded-For header (if present) or the remote address
        user_ip = request.headers.get('X-Forwarded-For', request.remote_addr)
        logging.debug(f"User IP: {user_ip}")
        # Check if the input contains any illegal content using the moderation API
        if check_for_illegal_content(data, user_ip):
            # If illegal content is detected, return a standardized message
            return jsonify({
                "content": {
                    "longresponse": "Illegal content detected.",
                    "shortresponse": "Illegal content detected.",
                    "title": "Illegal content detected.",
                },
                "price_info": {
                    "input_price": "0$",
                    "output_price": "0$",
                    "total_price": "0$"
                }
            })

        date = datetime.today().strftime("%Y-%m")

        messages = [
            {
                "role": "system",
                "content": (
                    f"Ti si inteligenti pomagac koji samo odgovara na srpskom/bosanskom jeziku. "
                    f"Pazi da koristiš nazive meseci na srpskom, na primer, koristi 'juni' umesto 'lipanj'. "
                    f"Trenutni datum je {date}. "
                    "Ako je potrebno da se ovo tačno odgovori, možeš pozvati funkciju search_google da bi našao više informacija. "
                    "U odgovoru pod nazivom 'longresponse', koristi HTML za formatiranje. "
                    "Formatiraj tekst koristeći tagove kao što su <br> za nove linije, <b> za podebljani tekst, <em> za italik itd. "
                    "Dodaj izvore kao naslov koji se može kliknuti koristeći <a> tag sa atributom href. "
                    "Nikada ne koristi HTML u odgovoru pod nazivom 'shortresponse'. Samo koristi čisti tekst bez dodatnog formatiranja. "
                    "Za shortcontent imas limit od 50 rijeci, a u longcontent mozes napisati najvise 200 rijeci"
                )
            },
            {"role": "user", "content": data}
        ]

        result = chat_with_tools(messages, tools)

        if isinstance(result, dict):
            # Append used links to the messages for context
            messages.append({
                "role": "system",
                "content": f"Ovo su izvori koji su korišteni: {', '.join(result.get('used_links', []))}"
            })

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

