package tools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/tools"
)

func TestReadFileTool(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "demo.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}

	if got := tools.ReadFileTool.Description(); got != "Reads file from filesystem" {
		t.Fatalf("Description() = %q, want %q", got, "Reads file from filesystem")
	}
	if got := tools.ReadFileTool.Arguments(); len(got) != 1 || got[0] != (loop.ToolArgument{Name: "fileName", Type: "string"}) {
		t.Fatalf("Arguments() = %#v, want one fileName argument", got)
	}
	if got := tools.ReadFileTool.Call(map[string]any{"fileName": path}); got != "hello" {
		t.Fatalf("Call() = %q, want %q", got, "hello")
	}
}

func TestReadFileToolReturnsErrorForMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.txt")

	got := tools.ReadFileTool.Call(map[string]any{"fileName": path})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestReadFileToolReturnsErrorForMissingOrInvalidArgument(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args map[string]any
	}{
		{
			name: "missing argument",
			args: map[string]any{},
		},
		{
			name: "invalid type",
			args: map[string]any{"fileName": 123},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tools.ReadFileTool.Call(tc.args)
			if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
				t.Fatalf("Call() = %q, want tools calling error", got)
			}
		})
	}
}

func TestListFilesTool(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.md"), "a")
	writeFile(t, filepath.Join(root, "b.txt"), "b")
	writeFile(t, filepath.Join(root, "nested", "c.md"), "c")

	if got := tools.ListFilesTool.Description(); got != "Lists files in directory using glob pattern" {
		t.Fatalf("Description() = %q, want %q", got, "Lists files in directory using glob pattern")
	}
	wantArgs := []loop.ToolArgument{
		{Name: "dir", Type: "string"},
		{Name: "globPattern", Type: "string"},
	}
	if got := tools.ListFilesTool.Arguments(); len(got) != 2 || got[0] != wantArgs[0] || got[1] != wantArgs[1] {
		t.Fatalf("Arguments() = %#v, want %#v", got, wantArgs)
	}

	got := tools.ListFilesTool.Call(map[string]any{
		"dir":         root,
		"globPattern": "*.md",
	})

	var files []string
	if err := json.Unmarshal([]byte(got), &files); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", got, err)
	}

	want := []string{filepath.Join(root, "a.md")}
	if len(files) != len(want) || files[0] != want[0] {
		t.Fatalf("Call() = %#v, want %#v", files, want)
	}
}

func TestWriteFileToolCreatesMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "new.txt")

	if got := tools.WriteFileTool.Description(); got != "Replaces first occurence of oldString with newString in a file" {
		t.Fatalf("Description() = %q, want %q", got, "Replaces first occurence of oldString with newString in a file")
	}
	wantArgs := []loop.ToolArgument{
		{Name: "fileName", Type: "string"},
		{Name: "oldString", Type: "string"},
		{Name: "newString", Type: "string"},
	}
	if got := tools.WriteFileTool.Arguments(); len(got) != 3 || got[0] != wantArgs[0] || got[1] != wantArgs[1] || got[2] != wantArgs[2] {
		t.Fatalf("Arguments() = %#v, want %#v", got, wantArgs)
	}

	got := tools.WriteFileTool.Call(map[string]any{
		"fileName":  path,
		"oldString": "ignored",
		"newString": "created",
	})
	if got != "OK" {
		t.Fatalf("Call() = %q, want %q", got, "OK")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if string(data) != "created" {
		t.Fatalf("file content = %q, want %q", string(data), "created")
	}
}

func TestWriteFileToolReplacesOnlyFirstOccurrence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "demo.txt")
	writeFile(t, path, "old old old")

	got := tools.WriteFileTool.Call(map[string]any{
		"fileName":  path,
		"oldString": "old",
		"newString": "new",
	})
	if got != "OK" {
		t.Fatalf("Call() = %q, want %q", got, "OK")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if string(data) != "new old old" {
		t.Fatalf("file content = %q, want %q", string(data), "new old old")
	}
}

func TestCommandTool(t *testing.T) {
	t.Parallel()

	if got := tools.CommandTool.Description(); got != "Runs shell command" {
		t.Fatalf("Description() = %q, want %q", got, "Runs shell command")
	}
	if got := tools.CommandTool.Arguments(); len(got) != 1 || got[0] != (loop.ToolArgument{Name: "command", Type: "string"}) {
		t.Fatalf("Arguments() = %#v, want one command argument", got)
	}

	command := "printf hello"
	if runtime.GOOS == "windows" {
		command = "echo hello"
	}

	got := tools.CommandTool.Call(map[string]any{"command": command})
	if runtime.GOOS == "windows" {
		got = strings.TrimSpace(got)
	}
	if got != "hello" {
		t.Fatalf("Call() = %q, want %q", got, "hello")
	}
}

func TestCommandToolReturnsErrorForFailingCommand(t *testing.T) {
	t.Parallel()

	command := "exit 3"
	if runtime.GOOS == "windows" {
		command = "exit /b 3"
	}

	got := tools.CommandTool.Call(map[string]any{"command": command})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestCommandToolReturnsErrorForInvalidArgument(t *testing.T) {
	t.Parallel()

	got := tools.CommandTool.Call(map[string]any{"command": 123})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestWebFetchTool(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("response body"))
	}))
	defer server.Close()

	if got := tools.WebFetchTool.Description(); got != "Fetches content from URL" {
		t.Fatalf("Description() = %q, want %q", got, "Fetches content from URL")
	}
	if got := tools.WebFetchTool.Arguments(); len(got) != 1 || got[0] != (loop.ToolArgument{Name: "url", Type: "string"}) {
		t.Fatalf("Arguments() = %#v, want one url argument", got)
	}

	got := tools.WebFetchTool.Call(map[string]any{"url": server.URL})
	if got != "response body" {
		t.Fatalf("Call() = %q, want %q", got, "response body")
	}
}

func TestWebFetchToolReturnsErrorForInvalidURL(t *testing.T) {
	t.Parallel()

	got := tools.WebFetchTool.Call(map[string]any{"url": "://bad url"})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestWriteFileToolReturnsErrorForInvalidArgument(t *testing.T) {
	t.Parallel()

	got := tools.WriteFileTool.Call(map[string]any{
		"fileName": 123,
	})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestListFilesToolReturnsErrorForInvalidArgument(t *testing.T) {
	t.Parallel()

	got := tools.ListFilesTool.Call(map[string]any{
		"dir":         t.TempDir(),
		"globPattern": 123,
	})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func TestWebFetchToolReturnsErrorForMissingArgument(t *testing.T) {
	t.Parallel()

	got := tools.WebFetchTool.Call(map[string]any{})
	if !strings.HasPrefix(got, "ERROR: tools calling error: ") {
		t.Fatalf("Call() = %q, want tools calling error", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
