package loop

import (
	"context"
	"encoding/json"

	"github.com/mishankov/hrns/openai"
)

type ChatCompletionStreamer interface {
	StreamChatCompletion(context.Context, openai.ChatCompletionRequest) (<-chan openai.StreamEvent, error)
}

type Loop struct {
	openAIClient ChatCompletionStreamer
	tools        map[string]Tool
	messages     []openai.Message
	chunkCh      chan Chunk
}

func New(openAIClient ChatCompletionStreamer, tools map[string]Tool) *Loop {
	return &Loop{
		openAIClient: openAIClient,
		tools:        tools,
		messages:     []openai.Message{},
		chunkCh:      make(chan Chunk),
	}
}

func (a *Loop) RunLoop(ctx context.Context, messages []openai.Message, model string) {
	// Composing tools for agent
	tools := []openai.Tool{}
	for name, tool := range a.tools {
		properties := map[string]map[string]string{}
		required := []string{}
		for _, argument := range tool.Arguments() {
			properties[argument.Name] = map[string]string{
				"type": argument.Type,
			}
			required = append(required, argument.Name)
		}

		tools = append(tools, openai.Tool{
			Type: "function",
			Function: map[string]any{
				"name":        name,
				"description": tool.Description(),
				"parameters": map[string]any{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		})
	}

	for {
		// Creating streaming chat completions object
		stream, err := a.openAIClient.StreamChatCompletion(ctx, openai.ChatCompletionRequest{
			Messages: messages,
			Tools:    tools,
			Model:    model,
		})
		if err != nil {
			a.sendChunk(NewChunkError("error in LLM stream: " + err.Error()))
			break
		}

		accumulator := openai.ChatCompletionAccumulator{}

		// Reading from response stream
		for event := range stream {
			if event.Error != nil {
				a.sendChunk(NewChunkError("error in LLM stream: " + event.Error.Error()))
				break
			}
			if event.Done || event.Data == nil {
				continue
			}

			if len(event.Data.Choices) == 0 {
				continue
			}

			chunk := event.Data
			accumulator.AddChunk(*chunk)

			delta := chunk.Choices[0].Delta
			if delta == nil {
				continue
			}

			if content := openai.MessageText(delta.Content); content != "" {
				// Process regular chunk
				a.sendChunk(NewChunkMessage(content))
			} else {
				// Process reasoning chunk
				if reasoning, _ := delta.Extra["reasoning"].(string); reasoning != "" {
					a.sendChunk(NewChunkReasoning(reasoning))
				}
			}
		}

		choices := accumulator.Choices()
		if len(choices) > 0 {
			messages = append(messages, choices[0].Message)
		}

		// Tool calling
		toolsCalled := false
		for _, choice := range choices {
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

func (a *Loop) Messages() []openai.Message {
	return a.messages
}

func (a *Loop) sendChunk(chunk Chunk) {
	a.chunkCh <- chunk
}
