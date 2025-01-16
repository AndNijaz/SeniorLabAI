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
	"github.com/openai/openai-go/shared"
)

// initializeLogger sets up the logger for the application.
func initializeLogger() (*slog.Logger, error) {
	file, err := os.OpenFile("/app/logs/logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	logger, err := initializeLogger()
	if err != nil {
		return nil
	}

	startTime := time.Now()
	logger.Info("Starting schema generation",
		"type", fmt.Sprintf("%T", *new(T)))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	logger.Info("Schema generation completed",
		"type", fmt.Sprintf("%T", *new(T)),
		"duration_ms", time.Since(startTime).Milliseconds(),
		"schema_size", len(fmt.Sprintf("%v", schema)))

	return schema
}

// MakeChatCompletionCall calls the OpenAI Chat Completion API.
func MakeChatCompletionCall(client *openai.Client, ctx context.Context, params *openai.ChatCompletionNewParams, logger *slog.Logger) (*openai.ChatCompletion, error) {
	startTime := time.Now()
	logger.Info("Starting ChatGPT request",
		"model", params.Model.Value,
		"message_count", len(params.Messages.Value),
		"has_tools", params.Tools.Value != nil,
		"has_response_format", params.ResponseFormat.Value != nil)

	result, err := client.Chat.Completions.New(ctx, *params)
	if err != nil {
		logger.Error("ChatGPT request failed",
			"error", err,
			"model", params.Model.Value,
			"message_count", len(params.Messages.Value),
			"duration_ms", time.Since(startTime).Milliseconds())
		return result, err
	}

	logger.Info("ChatGPT request successful",
		"duration_ms", time.Since(startTime).Milliseconds(),
		"completion_tokens", result.Usage.CompletionTokens,
		"prompt_tokens", result.Usage.PromptTokens,
		"total_tokens", result.Usage.TotalTokens,
		"finish_reason", result.Choices[0].FinishReason,
		"has_tool_calls", result.Choices[0].Message.ToolCalls != nil,
		"response_length", len(result.Choices[0].Message.Content))

	return result, nil
}

// ProcessToolCalls handles any tool calls returned by the assistant.
func ProcessToolCalls(result *openai.ChatCompletion, logger *slog.Logger) ([]openai.ChatCompletionMessageParamUnion, bool, error) {
	startTime := time.Now()
	searchUsed := false
	toolMessages := []openai.ChatCompletionMessageParamUnion{}

	logger.Info("Starting tool calls processing",
		"has_choices", len(result.Choices) > 0,
		"first_choice_has_tool_calls", result.Choices[0].Message.ToolCalls != nil)

	for choiceIndex, choice := range result.Choices {
		if choice.Message.ToolCalls != nil {
			for callIndex, toolCall := range choice.Message.ToolCalls {
				logger.Info("Processing tool call",
					"choice_index", choiceIndex,
					"call_index", callIndex,
					"tool_id", toolCall.ID,
					"function_name", toolCall.Function.Name)

				if toolCall.Function.Name == "search_google" {
					searchStartTime := time.Now()
					logger.Info("Processing search_google tool call",
						"tool_id", toolCall.ID,
						"raw_arguments", toolCall.Function.Arguments)

					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logger.Error("Failed to parse tool call arguments",
							"error", err,
							"raw_arguments", toolCall.Function.Arguments,
							"duration_ms", time.Since(searchStartTime).Milliseconds())
						return nil, false, fmt.Errorf("error parsing tool call arguments: %v", err)
					}

					queryValue, ok := args["query"]
					if !ok {
						errMsg := "Error: 'query' field not found in tool call arguments"
						logger.Error(errMsg,
							"available_args", args,
							"duration_ms", time.Since(searchStartTime).Milliseconds())
						return nil, false, fmt.Errorf("%s", errMsg)
					}

					searchQuery, ok := queryValue.(string)
					if !ok {
						errMsg := "Error: 'query' field is not of type string"
						logger.Error(errMsg,
							"query_type", fmt.Sprintf("%T", queryValue),
							"duration_ms", time.Since(searchStartTime).Milliseconds())
						return nil, false, fmt.Errorf("%s", errMsg)
					}

					logger.Info("Executing Google search",
						"query", searchQuery,
						"max_results", 10)

					// Perform the search using the webpage scraper
					searchResults := webpagescraper.GoogleSearch(searchQuery, 10)
					searchUsed = true

					// Create a tool message response to pass back to the assistant
					toolMessage := openai.ToolMessage(toolCall.ID, searchResults)
					toolMessages = append(toolMessages, toolMessage)

					logger.Info("Search completed",
						"tool_id", toolCall.ID,
						"results_length", len(searchResults),
						"duration_ms", time.Since(searchStartTime).Milliseconds())
				}
			}
		}
	}

	logger.Info("Tool calls processing completed",
		"duration_ms", time.Since(startTime).Milliseconds(),
		"tool_messages_count", len(toolMessages),
		"search_used", searchUsed)

	return toolMessages, searchUsed, nil
}

// ChatGPTAnalyse processes the prompt using OpenAI's API and returns the response.
func ChatGPTAnalyse(prompt, apikey string) string {
	startTime := time.Now()
	logger, err := initializeLogger()
	if err != nil {
		return "Failed to initialize logger"
	}

	logger.Info("Starting ChatGPT analysis",
		"prompt_length", len(prompt),
		"api_key_length", len(apikey))

	client := openai.NewClient(option.WithAPIKey(apikey))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Prepare system message
	currentDate := time.Now().Format("02.01.2006.")
	logger.Info("Preparing system message",
		"current_date", currentDate)

	systemMessageContent := "You are an intelligent assistant that responds exclusively in Serbian/Bosnian. " +
		fmt.Sprintf("Use Serbian month names (e.g., 'juni' instead of 'lipanj'). The current date is %s.", currentDate) +
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
		"   - Roadside Assistance: 1282/1285/1288."

	logger.Info("System message prepared",
		"message_length", len(systemMessageContent),
		"contains_html_instructions", true,
		"contains_emergency_numbers", true)

	systemMessage := openai.SystemMessage(systemMessageContent)
	userMessage := openai.UserMessage("User prompt: " + prompt)

	// Generate schema
	schemaStartTime := time.Now()
	logger.Info("Starting schema generation")

	chatgptResponseSchema := GenerateSchema[chatResponseContent]()
	schemaParam := shared.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("Response"),
		Description: openai.F("Answers of the prompt with given information"),
		Schema:      openai.F(chatgptResponseSchema),
		Strict:      openai.Bool(true),
	}

	logger.Info("Schema generation completed",
		"duration_ms", time.Since(schemaStartTime).Milliseconds(),
		"schema_size", len(fmt.Sprintf("%v", chatgptResponseSchema)))

	// Prepare API parameters
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
		attemptStartTime := time.Now()
		logger.Info("Starting API call attempt",
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"message_count", len(params.Messages.Value))

		// First API call
		result, err := MakeChatCompletionCall(client, ctx, &params, logger)
		if err != nil {
			logger.Error("First API call failed",
				"attempt", attempt,
				"duration_ms", time.Since(attemptStartTime).Milliseconds(),
				"error", err)
			return fmt.Sprintf("An error occurred: %v", err.Error())
		}

		logger.Info("First API call completed",
			"attempt", attempt,
			"duration_ms", time.Since(attemptStartTime).Milliseconds(),
			"completion_tokens", result.Usage.CompletionTokens,
			"prompt_tokens", result.Usage.PromptTokens,
			"total_tokens", result.Usage.TotalTokens)

		// Process tool calls
		toolStartTime := time.Now()
		logger.Info("Processing tool calls",
			"attempt", attempt,
			"has_tool_calls", result.Choices[0].Message.ToolCalls != nil)

		toolMessages, toolSearchUsed, err := ProcessToolCalls(result, logger)
		searchUsed = searchUsed || toolSearchUsed
		if err != nil {
			logger.Error("Tool calls processing failed",
				"attempt", attempt,
				"duration_ms", time.Since(toolStartTime).Milliseconds(),
				"error", err)
			return err.Error()
		}

		logger.Info("Tool calls processed",
			"attempt", attempt,
			"duration_ms", time.Since(toolStartTime).Milliseconds(),
			"tool_messages_count", len(toolMessages),
			"search_used", toolSearchUsed)

		// Update conversation with tool results
		params.Messages.Value = append(params.Messages.Value, result.Choices[0].Message)
		if len(toolMessages) > 0 {
			params.Messages.Value = append(params.Messages.Value, toolMessages...)
			logger.Info("Added tool messages to conversation",
				"attempt", attempt,
				"new_message_count", len(params.Messages.Value))
		}

		// Second API call
		secondCallStartTime := time.Now()
		logger.Info("Starting second API call",
			"attempt", attempt,
			"message_count", len(params.Messages.Value))

		result, err = MakeChatCompletionCall(client, ctx, &params, logger)
		if err != nil {
			logger.Error("Second API call failed",
				"attempt", attempt,
				"duration_ms", time.Since(secondCallStartTime).Milliseconds(),
				"error", err)
			return fmt.Sprintf("An error occurred during reprocessing: %v", err.Error())
		}

		responseContent := result.Choices[0].Message.Content
		logger.Info("Second API call completed",
			"attempt", attempt,
			"duration_ms", time.Since(secondCallStartTime).Milliseconds(),
			"response_length", len(responseContent),
			"completion_tokens", result.Usage.CompletionTokens,
			"prompt_tokens", result.Usage.PromptTokens,
			"total_tokens", result.Usage.TotalTokens)

		// Process response
		if responseContent == "" {
			logger.Warn("Received empty response",
				"attempt", attempt,
				"will_retry", attempt < maxAttempts)
			time.Sleep(1 * time.Second)
			continue
		}

		// Parse JSON response
		unmarshalStartTime := time.Now()
		if err := json.Unmarshal([]byte(responseContent), &crContent); err != nil {
			logger.Error("JSON unmarshalling failed",
				"attempt", attempt,
				"duration_ms", time.Since(unmarshalStartTime).Milliseconds(),
				"error", err,
				"response_content", responseContent)
			return fmt.Sprintf("An error occurred during JSON Unmarshalling: %v Message content: %s", err.Error(), responseContent)
		}

		logger.Info("Response parsed successfully",
			"attempt", attempt,
			"duration_ms", time.Since(unmarshalStartTime).Milliseconds(),
			"has_long_response", crContent.Longresponse != "",
			"has_short_response", crContent.Shortresponse != "",
			"has_title", crContent.Title != "")

		if crContent.Longresponse != "" || crContent.Shortresponse != "" {
			break
		}

		logger.Warn("Received incomplete response",
			"attempt", attempt,
			"will_retry", attempt < maxAttempts)
		time.Sleep(1 * time.Second)
	}

	// Prepare final response
	cr := chatResponse{
		Content:        crContent,
		InternetSearch: searchUsed,
	}

	finalJSON, err := json.Marshal(cr)
	if err != nil {
		logger.Error("JSON marshalling failed",
			"error", err,
			"content_size", len(fmt.Sprintf("%v", cr)))
		return fmt.Sprintf("An error occurred during JSON Marshalling: %v", err.Error())
	}

	logger.Info("ChatGPT analysis completed successfully",
		"total_duration_ms", time.Since(startTime).Milliseconds(),
		"prompt_length", len(prompt),
		"response_length", len(finalJSON),
		"internet_search_used", searchUsed)

	return string(finalJSON)
}
