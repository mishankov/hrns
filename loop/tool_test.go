package loop_test

import (
	"reflect"
	"testing"

	"github.com/mishankov/hrns/loop"
)

func TestNewSimpleTool(t *testing.T) {
	t.Parallel()

	called := false
	tool := loop.NewSimpleTool(
		"Echoes value",
		[]loop.ToolArgument{{Name: "value", Type: "string"}},
		func(args map[string]any) string {
			called = true
			return args["value"].(string)
		},
	)

	if got := tool.Description(); got != "Echoes value" {
		t.Fatalf("Description() = %q, want %q", got, "Echoes value")
	}
	if got := tool.Arguments(); !reflect.DeepEqual(got, []loop.ToolArgument{{Name: "value", Type: "string"}}) {
		t.Fatalf("Arguments() = %#v", got)
	}
	if got := tool.Call(map[string]any{"value": "hello"}); got != "hello" {
		t.Fatalf("Call() = %q, want %q", got, "hello")
	}
	if !called {
		t.Fatal("Call() did not invoke the provided function")
	}
}
