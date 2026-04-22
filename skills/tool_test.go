package skills_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/skills"
)

func TestNewLoadSkillToolExposesMetadata(t *testing.T) {
	t.Parallel()

	tool := skills.NewLoadSkillTool(nil)

	if got := tool.Description(); got != "Loads full skill body by its name" {
		t.Fatalf("Description() = %q, want %q", got, "Loads full skill body by its name")
	}

	want := []loop.ToolArgument{{Name: "name", Type: "string"}}
	if got := tool.Arguments(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Arguments() = %#v, want %#v", got, want)
	}
}

func TestLoadSkillToolCallReturnsSkillBody(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "demo", "SKILL.md")
	content := skillFile("demo", "Demo skill")
	writeFile(t, path, content)

	tool := skills.NewLoadSkillTool([]skills.Skill{
		{Name: "demo", Path: path, Description: "Demo skill"},
	})

	got := tool.Call(map[string]any{"name": "demo"})
	if got != content {
		t.Fatalf("Call() = %q, want %q", got, content)
	}
}

func TestLoadSkillToolCallReturnsErrorsForUnknownOrUnreadableSkill(t *testing.T) {
	t.Parallel()

	tool := skills.NewLoadSkillTool([]skills.Skill{
		{Name: "missing", Path: filepath.Join(t.TempDir(), "missing", "SKILL.md")},
	})

	if got := tool.Call(map[string]any{"name": "unknown"}); got != "ERROR: unknown skill name" {
		t.Fatalf("Call(unknown) = %q, want %q", got, "ERROR: unknown skill name")
	}

	got := tool.Call(map[string]any{"name": "missing"})
	if !strings.HasPrefix(got, "ERROR: loading skill file:") {
		t.Fatalf("Call(unreadable) = %q, want loading error", got)
	}
}
