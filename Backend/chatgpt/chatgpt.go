package chatgpt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"log/slog"

	"code.com/webpagescraper"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared" // Import the shared package
)

// initializeLogger sets up the logger for the application.
func initializeLogger() (*slog.Logger, error) {
	file, err := os.OpenFile("./logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Default().Error("Error opening log file", "error", err)
		return nil, err
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger, nil
}

// chatResponseContent defines the structure of the response content.
type chatResponseContent struct {
	Longresponse  string `json:"longresponse"`
	Shortresponse string `json:"shortresponse"`
	Title         string `json:"title"`
}

// chatResponse wraps the content and indicates if an internet search was used.
type chatResponse struct {
	Content        chatResponseContent `json:"content"`
	InternetSearch bool                `json:"internet_search"`
}

// GenerateSchema generates a JSON schema for the given type.
func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

// MakeChatCompletionCall calls the OpenAI Chat Completion API.
func MakeChatCompletionCall(client *openai.Client, ctx context.Context, params *openai.ChatCompletionNewParams, logger *slog.Logger) (*openai.ChatCompletion, error) {
	result, err := client.Chat.Completions.New(ctx, *params)
	if err != nil {
		logger.Error("Error during ChatGPT request", "error", err)
		return result, err
	}
	logger.Info("ChatGPT request successful")
	return result, nil
}

// ProcessToolCalls handles any tool calls returned by the assistant.
func ProcessToolCalls(result *openai.ChatCompletion, logger *slog.Logger) ([]openai.ChatCompletionMessageParamUnion, bool, error) {
	searchUsed := false
	toolMessages := []openai.ChatCompletionMessageParamUnion{}

	for _, choice := range result.Choices {
		if choice.Message.ToolCalls != nil {
			for _, toolCall := range choice.Message.ToolCalls {
				if toolCall.Function.Name == "search_google" {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logger.Error("Error parsing tool call arguments", "error", err)
						return nil, false, fmt.Errorf("error parsing tool call arguments: %v", err)
					}

					queryValue, ok := args["query"]
					if !ok {
						errMsg := "Error: 'query' field not found in tool call arguments"
						logger.Error(errMsg)
						return nil, false, fmt.Errorf("%s", errMsg)
					}

					searchQuery, ok := queryValue.(string)
					if !ok {
						errMsg := "Error: 'query' field is not of type string"
						logger.Error(errMsg)
						return nil, false, fmt.Errorf("%s", errMsg)
					}

					// Perform the search using the webpage scraper
					searchResults := webpagescraper.GoogleSearch(searchQuery, 10)
					searchUsed = true

					// Create a tool message response to pass back to the assistant
					toolMessage := openai.ToolMessage(toolCall.ID, searchResults)
					toolMessages = append(toolMessages, toolMessage)
				}
			}
		}
	}

	return toolMessages, searchUsed, nil
}

// ChatGPTAnalyse processes the prompt using OpenAI's API and returns the response.
func ChatGPTAnalyse(prompt, apikey string) string {
	logger, err := initializeLogger()
	if err != nil {
		return "Failed to initialize logger"
	}

	logger.Info("Start processing ChatGPT analysis", "prompt", prompt)

	client := openai.NewClient(option.WithAPIKey(apikey))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	systemMessage := openai.SystemMessage(
		"You are an intelligent assistant that responds exclusively in Serbian/Bosnian. " +
			fmt.Sprintf("Use Serbian month names (e.g., 'juni' instead of 'lipanj'). The current date is %s.", time.Now().Format("02.01.2006.")) +
			"If exact data is needed, use the search_google function to retrieve additional information. " +
			"\n\n" +
			"1) In the response named 'longresponse', always use HTML for formatting. " +
			"   - Use <br> instead of \\n for new lines. " +
			"   - Use <b> for bold text and <em> for italics. " +
			"   - Use HTML tags instead of markdown, and under no circumstances can you use markdown." +
			"   - Add clickable sources using <a> tags with href attributes pointing to references found " +
			"     via the search_google function. " +
			"   - Ensure the domain is correct and does not include extra slashes. " +
			"   - Make links open in a new tab and display in a blue color (VERY IMPORTANT). " +
			"\n" +
			"2) In the 'shortresponse', never use HTML. " +
			"   - Use only plain text. " +
			"   - Limit is 50 words. " +
			"\n" +
			"3) The 'longresponse' is limited to 200 words. " +
			"\n" +
			"4) If the user requests emergency service numbers (police, ambulance, fire brigade, or " +
			"   domestic violence hotlines), always provide: " +
			"   - Domestic violence helpline in FBIH: 1265 " +
			"   - Civil Protection Operational Centers: 121 " +
			"   - Police: 122 " +
			"   - Fire Department: 123 " +
			"   - Emergency Medical Services: 124 " +
			"   - Roadside Assistance: 1282/1285/1288.",
	)

	userMessage := openai.UserMessage("User prompt: " + prompt)

	// Generate the JSON schema for the expected response format
	chatgptResponseSchema := GenerateSchema[chatResponseContent]()
	schemaParam := shared.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("Response"),
		Description: openai.F("Answers of the prompt with given information"),
		Schema:      openai.F(chatgptResponseSchema),
		Strict:      openai.Bool(true),
	}

	params := openai.ChatCompletionNewParams{
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			shared.ResponseFormatJSONSchemaParam{
				Type:       openai.F(shared.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{systemMessage, userMessage}),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String("search_google"),
					Description: openai.String("Search Google for additional information"),
					Parameters: openai.F(openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]string{
								"type": "string",
							},
						},
						"required": []string{"query"},
					}),
				}),
			},
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	}

	searchUsed := false
	var crContent chatResponseContent
	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// First API call
		result, err := MakeChatCompletionCall(client, ctx, &params, logger)
		if err != nil {
			return fmt.Sprintf("An error occurred: %v", err.Error())
		}

		// Process any tool calls (e.g., search_google)
		toolMessages, toolSearchUsed, err := ProcessToolCalls(result, logger)
		searchUsed = searchUsed || toolSearchUsed
		if err != nil {
			return err.Error()
		}

		// Append the assistant's response and any tool messages to the conversation
		params.Messages.Value = append(params.Messages.Value, result.Choices[0].Message)
		if len(toolMessages) > 0 {
			params.Messages.Value = append(params.Messages.Value, toolMessages...)
		}

		// Second API call with updated messages
		result, err = MakeChatCompletionCall(client, ctx, &params, logger)
		if err != nil {
			return fmt.Sprintf("An error occurred during reprocessing: %v", err.Error())
		}
		logger.Info("Received response from OpenAI", "response", result.Choices[0].Message.Content)
		if result.Choices[0].Message.Content == "" {
			logger.Warn("Received blank response, retrying...", "attempt", attempt)
			time.Sleep(1 * time.Second)
			continue
		}
		if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &crContent); err != nil {
			logger.Error("Error during JSON Unmarshalling", "error", err, "messageContent", result.Choices[0].Message.Content)
			return fmt.Sprintf("An error occurred during JSON Unmarshalling: %v Message content: %s", err.Error(), result.Choices[0].Message.Content)
		}

		if crContent.Longresponse != "" || crContent.Shortresponse != "" {
			break
		}

		logger.Warn("Received incomplete response, retrying...", "attempt", attempt)
		time.Sleep(1 * time.Second)
	}
	cr := chatResponse{
		Content:        crContent,
		InternetSearch: searchUsed,
	}

	finalJSON, err := json.Marshal(cr)
	if err != nil {
		logger.Error("Error during JSON Marshalling", "error", err)
		return fmt.Sprintf("An error occurred during JSON Marshalling: %v", err.Error())
	}

	logger.Info("ChatGPT analysis completed successfully", "prompt", prompt, "result", string(finalJSON))

	return string(finalJSON)
}
