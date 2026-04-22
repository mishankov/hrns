package openai_test

import (
	"testing"

	"github.com/mishankov/hrns/openai"
)

func TestChatCompletionAccumulatorChoicesStartsEmpty(t *testing.T) {
	t.Parallel()

	var acc openai.ChatCompletionAccumulator

	if choices := acc.Choices(); choices != nil {
		t.Fatalf("Choices() = %#v, want nil", choices)
	}
}

func TestChatCompletionAccumulatorAccumulatesChoicesAcrossChunks(t *testing.T) {
	t.Parallel()

	var acc openai.ChatCompletionAccumulator

	acc.AddChunk(openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 1,
				Delta: &openai.Message{
					Role:    "assistant",
					Content: "Hel",
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: openai.ToolCallFunction{
								Name:      "lookup_user",
								Arguments: "{\"user\":",
							},
						},
					},
				},
			},
			{
				Index: 0,
				Message: openai.Message{
					Role:    "assistant",
					Content: "ready",
				},
				FinishReason: "stop",
			},
		},
	})

	acc.AddChunk(openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 1,
				Delta: &openai.Message{
					Content: "lo",
					ToolCalls: []openai.ToolCall{
						{
							Function: openai.ToolCallFunction{
								Arguments: "\"alice\"}",
							},
						},
					},
				},
				Extra: map[string]any{
					"provider": "test-provider",
				},
			},
		},
	})

	logprobs := map[string]any{"tokens": []any{"Hello"}}
	acc.AddChunk(openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        1,
				FinishReason: "tool_calls",
				Logprobs:     logprobs,
			},
		},
	})

	choices := acc.Choices()
	if len(choices) != 2 {
		t.Fatalf("len(Choices()) = %d, want 2", len(choices))
	}

	if choices[0].Index != 0 {
		t.Fatalf("choices[0].Index = %d, want 0", choices[0].Index)
	}
	if choices[0].FinishReason != "stop" {
		t.Fatalf("choices[0].FinishReason = %q, want %q", choices[0].FinishReason, "stop")
	}
	if choices[0].Message.Role != "assistant" {
		t.Fatalf("choices[0].Message.Role = %q, want %q", choices[0].Message.Role, "assistant")
	}
	if got := openai.MessageText(choices[0].Message.Content); got != "ready" {
		t.Fatalf("choices[0].Message.Content = %q, want %q", got, "ready")
	}

	if choices[1].Index != 1 {
		t.Fatalf("choices[1].Index = %d, want 1", choices[1].Index)
	}
	if choices[1].Message.Role != "assistant" {
		t.Fatalf("choices[1].Message.Role = %q, want %q", choices[1].Message.Role, "assistant")
	}
	if got := openai.MessageText(choices[1].Message.Content); got != "Hello" {
		t.Fatalf("choices[1].Message.Content = %q, want %q", got, "Hello")
	}
	if choices[1].FinishReason != "tool_calls" {
		t.Fatalf("choices[1].FinishReason = %q, want %q", choices[1].FinishReason, "tool_calls")
	}
	if len(choices[1].Message.ToolCalls) != 1 {
		t.Fatalf("len(choices[1].Message.ToolCalls) = %d, want 1", len(choices[1].Message.ToolCalls))
	}

	toolCall := choices[1].Message.ToolCalls[0]
	if toolCall.ID != "call_1" {
		t.Fatalf("toolCall.ID = %q, want %q", toolCall.ID, "call_1")
	}
	if toolCall.Type != "function" {
		t.Fatalf("toolCall.Type = %q, want %q", toolCall.Type, "function")
	}
	if toolCall.Function.Name != "lookup_user" {
		t.Fatalf("toolCall.Function.Name = %q, want %q", toolCall.Function.Name, "lookup_user")
	}
	if toolCall.Function.Arguments != "{\"user\":\"alice\"}" {
		t.Fatalf("toolCall.Function.Arguments = %q, want %q", toolCall.Function.Arguments, "{\"user\":\"alice\"}")
	}
	if choices[1].Extra["provider"] != "test-provider" {
		t.Fatalf("choices[1].Extra[provider] = %#v, want %q", choices[1].Extra["provider"], "test-provider")
	}
	if got, ok := choices[1].Logprobs.(map[string]any); !ok || len(got) != 1 || got["tokens"] == nil {
		t.Fatalf("choices[1].Logprobs = %#v, want tokens payload", choices[1].Logprobs)
	}
}

func TestChatCompletionAccumulatorPreservesStructuredMessageContent(t *testing.T) {
	t.Parallel()

	var acc openai.ChatCompletionAccumulator
	content := []any{
		map[string]any{
			"type": "output_text",
			"text": "structured",
		},
	}

	acc.AddChunk(openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Delta: &openai.Message{
					Role:    "assistant",
					Content: content,
				},
			},
		},
	})

	choices := acc.Choices()
	if len(choices) != 1 {
		t.Fatalf("len(Choices()) = %d, want 1", len(choices))
	}
	if choices[0].Message.Content == nil {
		t.Fatal("choices[0].Message.Content = nil, want structured content")
	}

	got, ok := choices[0].Message.Content.([]any)
	if !ok {
		t.Fatalf("choices[0].Message.Content type = %T, want []any", choices[0].Message.Content)
	}
	if len(got) != 1 {
		t.Fatalf("len(choices[0].Message.Content) = %d, want 1", len(got))
	}
	part, ok := got[0].(map[string]any)
	if !ok {
		t.Fatalf("choices[0].Message.Content[0] type = %T, want map[string]any", got[0])
	}
	if part["text"] != "structured" {
		t.Fatalf("choices[0].Message.Content[0][text] = %#v, want %q", part["text"], "structured")
	}
}
