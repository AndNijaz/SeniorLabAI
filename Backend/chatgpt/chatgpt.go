package chatgpt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"code.com/webpagescraper" // Custom package for Google search function
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type chatResponseContent struct {
	Longresponse  string `json:"longresponse"`
	Shortresponse string `json:"shortresponse"`
	Title         string `json:"title"`
}

type chatResponse struct {
	Content        chatResponseContent `json:"content"`
	InternetSearch bool                `json:"internet_search"`
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

func ChatGPTAnalyse(prompt, apikey string) string {
	client := openai.NewClient(option.WithAPIKey(apikey))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	systemMessage := openai.SystemMessage(
		"Ti si inteligenti pomagac koji samo odgovara na srpskom/bosanskom jeziku. " +
			"Pazi da koristiš nazive meseci na srpskom, na primer, koristi 'juni' umesto 'lipanj'. " +
			fmt.Sprintf("Trenutni datum je %s. Ako je potrebno da se ovo tačno odgovori, možeš pozvati funkciju ", time.Now().Format("02.01.2006.")) +
			"search_google da bi našao više informacija. U odgovoru pod nazivom 'longresponse', koristi HTML " +
			"za formatiranje. Formatiraj tekst koristeći tagove kao što su <br> za nove linije, <b> za " +
			"podebljani tekst, <em> za italik, <br> za novu liniju itd. Nemoj koristiti \\n za novu liniju! Dodaj izvore kao naslov koji se može kliknuti koristeći " +
			"<a> tag sa atributom href. Koristi to da referenciras izvore koje si koristio! Stavi <a> tag i u njemu stavi href na stranicu iz koje si uzeo tu informaciju! Moras praviti reference pomocu podataka sto si dobio od search_google funkcije! Kada pises domain, pazi da je pravilno napisano i da nema dodatni slash. Takodjer napravi da link otvori u novom tabu. Napravi da su tagovi plave boje da se vidi šta je link a šta nije! " +
			"Nikada ne koristi HTML u odgovoru pod nazivom 'shortresponse'. Samo koristi čisti tekst bez " +
			"dodatnog formatiranja. Za shortcontent imaš limit od 50 riječi, a u longcontent možeš napisati " +
			"najviše 200 riječi. Ako korisnik pita za pomoćne službe, kao što su policija, hitna, vatrogasci, " +
			"ili broj za pomoć za nasilje, daj im ove brojeve: Broj za nasilje na području FBIH: 1265, " +
			"Operativni centri Civilne zaštite: 121, Policija: 122, Vatrogasci: 123, Hitna medicinska pomoć: 124, " +
			"Pomoć na cesti: 1282/1285/1288.",
	)
	userMessage := openai.UserMessage("User prompt: " + prompt)

	chatgptResponseSchema := GenerateSchema[chatResponseContent]()
	schemaparam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("Response"),
		Description: openai.F("Answers of the prompt with given information"),
		Schema:      openai.F(chatgptResponseSchema),
		Strict:      openai.Bool(true),
	}

	params := openai.ChatCompletionNewParams{
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaparam),
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

	// Initial chat completion request
	result, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "An error occurred: " + err.Error()
	}

	// Process tool calls
	for _, toolCall := range result.Choices[0].Message.ToolCalls {
		if toolCall.Function.Name == "search_google" {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return "Error parsing tool call arguments: " + err.Error()
			}
			searchQuery := args["query"].(string)
			searchResults := webpagescraper.GoogleSearch(searchQuery, 10)
			searchUsed = true

			// Append the search result message correctly
			toolMessage := openai.ToolMessage(toolCall.ID, searchResults)
			params.Messages.Value = append(params.Messages.Value, result.Choices[0].Message, toolMessage)
		}
	}

	// Second request to get the final content
	result, err = client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "An error occurred during reprocessing: " + err.Error()
	}

	var crContent chatResponseContent
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &crContent); err != nil {
		return "An error occurred during JSON Unmarshalling: " + err.Error() + "Message content: " + string([]byte(result.Choices[0].Message.Content))
	}

	cr := chatResponse{
		Content:        crContent,
		InternetSearch: searchUsed,
	}

	finalJSON, err := json.Marshal(cr)
	if err != nil {
		return "An error occurred during JSON Marshalling: " + err.Error()
	}

	return string(finalJSON)
}
