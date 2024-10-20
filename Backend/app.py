from readability import Document
import requests
from bs4 import BeautifulSoup
import openai
import json
from dotenv import load_dotenv
import os
from datetime import datetime
from googlesearch import search
from flask import Flask, request
import html2text
from concurrent.futures import ThreadPoolExecutor, as_completed
import logging
import traceback
import threading
import time

# This set's up the app.log for logging, and because of filemode="w" it deletes every restart of the program
# Have to figure out a way to auto clean the log if it is ran for a longer amount of time so it doesn't get too big
logging.basicConfig(
    level=logging.DEBUG,
    filename="app.log",
    encoding="utf-8",
    filemode="w",
    format="{asctime} - {levelname} - {message}",
    style="{",
    datefmt="%Y-%m-%d %H:%M",
)
# Variables used to get the usage statistics, as well as to calculate price
tokens = 0
tokensoutput = 0

load_dotenv()
# Code to get the api key from the .env file. You will need to create this yourself with the provided key

openai.api_key = os.getenv("OPENAI_KEY")


def scrape_webpage(url, data):

    # Scrapes the content of a webpage and returns the text.
    try:

        # Checks if it is a valid link
        if not url.startswith(("https://", "http://")):
            url = "https://" + url
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        # Get's the html file and uses that to get the content using html2text
        html_content = response.text
        text_maker = html2text.HTML2Text()
        text_maker.ignore_links = False  # Optionally ignore links
        text_maker.ignore_images = True  # Optionally ignore images
        text = text_maker.handle(html_content)
        # Debugger to get info inside the app.log file
        return text

    # Error checking, so if it fails for some reason, it returns "False" so that it can be skipped
    except requests.exceptions.Timeout:
        logging.error(f"Timeout error while scraping URL: {url}")
        return "Fail"
    except requests.exceptions.HTTPError as http_err:
        logging.error(f"HTTP error while scraping URL {url}: {http_err}")
        return "Fail"
    except Exception as err:
        logging.error(f"Error occurred while scraping URL {url}: {err}")
        logging.debug(traceback.format_exc())
        return "Fail"


# This part of the code checks the upload date of the article or the website that is being scraped, which is provided to the AI


def extract_date(soup):
    date = None
    time_tag = soup.find("time")
    if time_tag and time_tag.get("datetime"):
        date = time_tag.get("datetime")
    elif time_tag:
        date = time_tag.get_text()

    if not date:
        meta_date = (
            soup.find("meta", {"property": "article:published_time"})
            or soup.find("meta", {"name": "article:published_time"})
            or soup.find("meta", {"name": "publish_date"})
            or soup.find("meta", {"name": "pubdate"})
            or soup.find("meta", {"property": "og:pubdate"})
            or soup.find("meta", {"property": "og:updated_time"})
        )
        if meta_date:
            date = meta_date.get("content")

    if not date:
        date_span = (
            soup.find("span", class_="date")
            or soup.find("span", class_="publish-date")
            or soup.find("div", class_="date")
            or soup.find("div", class_="publish-date")
        )
        if date_span:
            date = date_span.get_text()

    return date


def google_search(query, data):

    # Performs a Google search and returns the top results with summaries.
    # Continues until it successfully scrapes 5 websites without errors.
    logging.info(f"Search Query: {query}")

    links = []
    # Counter for the number of websites it has done so far
    successful_scrapes = 0
    # How many websites it needs to complete to finish the function
    max_successful_scrapes = 3
    # How many threads, the amount of these functions that can run at the same time
    max_workers = 5
    # Maximum number of results to search before giving up completely
    search_results_limit = 50

    # Thread-safe counter, so that it doesn't miscount when counting successful scrapes
    scrape_counter_lock = threading.Lock()

    # Collect a larger number of search results to account for potential errors
    search_results = list(search(query, num_results=search_results_limit))

    # Function to process each search result
    def process_result(j):
        nonlocal successful_scrapes
        try:
            with scrape_counter_lock:
                if successful_scrapes >= max_successful_scrapes:
                    return None  # Exit if we've reached the target

            headers = {
                "User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/36.0  Mobile/15E148 Safari/605.1.15"
            }
            max_retries = 3  # Number of retries for errors
            # Factor how much to slow down request each time, to fix issues like to many requests
            backoff_factor = 2
            for attempt in range(max_retries):
                try:
                    response = requests.get(j, headers=headers, timeout=10)
                    response.raise_for_status()
                    break  # Success, exit the retry loop
                except (
                    requests.exceptions.ConnectionError,
                    requests.exceptions.Timeout,
                ) as e:
                    # Reporting of the errors and amount of erorrs
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
            # Code to get info about the website, like the Title, the content of the website, as well as the article date
            doc = Document(response.text)
            title = doc.title()
            body = scrape_webpage(j, data)
            if body == "Fail":
                return None
            full_soup = BeautifulSoup(response.text, "html.parser")
            article_date = extract_date(full_soup)
            # This is the result that is given to chatGPT for each of the processed page
            result = f"""[[Title]]{title}[[/Title]]\n[[Body]]{body}[[/Body]]\n[[Date]]{
                article_date}[[/Date]]\n[[Link]]{j}[[/Link]]\n\n"""

            with scrape_counter_lock:
                if successful_scrapes < max_successful_scrapes:
                    successful_scrapes += 1
                    return result
                else:
                    return None
        # Error checking for the websites
        except requests.exceptions.HTTPError as http_err:
            logging.error(f"HTTP error while accessing URL {j}: {http_err}")
            return None
        except Exception as err:
            logging.error(f"Error occurred while processing URL {j}: {err}")
            logging.debug(traceback.format_exc())
            return None

    # Use ThreadPoolExecutor to process results in parallel
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = []
        for j in search_results:
            with scrape_counter_lock:
                if successful_scrapes >= max_successful_scrapes:
                    break  # Exit if we've reached the target
            future = executor.submit(process_result, j)
            futures.append(future)

        # Collect the results as they complete
        for future in as_completed(futures):
            result = future.result()
            if result:
                links.append(result)
            with scrape_counter_lock:
                if successful_scrapes >= max_successful_scrapes:
                    break  # Stop collecting results if target is met
    # We use this .join command so that it doesn't throw an error if all pages only returned an error
    return "".join(links)


def chat_completion_request(messages, tools):
    # Sends a request to the OpenAI API to generate a chat response.
    global tokens
    global tokensoutput
    response = openai.chat.completions.create(
        model="gpt-4o-mini", messages=messages, tools=tools
    )
    # Using the data given after the request to get amount of input and output tokens used
    tokens += int(response.usage.prompt_tokens)
    tokensoutput += int(response.usage.completion_tokens)
    return response


def chat_with_tools(messages, tools, data):
    # This is the main code used for the request to the AI
    global tokensoutput
    global tokens

    try:
        truthvalue = False
        # We use this if the AI want's to use the search google function multiple times
        while True:
            response = chat_completion_request(messages, tools)
            tool_calls = response.choices[0].message.tool_calls
            # Checking if the AI requested to use a function
            if tool_calls:
                tool_call = tool_calls[0]
                # If it called the search_google function
                if tool_call.function.name == "search_google":
                    truthvalue = True
                    prompt_to_scrape = json.loads(tool_call.function.arguments)[
                        "search"
                    ]
                    scraping_result = google_search(prompt_to_scrape, data)
                    # Adding the result to the message that is being sent to the AI
                    messages.append(
                        {
                            "role": "assistant",
                            "content": f"Scraping result: {scraping_result}",
                        }
                    )
            # If the AI doesn't request to use functions anymore, this part of the code is called and the output is returned
            # The internet_search value is also returned so we know if the AI used google or not
            else:
                return {
                    "content": response.choices[0].message.content,
                    "internet_search": truthvalue,
                    "price_info": 0,
                }

    except Exception as e:
        print(f"An error occurred: {e}")
        return "An error occurred"


# The json format of how the function works, which is provided to the AI so it knows how to use it
tools = [
    {
        "type": "function",
        "function": {
            "name": "search_google",
            "description": "Search google regarding a certian topic",
            "parameters": {
                "type": "object",
                "properties": {
                    "search": {
                        "type": "string",
                        "description": "The search parameter that you will use in Google.",
                    }
                },
                "required": ["search"],
            },
        },
    }
]
# Flask app, which is used to recieve the request
app = Flask(__name__)


# The root, this is what is called when someone accesses the main site, so whet they access http://localhost:5000/
@app.route("/", methods=["POST"])
def index():
    # Only accepts a POST connection, so we can't open this normally in a web browser
    # We have to use something like curl or an API tester
    # The code under the if statement is called when a connection is made
    if request.method == "POST":
        global tokens
        global tokensoutput
        logging.info("Client connected")
        # This code gets the data from the POST request
        data = request.form.get("text")
        date = datetime.today().strftime("%Y-%m")
        # This is the starting message, here is the system prompt that controls how the AI controls as a whole
        # Here we also give it information about the functions that are available to it
        # Here in the user section we give it the prompt that was passed using the POST request
        messages = [
            {"role": "system", "content": f"Ti si inteligenti pomagac koji samo odgovara na srpskom/bosanskom jeziku. Takodjer pazi za mjesece, ne moj pisati lipanj, nego pisi juni na primjer. OVO JE OBAVEZNO NEZAVISNO STA TI JE NA ULAZU. Ako bilo gdje u tvojim podatcima pisu mjeseci kao na primjer prosinac, to trebas da prevedes na decembar i tako isto za svaki drugi mjesec. Samo napisi text, nemoj nikakvog formatiranja dodati. Takodjer nemoj ni dodavati nove linije, samo cisti tekst. Trenutni datum je {date}\
            Ako je potrebno da se ovo tacno odgovori, mozes pozvati funkciju search_google da bi nasao vise informacija"},
            {"role": "user", "content": data}
        ]
        # Here we call the main function, and where we give it the messages, the available tools, and the passed prompt
        result = chat_with_tools(messages, tools, data)
        # Here we show the tokens that were user, for the input as well as the output, and doing the calculations to get the price of each request
        logging.info(f"Input tokens: {tokens}")
        logging.info(f"Input price: {(tokens/1000)*0.000150}$")
        logging.info(f"Output tokens: {tokensoutput}")
        logging.info(f"Output price: {(tokensoutput/1000)*0.000600}$")
        logging.info(
            f"Total price: {((tokens/1000)*0.000150)+((tokensoutput/1000)*0.000600)}$"
        )
        # This signals to the log that the request has been completed fully
        logging.info("Result created")
        logging.info(type(result))
        # Here we send back the result
        price_info = ({
            "input_price": f"{(tokens/1000)*0.000150}$",
            "output_price": f"{(tokensoutput/1000)*0.000600}$",
            "total_price": f"{((tokens/1000)*0.000150)+((tokensoutput/1000)*0.000600)}$",
        })
        result["price_info"] = price_info
        return result
