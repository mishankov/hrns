package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/mishankov/hrns/openai"
)

func TestMessageMarshalAndUnmarshalPreservesExtraFields(t *testing.T) {
	t.Parallel()

	original := openai.Message{
		Role:    "assistant",
		Content: "hello",
		Name:    "writer",
		Extra: map[string]any{
			"reasoning": map[string]any{"effort": "low"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal(original) error = %v", err)
	}

	var got openai.Message
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}

	if got.Role != "assistant" {
		t.Fatalf("got.Role = %q, want %q", got.Role, "assistant")
	}
	if openai.MessageText(got.Content) != "hello" {
		t.Fatalf("got.Content = %q, want %q", openai.MessageText(got.Content), "hello")
	}
	if got.Name != "writer" {
		t.Fatalf("got.Name = %q, want %q", got.Name, "writer")
	}

	reasoning, ok := got.Extra["reasoning"].(map[string]any)
	if !ok {
		t.Fatalf("got.Extra[reasoning] type = %T, want map[string]any", got.Extra["reasoning"])
	}
	if reasoning["effort"] != "low" {
		t.Fatalf("got.Extra[reasoning][effort] = %#v, want %q", reasoning["effort"], "low")
	}
}

func TestMessageConstructors(t *testing.T) {
	t.Parallel()

	system := openai.SystemMessage("rules")
	if system.Role != "system" || openai.MessageText(system.Content) != "rules" {
		t.Fatalf("SystemMessage() = %#v, want role/content set", system)
	}

	user := openai.UserMessage("hello")
	if user.Role != "user" || openai.MessageText(user.Content) != "hello" {
		t.Fatalf("UserMessage() = %#v, want role/content set", user)
	}

	tool := openai.ToolMessage("done", "call_1")
	if tool.Role != "tool" || openai.MessageText(tool.Content) != "done" || tool.ToolCallID != "call_1" {
		t.Fatalf("ToolMessage() = %#v, want role/content/tool_call_id set", tool)
	}
}
