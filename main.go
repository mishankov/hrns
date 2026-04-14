package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path"

	"github.com/mishankov/hrns/terminal"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

func main() {
	ctx := context.Background()

	key, _ := os.LookupEnv("HRNS_KEY")
	baseUrl, _ := os.LookupEnv("HRNS_BASE_URL")
	client := openai.NewClient(
		option.WithAPIKey(key),
		option.WithBaseURL(baseUrl),
	)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a coding assistant that talks like a pirate."),
	}

	toolsCalled := false

	for {
		if !toolsCalled {
			reader := bufio.NewReader(os.Stdin)
			terminal.PrintUserInputPrompt()
			messageText, _ := reader.ReadString('\n')
			messages = append(messages, openai.UserMessage(messageText))
		}

		toolsCalled = false

		stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: messages,
			Tools: []openai.ChatCompletionToolUnionParam{{
				OfFunction: &openai.ChatCompletionFunctionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name: "read_file",
						Description: param.Opt[string]{
							Value: "Reads file from filesystem",
						},
						Parameters: shared.FunctionParameters{
							"type": "object",
							"properties": map[string]any{
								"fileName": map[string]string{
									"type": "string",
								},
							},
							"required": []string{"fileName"},
						},
					},
				},
			}, {
				OfFunction: &openai.ChatCompletionFunctionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name: "list_files",
						Description: param.Opt[string]{
							Value: "Lists files in directory using glob pattern",
						},
						Parameters: shared.FunctionParameters{
							"type": "object",
							"properties": map[string]any{
								"dir": map[string]string{
									"type": "string",
								},
								"globPattern": map[string]string{
									"type": "string",
								},
							},
							"required": []string{"dir", "globPattern"},
						},
					},
				},
			}},
			Model: "z-ai/glm-5.1",
		})

		accumulator := openai.ChatCompletionAccumulator{}

		reasoningNow := false
		for stream.Next() {
			chunk := stream.Current()
			accumulator.AddChunk(chunk)
			if chunk.Choices[0].Delta.Content != "" {
				if reasoningNow {
					terminal.PrintNewLine()
					terminal.PrintNewLine()
				}
				reasoningNow = false

				terminal.PrintResponse(chunk.Choices[0].Delta.Content)
			} else {
				var jsonChunk map[string]string

				json.Unmarshal([]byte(chunk.Choices[0].Delta.RawJSON()), &jsonChunk)

				if jsonChunk["reasoning"] != "" {
					if !reasoningNow {
						terminal.PrintNewLine()
					}
					reasoningNow = true

					terminal.PrintReasoning(jsonChunk["reasoning"])
				}
			}
		}

		if stream.Err() != nil {
			panic(stream.Err())
		}

		messages = append(messages, accumulator.Choices[0].Message.ToParam())

		for _, choice := range accumulator.Choices {
			for _, toolCall := range choice.Message.ToolCalls {
				terminal.PrintToolCall(toolCall.Function.Name, toolCall.Function.Arguments)

				switch toolCall.Function.Name {
				case "read_file":
					var args map[string]string
					if err := json.Unmarshal(
						[]byte(toolCall.Function.Arguments),
						&args,
					); err != nil {
						log.Fatalf("invalid tool args: %v", err)
					}

					result := ""
					dat, err := os.ReadFile(args["fileName"])
					if err != nil {
						result = "ERROR: tools calling error: " + err.Error()
					} else {
						result = string(dat)
					}

					messages = append(
						messages,
						openai.ToolMessage(result, toolCall.ID),
					)
				case "list_files":
					var args map[string]string
					if err := json.Unmarshal(
						[]byte(toolCall.Function.Arguments),
						&args,
					); err != nil {
						log.Fatalf("invalid tool args: %v", err)
					}

					result := ""
					root := os.DirFS(args["dir"])

					mdFiles, err := fs.Glob(root, args["globPattern"])

					if err != nil {
						log.Fatal(err)
					}

					var files []string
					for _, v := range mdFiles {
						files = append(files, path.Join(args["dir"], v))
					}

					data, err := json.Marshal(files)
					if err != nil {
						result = "ERROR: tools calling error: " + err.Error()
					} else {
						result = string(data)
					}

					messages = append(
						messages,
						openai.ToolMessage(result, toolCall.ID),
					)
				default:
					messages = append(
						messages,
						openai.ToolMessage("ERROR: unknown tool", toolCall.ID),
					)
				}

				toolsCalled = true
			}
		}
	}
}
