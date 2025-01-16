package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
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

	logger.Info("Incoming request",
		"method", r.Method,
		"ip", clientIP,
		"user_agent", r.UserAgent(),
		"path", r.URL.Path,
		"content_length", r.ContentLength)

	switch r.Method {
	case "POST":

		// Load environment variables
		if err := godotenv.Load("./.env"); err != nil {
			logger.Error("Environment configuration error",
				"error", err,
				"ip", clientIP,
				"file", ".env")
			http.Error(w, "Configuration error", http.StatusInternalServerError)
			return
		}

		logger.Info("Environment loaded successfully", "ip", clientIP)

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			logger.Error("OPENAI_API_KEY is not set in .env file", "ip", clientIP)
			http.Error(w, "API key error", http.StatusInternalServerError)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Request body read error",
				"error", err,
				"ip", clientIP,
				"content_length", r.ContentLength)
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		logger.Info("Request body read successfully",
			"ip", clientIP,
			"body_size", len(body))

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
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		logger.Info("Authentication attempt",
			"ip", clientIP,
			"path", r.URL.Path,
			"user_agent", r.UserAgent())

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		auth := r.Header.Get("Authorization")
		if auth == "" {
			logger.Warn("Missing authorization header",
				"ip", clientIP,
				"path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Decode the "Basic " prefix
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			logger.Error("Invalid authorization header format",
				"error", err,
				"ip", clientIP,
				"path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := string(payload)
		expectedPair := username + ":" + password
		if pair != expectedPair {
			logger.Warn("Invalid credentials",
				"ip", clientIP,
				"path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Info("Authentication successful",
			"ip", clientIP,
			"path", r.URL.Path,
			"username", username)

		next.ServeHTTP(w, r)
	})
}

// LogEntry represents a single log entry in JSON format
type LogEntry map[string]interface{}

func parseLogFile(filepath string) ([]LogEntry, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var entries []LogEntry
	lines := strings.Split(string(file), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}
		entries = append(entries, entry)
	}

	// Sort entries by time in descending order (newest first)
	sort.Slice(entries, func(i, j int) bool {
		timeI, okI := entries[i]["time"].(string)
		timeJ, okJ := entries[j]["time"].(string)
		if !okI || !okJ {
			return false
		}
		return timeI > timeJ
	})

	return entries, nil
}

func serveLogViewer(w http.ResponseWriter, r *http.Request, logFile string) {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	logger.Info("Log viewer request",
		"ip", clientIP,
		"path", r.URL.Path)

	// Serve the HTML template
	http.ServeFile(w, r, "./templates/logs.html")
}

func serveLogData(w http.ResponseWriter, r *http.Request, logFile string) {
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	logger.Info("Log data request",
		"ip", clientIP,
		"path", r.URL.Path)

	entries, err := parseLogFile(logFile)
	if err != nil {
		logger.Error("Failed to parse log file",
			"error", err,
			"ip", clientIP,
			"path", logFile)
		http.Error(w, "Error reading log file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func SendLogs(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/data") {
		serveLogData(w, r, "/app/logs/logfile.log")
	} else {
		serveLogViewer(w, r, "/app/logs/logfile.log")
	}
}

func SendUsage(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/data") {
		serveLogData(w, r, "/app/logs/usage.log")
	} else {
		serveLogViewer(w, r, "/app/logs/usage.log")
	}
}

func main() {
	slog.Info("Initializing server...")

	// Open the log file
	file, err := os.OpenFile("/app/logs/logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// If logging setup fails, use the default logger to report the error and exit
		slog.Default().Error("Error opening log file",
			"error", err,
			"path", "/app/logs/logfile.log")
		os.Exit(1)
	}
	defer file.Close()

	file2, err := os.OpenFile("/app/logs/usage.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// If logging setup fails, use the default logger to report the error and exit
		slog.Default().Error("Error opening usage log file",
			"error", err,
			"path", "/app/logs/usage.log")
		os.Exit(1)
	}
	defer file2.Close()

	slog.Info("Log files opened successfully",
		"logfile", "/app/logs/logfile.log",
		"usagelog", "/app/logs/usage.log")

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

	// Handle both the viewer and data endpoints
	http.Handle("/logfile", logHandler)
	http.Handle("/logfile/data", logHandler)
	http.Handle("/usage", usageHandler)
	http.Handle("/usage/data", usageHandler)

	logger.Info("Starting server",
		"port", 8468,
		"handlers", []string{"/", "/logfile", "/usage"})

	if err := http.ListenAndServe(":8468", nil); err != nil {
		logger.Error("Server startup failed",
			"error", err,
			"port", 8468)
		os.Exit(1)
	}
}
