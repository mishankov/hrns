package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/mishankov/hrns/openai"
)

func TestToolCallMarshalAndUnmarshalPreservesExtraFields(t *testing.T) {
	t.Parallel()

	original := openai.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: openai.ToolCallFunction{
			Name:      "lookup_user",
			Arguments: "{\"id\":1}",
			Extra: map[string]any{
				"schema_version": "1",
			},
		},
		Extra: map[string]any{
			"status": "partial",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal(original) error = %v", err)
	}

	var got openai.ToolCall
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}

	if got.ID != "call_1" {
		t.Fatalf("got.ID = %q, want %q", got.ID, "call_1")
	}
	if got.Type != "function" {
		t.Fatalf("got.Type = %q, want %q", got.Type, "function")
	}
	if got.Function.Name != "lookup_user" {
		t.Fatalf("got.Function.Name = %q, want %q", got.Function.Name, "lookup_user")
	}
	if got.Function.Arguments != "{\"id\":1}" {
		t.Fatalf("got.Function.Arguments = %q, want %q", got.Function.Arguments, "{\"id\":1}")
	}
	if got.Extra["status"] != "partial" {
		t.Fatalf("got.Extra[status] = %#v, want %q", got.Extra["status"], "partial")
	}
	if got.Function.Extra["schema_version"] != "1" {
		t.Fatalf("got.Function.Extra[schema_version] = %#v, want %q", got.Function.Extra["schema_version"], "1")
	}
}
