package loop

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

type Loop struct {
	openAIClient *openai.Client
	systemPrompt string
	model        string
	tools        map[string]Tool
	messages     []openai.ChatCompletionMessageParamUnion
	chunkCh      chan Chunk
}

func New(openAIClient *openai.Client, systemPrompt string, tools map[string]Tool) *Loop {
	return &Loop{
		openAIClient: openAIClient,
		systemPrompt: systemPrompt,
		tools:        tools,
		messages:     []openai.ChatCompletionMessageParamUnion{},
		chunkCh:      make(chan Chunk),
	}
}

func (a *Loop) RunLoop(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) {
	// Creating messages with system propt as first messages
	messages = append([]openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(a.systemPrompt),
	}, messages...)

	// Composing tools for agent
	tools := []openai.ChatCompletionToolUnionParam{}
	for name, tool := range a.tools {
		properties := map[string]map[string]string{}
		required := []string{}
		for _, argument := range tool.Arguments() {
			properties[argument.Name] = map[string]string{
				"type": argument.Type,
			}
			required = append(required, argument.Name)
		}

		parameters := shared.FunctionParameters{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}

		tools = append(tools, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name: name,
					Description: param.Opt[string]{
						Value: tool.Description(),
					},
					Parameters: parameters,
				},
			},
		})
	}

	for {
		// Creating streaming chat completions object
		stream := a.openAIClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: messages,
			Tools:    tools,
			Model:    model,
		})

		// Creating accumulator to accumulate model response chunks into single object
		accumulator := openai.ChatCompletionAccumulator{}

		// Reading from response stream
		for stream.Next() {
			chunk := stream.Current()
			accumulator.AddChunk(chunk)

			if len(chunk.Choices) == 0 {
				continue
			}

			if chunk.Choices[0].Delta.Content != "" {
				// Process regular chunk
				a.sendChunk(NewChunkMessage(chunk.Choices[0].Delta.Content))
			} else {
				// Process reasoning chunk
				var jsonChunk map[string]string

				json.Unmarshal([]byte(chunk.Choices[0].Delta.RawJSON()), &jsonChunk)

				if jsonChunk["reasoning"] != "" {
					a.sendChunk(NewChunkReasoning(jsonChunk["reasoning"]))
				}
			}
		}

		if len(accumulator.Choices) > 0 {
			messages = append(messages, accumulator.Choices[0].Message.ToParam())
		}

		if stream.Err() != nil {
			a.sendChunk(NewChunkError("error in LLM stream: " + stream.Err().Error()))
		}

		// Tool calling
		toolsCalled := false
		for _, choice := range accumulator.Choices {
			for _, toolCall := range choice.Message.ToolCalls {
				toolsCalled = true

				tool, toolExists := a.tools[toolCall.Function.Name]
				if !toolExists {
					messages = append(
						messages,
						openai.ToolMessage("ERROR: tool not found", toolCall.ID),
					)

					a.sendChunk(NewChunkToolCallError(toolCall.Function.Name, "tool not found"))
					continue
				}

				var args map[string]any
				if err := json.Unmarshal(
					[]byte(toolCall.Function.Arguments),
					&args,
				); err != nil {
					messages = append(
						messages,
						openai.ToolMessage("ERROR: tools calling error: "+err.Error(), toolCall.ID),
					)

					a.sendChunk(NewChunkToolCallError(toolCall.Function.Name, "tools calling error: "+err.Error()))
					continue
				}

				a.sendChunk(NewChunkToolCallStart(toolCall.Function.Name, args))

				// Calling tool here
				result := tool.Call(args)

				messages = append(
					messages,
					openai.ToolMessage(result, toolCall.ID),
				)
				a.sendChunk(NewChunkToolCallResult(toolCall.Function.Name, result))
			}
		}

		if !toolsCalled {
			break
		}
	}

	a.messages = messages

	a.sendChunk(NewChunkEnd())
}

func (a *Loop) Chunks() chan Chunk {
	return a.chunkCh
}

func (a *Loop) Messages() []openai.ChatCompletionMessageParamUnion {
	return a.messages
}

func (a *Loop) sendChunk(chunk Chunk) {
	a.chunkCh <- chunk
}
