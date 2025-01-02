package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

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
	switch r.Method {
	case "POST":
		logger.Info("Received request", "method", r.Method, "ip", r.RemoteAddr)

		// Load environment variables
		if err := godotenv.Load("./.env"); err != nil {
			logger.Error("Error loading .env file", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			logger.Error("OPENAI_API_KEY is not set in .env file", "ip", r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Unable to read request body", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the JSON input
		var input Input
		if err := json.Unmarshal(body, &input); err != nil {
			logger.Error("Invalid JSON format", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Access the `text` field from the input
		text := input.Text
		requestdata.Info("Received text", "text", text, "ip", r.RemoteAddr)

		// Process the `text` variable as needed
		// Example: Respond with the received text
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"received": text}
		jsonResp, err := json.Marshal(response)
		if err != nil {
			logger.Error("Error forming JSON response", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resultingText := chatgpt.ChatGPTAnalyse(string(jsonResp), apiKey)
		requestdata.Info("Resulting text", "text", resultingText, "ip", r.RemoteAddr)
		_, err = w.Write([]byte(resultingText))
		if err != nil {
			logger.Error("Error writing response", "error", err, "ip", r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	default:
		logger.Error("Invalid request method", "method", r.Method, "ip", r.RemoteAddr)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
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
	http.HandleFunc("/logfile", SendLogs)
	http.HandleFunc("/usage", SendUsage)
	logger.Info("Starting server on :8468")

	// Start the HTTP server
	logger.Info("Server started on :8468")
	if err := http.ListenAndServe(":8468", nil); err != nil {
		logger.Error("ListenAndServe failed", "error", err)
		os.Exit(1)
	}

	// This line won't be reached because ListenAndServe is blocking unless it fails

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
