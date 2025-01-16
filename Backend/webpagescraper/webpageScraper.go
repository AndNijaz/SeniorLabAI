package webpagescraper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/markusmobius/go-htmldate"
	"github.com/pkoukk/tiktoken-go"
	"golang.org/x/net/html"
)

func initializeLogger() (*slog.Logger, error) {
	file, err := os.OpenFile("./logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// If logging setup fails, use the default logger to report the error and exit
		slog.Default().Error("Error opening log file", "error", err)
		return nil, err
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger, nil
}

func scrapeWebpage(url string) (string, error) {
	logger, err := initializeLogger()
	if err != nil {
		return "", errors.New("failed to initialize logger")
	}

	logger.Info("Processing URL for scraping",
		"original_url", url)

	// Ensure the URL has a valid scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		originalURL := url
		url = "https://" + url
		logger.Info("Added HTTPS scheme to URL",
			"original_url", originalURL,
			"modified_url", url)
	}

	// Fetch the URL
	logger.Info("Fetching URL", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Error fetching URL",
			"url", url,
			"error", err,
			"status_code", resp.StatusCode)
		return "", err
	}
	defer resp.Body.Close()

	// Read the entire response body
	var result string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body from URL %s: %v", url, err)
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
		"p", "span", "em", "string", "blockquote", "q", "cite", "h1", "h2", "h3", "h4", "h5", "h6",
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
		logger.Error("Failed to extract date", "error", err.Error())
	}

	// Convert the byte slice into a string for output

	// Construct the final result
	finalResult := "<HTML CONTENT>\n" + result + "\n</HTML CONTENT>" +
		"\n<PUBLISHED DATE>\n" + date.Format("2006-01-02") + "\n</PUBLISHED DATE>\n" + "<URL>\n" + url + "\n</URL>"

	return finalResult, nil
}

func WebpageAnalyse(url string) string {
	logger, err := initializeLogger()
	if err != nil {
		return "Failed to initialize logger"
	}

	content, err := scrapeWebpage(url)
	if err != nil {
		logger.Error("Failed to scrape webpage", "error", err, "Url", url)
	}
	return content
}

func TokenCounter(text string) int {
	logger, err := initializeLogger()
	if err != nil {
		return -1
	}

	logger.Info("Starting token count",
		"text_length", len(text))

	tke, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		logger.Error("Failed to create token encoder",
			"error", err,
			"model", "gpt-4o-mini")
		return -1
	}

	tokens := tke.Encode(text, nil, nil)
	tokenCount := len(tokens)

	logger.Info("Token count completed",
		"text_length", len(text),
		"token_count", tokenCount)

	return tokenCount
}

func GoogleSearch(query string, count int) string {
	logger, err := initializeLogger()
	if err != nil {
		return "Failed to initialize logger"
	}

	logger.Info("Starting Google search",
		"query", query,
		"requested_results", count)

	encodedQuery := url.QueryEscape(query)
	searchURL := "http://searxng:8080/search?q=" + encodedQuery + "&format=json&safesearch=1"
	logger.Info("Search request prepared",
		"encoded_query", encodedQuery,
		"search_url", searchURL)
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request",
			"error", err,
			"url", searchURL)
		return "Failed to create search request"
	}

	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request",
			"error", err,
			"url", searchURL)
		return "Failed to execute search request"
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body",
			"error", err,
			"url", searchURL)
		return "Failed to read search results"
	}

	var data map[string]interface{}
	err = json.Unmarshal(bodyText, &data)
	if err != nil {
		logger.Error("Failed to parse JSON response",
			"error", err,
			"response_body", string(bodyText))
		return "Failed to parse search results"
	}

	logger.Info("Successfully retrieved search results",
		"status_code", resp.StatusCode,
		"body_size", len(bodyText))
	URLlist := ""
	re := regexp.MustCompile(`^https:\/\/(?:old\.)?reddit\.com.*$`)
	if results, ok := data["results"].([]interface{}); ok {
		for i, result := range results {
			if i >= count {
				break
			}
			if re.MatchString(searchURL) {
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
		logger.Error("No 'results' key found or it's not an array")
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
				logger.Info("Context cancelled, stopping webpage analysis",
					"url", url)
				return
			default:
			}

			logger.Info("Starting webpage analysis in goroutine",
				"url", url)
			analysis := WebpageAnalyse(url)

			mu.Lock()
			originalLength := len(prompt)
			prompt += analysis
			currentTokenCount := TokenCounter(prompt)
			logger.Info("Updated prompt in goroutine",
				"url", url,
				"original_length", originalLength,
				"new_length", len(prompt),
				"token_count", currentTokenCount)

			if currentTokenCount > 70000 {
				logger.Warn("Token limit exceeded, cancelling remaining operations",
					"token_count", currentTokenCount,
					"limit", 70000)
				cancel()
			}
			mu.Unlock()
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
