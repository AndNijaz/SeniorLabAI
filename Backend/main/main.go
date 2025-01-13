package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"log/slog"

	"code.com/chatgpt"
	"github.com/joho/godotenv"
)

type Input struct {
	Text string `json:"text"`
}

var logger *slog.Logger
var requestdata *slog.Logger

func ChatGPTHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the IP address from X-Forwarded-For header or fall back to RemoteAddr
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	} else {
		// Extract the first IP if there are multiple IPs
		clientIP = strings.Split(clientIP, ",")[0]
	}

	switch r.Method {
	case "POST":
		logger.Info("Received request", "method", r.Method, "ip", clientIP)

		// Load environment variables
		if err := godotenv.Load("./.env"); err != nil {
			logger.Error("Error loading .env file", "error", err, "ip", clientIP)
			http.Error(w, "Env loading error", http.StatusInternalServerError)
			return
		}

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			logger.Error("OPENAI_API_KEY is not set in .env file", "ip", clientIP)
			http.Error(w, "API key error", http.StatusInternalServerError)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Unable to read request body", "error", err, "ip", clientIP)
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the JSON input
		var input Input
		if err := json.Unmarshal(body, &input); err != nil {
			logger.Error("Invalid JSON format", "error", err, "ip", clientIP)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Access the `text` field from the input
		text := input.Text
		requestdata.Info("Received text", "text", text, "ip", clientIP)

		// Process the `text` variable as needed
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"received": text}
		jsonResp, err := json.Marshal(response)
		if err != nil {
			logger.Error("Error forming JSON response", "error", err, "ip", clientIP)
			http.Error(w, "JSON error response", http.StatusInternalServerError)
			return
		}

		resultingText := chatgpt.ChatGPTAnalyse(string(jsonResp), apiKey)
		requestdata.Info("Resulting text", "text", resultingText, "ip", clientIP)
		_, err = w.Write([]byte(resultingText))
		if err != nil {
			logger.Error("Error writing response", "error", err, "ip", clientIP)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	default:
		logger.Error("Invalid request method", "method", r.Method, "ip", clientIP)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// BasicAuth middleware provides simple HTTP basic authentication
func BasicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Decode the "Basic " prefix
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := string(payload)
		expectedPair := username + ":" + password
		if pair != expectedPair {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func SendLogs(w http.ResponseWriter, r *http.Request) {
	// Read the contents of the log file
	data, err := os.ReadFile("./logfile.log")
	if err != nil {
		http.Error(w, "Error reading log file", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func SendUsage(w http.ResponseWriter, r *http.Request) {
	// Read the contents of the log file
	data, err := os.ReadFile("./usage.log")
	if err != nil {
		http.Error(w, "Error reading log file", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func main() {
	// Open the log file
	file, err := os.OpenFile("./logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// If logging setup fails, use the default logger to report the error and exit
		slog.Default().Error("Error opening log file", "error", err)
		os.Exit(1)
	}
	defer file.Close()
	file2, err := os.OpenFile("./usage.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// If logging setup fails, use the default logger to report the error and exit
		slog.Default().Error("Error opening log file", "error", err)
		os.Exit(1)
	}
	defer file2.Close()

	// Create a JSON handler for structured logging
	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set the desired log level
	})
	handler2 := slog.NewJSONHandler(file2, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
	requestdata = slog.New(handler2)
	slog.SetDefault(logger) // Set as the default logger
	http.HandleFunc("/", ChatGPTHandler)

	// Wrap the log endpoints with BasicAuth middleware
	logHandler := BasicAuth(http.HandlerFunc(SendLogs), "SeniorLAB", "jenajbolji")
	usageHandler := BasicAuth(http.HandlerFunc(SendUsage), "SeniorLAB", "jenajbolji")

	http.Handle("/logfile", logHandler)
	http.Handle("/usage", usageHandler)

	logger.Info("Starting server on :8468")
	if err := http.ListenAndServe(":8468", nil); err != nil {
		logger.Error("ListenAndServe failed", "error", err)
		os.Exit(1)
	}
}
