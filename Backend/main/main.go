package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"code.com/chatgpt"
	"github.com/joho/godotenv"
)

type Input struct {
	Text string `json:"text"`
}

func ChatGPTHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("Recieved request")
		err := godotenv.Load("./.env")
		if err != nil {
			fmt.Println("Error loading .env file")
			return
		}
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("OPENAI_API_KEY is not set in .env file")
			return
		}
		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the JSON input
		var input Input
		if err := json.Unmarshal(body, &input); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Access the `text` field from the input
		text := input.Text

		// Process the `text` variable as needed
		// Example: Respond with the received text
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"received": text}
		jsonResp, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Error forming JSON response", http.StatusInternalServerError)
			return
		}
		resultingtext := chatgpt.ChatGPTAnalyse(string(jsonResp), apiKey)
		io.WriteString(w, resultingtext)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	http.HandleFunc("/", ChatGPTHandler)
	log.Println("Starting server")
	err := http.ListenAndServe(":8468", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		log.Println("Server started on :8468")
	}
}
