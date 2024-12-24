package webpagescraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/markusmobius/go-htmldate"
	"github.com/pkoukk/tiktoken-go"
	"golang.org/x/net/html"
)

func scrapeWebpage(url string) (string, error) {
	// Ensure the URL has a valid scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Fetch the URL
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching URL %s: %v", url, err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the entire response body
	var result string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body from URL %s: %v", url, err)
		result = "There was an error reading this page, disregard it!"
	}
	if result != "" {
		finalResult := "<HTML CONTENT>\n" + result + "\n</HTML CONTENT>" +
			"\n<PUBLISHED DATE>\n" + "No published date found" + "\n</PUBLISHED DATE>\n" + "<URL>\n" + url + "\n</URL>"
		return finalResult, nil
	}
	tokenreader := bytes.NewReader(bodyBytes)
	tokenizer := html.NewTokenizer(tokenreader)
	textTags := []string{
		"a", "p", "span", "em", "string", "blockquote", "q", "cite", "h1", "h2", "h3", "h4", "h5", "h6",
	}
	tag := ""
	enter := false
	attrs := map[string]string{}

	for {
		tt := tokenizer.Next()
		token := tokenizer.Token()
		err := tokenizer.Err()
		if err == io.EOF {
			break
		}

		switch tt {
		case html.ErrorToken:
			log.Fatal(err)
		case html.StartTagToken, html.SelfClosingTagToken:
			enter = false
			attrs = map[string]string{}

			tag = token.Data
			for _, ttt := range textTags {
				if tag == ttt {
					enter = true
					for _, attr := range token.Attr {
						attrs[attr.Key] = attr.Val
					}
					break
				}
			}
		case html.TextToken:
			if enter {
				data := strings.TrimSpace(token.Data)

				if len(data) > 0 {
					switch tag {
					case "a":
						result += "[" + data + "](" + attrs["href"] + ")\n"
					case "h1", "h2", "h3":
						result += "## " + data
					case "h4", "h5", "h6":
						result += "### " + data
					default:
						result += data
					}
				}
			}
		}
	}
	// Extract the date using a fresh reader over the bytes
	opts := htmldate.Options{
		ExtractTime:     true,
		UseOriginalDate: false,
		EnableLog:       false,
	}
	date, err := htmldate.FromReader(bytes.NewReader(bodyBytes), opts)
	if err != nil {
		log.Printf("Failed to extract date: %v", err)
	}

	// Convert the byte slice into a string for output

	// Construct the final result
	finalResult := "<HTML CONTENT>\n" + result + "\n</HTML CONTENT>" +
		"\n<PUBLISHED DATE>\n" + date.Format("2006-01-02") + "\n</PUBLISHED DATE>\n" + "<URL>\n" + url + "\n</URL>"

	return finalResult, nil
}

func WebpageAnalyse(url string) string {
	content, err := scrapeWebpage(url)
	if err != nil {
		log.Printf("Failed to scrape webpage: %v", err)
	}
	return content
}

func TokenCounter(text string) int {
	tke, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		fmt.Println(err)
	}
	tokens := tke.Encode(text, nil, nil)
	return len(tokens)
}

func GoogleSearch(query string, count int) string {
	encodedQuery := url.QueryEscape(query)
	url := "http://searxng:8080/search?q=" + encodedQuery + "&format=json&safesearch=1"
	log.Printf("%s", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(bodyText, &data)
	if err != nil {
		log.Fatal(err)
	}
	URLlist := ""
	re := regexp.MustCompile(`^https:\/\/(?:old\.)?reddit\.com.*$`)
	if results, ok := data["results"].([]interface{}); ok {
		for i, result := range results {
			if i >= count {
				break
			}
			if re.MatchString(url) {
				i -= 1
				continue
			}
			// Each result is a map, so convert it
			if resultMap, ok := result.(map[string]interface{}); ok {
				// Fetch the "parsed_url" field
				if parsedURLs, ok := resultMap["parsed_url"].([]interface{}); ok {
					fullURL := ""
					for i, part := range parsedURLs {
						if i == 0 {
							fullURL += part.(string) + "://"
						} else if i == 1 {
							fullURL += part.(string) + "/"
						} else {
							fullURL += part.(string)
						}
					}
					URLlist += fullURL + "\n"
				}
			}
		}
	} else {
		log.Println("No 'results' key found or it's not an array")
	}
	urlMap := urlsToMap(URLlist)
	var (
		prompt string
		mu     sync.Mutex
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	for _, url := range urlMap {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}
			analysis := WebpageAnalyse(url)
			mu.Lock()
			prompt += analysis
			currentTokenCount := TokenCounter(prompt)
			mu.Unlock()
			if currentTokenCount > 70000 {
				cancel()
			}
		}(url)
	}
	wg.Wait()
	return prompt
}

func urlsToMap(input string) map[int]string {
	result := make(map[int]string)
	lines := strings.Split(input, "\n") // Split by newlines
	for i, line := range lines {
		line = strings.TrimSpace(line) // Remove leading/trailing spaces
		if line != "" {                // Skip empty lines
			result[i+1] = line // Use 1-based indexing for keys
		}
	}
	return result
}
