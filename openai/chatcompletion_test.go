package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/mishankov/hrns/openai"
)

func TestChatCompletionRequestMarshalJSONIncludesExtraFields(t *testing.T) {
	t.Parallel()

	req := openai.ChatCompletionRequest{
		Model:    "gpt-test",
		Messages: []openai.Message{openai.UserMessage("hello")},
		Extra: map[string]any{
			"reasoning_effort": "medium",
			"store":            true,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal(req) error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}

	if got["model"] != "gpt-test" {
		t.Fatalf("model = %#v, want %q", got["model"], "gpt-test")
	}
	if got["reasoning_effort"] != "medium" {
		t.Fatalf("reasoning_effort = %#v, want %q", got["reasoning_effort"], "medium")
	}
	if got["store"] != true {
		t.Fatalf("store = %#v, want true", got["store"])
	}

	messages, ok := got["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("messages = %#v, want one message", got["messages"])
	}
}

func TestChatCompletionResponseUnmarshalCapturesExtraFieldsAndRawPayload(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"id":"resp_123",
		"object":"chat.completion",
		"model":"gpt-test",
		"choices":[
			{
				"index":0,
				"message":{
					"role":"assistant",
					"content":"hello",
					"reasoning":{"summary":"kept"}
				},
				"finish_reason":"stop",
				"provider":"test-provider"
			}
		],
		"usage":{"prompt_tokens":10},
		"trace_id":"trace-123"
	}`)

	var resp openai.ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}

	if resp.ID != "resp_123" {
		t.Fatalf("resp.ID = %q, want %q", resp.ID, "resp_123")
	}
	if resp.Object != "chat.completion" {
		t.Fatalf("resp.Object = %q, want %q", resp.Object, "chat.completion")
	}
	if resp.Model != "gpt-test" {
		t.Fatalf("resp.Model = %q, want %q", resp.Model, "gpt-test")
	}
	if resp.Extra["trace_id"] != "trace-123" {
		t.Fatalf("resp.Extra[trace_id] = %#v, want %q", resp.Extra["trace_id"], "trace-123")
	}
	if len(resp.Raw) == 0 {
		t.Fatal("resp.Raw = empty, want captured raw payload")
	}
	if string(resp.Raw["trace_id"]) != `"trace-123"` {
		t.Fatalf("resp.Raw[trace_id] = %s, want %q", string(resp.Raw["trace_id"]), `"trace-123"`)
	}
	if string(resp.Raw["choices"]) == "" {
		t.Fatal("resp.Raw[choices] = empty, want raw choices payload")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("len(resp.Choices) = %d, want 1", len(resp.Choices))
	}

	choice := resp.Choices[0]
	if choice.Extra["provider"] != "test-provider" {
		t.Fatalf("choice.Extra[provider] = %#v, want %q", choice.Extra["provider"], "test-provider")
	}
	if choice.Message.Role != "assistant" {
		t.Fatalf("choice.Message.Role = %q, want %q", choice.Message.Role, "assistant")
	}
	if openai.MessageText(choice.Message.Content) != "hello" {
		t.Fatalf("choice.Message.Content = %q, want %q", openai.MessageText(choice.Message.Content), "hello")
	}

	reasoning, ok := choice.Message.Extra["reasoning"].(map[string]any)
	if !ok {
		t.Fatalf("choice.Message.Extra[reasoning] type = %T, want map[string]any", choice.Message.Extra["reasoning"])
	}
	if reasoning["summary"] != "kept" {
		t.Fatalf("choice.Message.Extra[reasoning][summary] = %#v, want %q", reasoning["summary"], "kept")
	}
}
