package skills_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mishankov/hrns/skills"
)

func TestDiscoverSkillFilesFindsOnlyTopLevelSkillFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "ignored.md"), "not a skill")
	writeFile(t, filepath.Join(root, "alpha", "SKILL.md"), skillFile("alpha", "Alpha skill"))
	writeFile(t, filepath.Join(root, "alpha", "notes.md"), "notes")
	writeFile(t, filepath.Join(root, "beta", "SKILL.md"), skillFile("beta", "Beta skill"))
	writeFile(t, filepath.Join(root, "beta", "nested", "SKILL.md"), skillFile("nested", "Nested skill"))

	files, err := skills.DiscoverSkillFiles([]string{root})
	if err != nil {
		t.Fatalf("DiscoverSkillFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "alpha", "SKILL.md"),
		filepath.Join(root, "beta", "SKILL.md"),
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("DiscoverSkillFiles() = %#v, want %#v", files, want)
	}
}

func TestDiscoverSkillFilesIgnoresMissingRoots(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")

	files, err := skills.DiscoverSkillFiles([]string{root})
	if err != nil {
		t.Fatalf("DiscoverSkillFiles() error = %v, want nil", err)
	}
	if len(files) != 0 {
		t.Fatalf("DiscoverSkillFiles() = %#v, want empty result", files)
	}
}

func TestDiscoverSkillFilesExpandsHomeDirectory(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	dirName := filepath.Join(".hrns-test-skills", t.Name(), "demo")
	skillPath := filepath.Join(homeDir, dirName, "SKILL.md")
	writeFile(t, skillPath, skillFile("home-demo", "Home skill"))
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(homeDir, ".hrns-test-skills", t.Name()))
	})

	files, err := skills.DiscoverSkillFiles([]string{"~/" + filepath.Dir(dirName)})
	if err != nil {
		t.Fatalf("DiscoverSkillFiles() error = %v", err)
	}

	want := []string{skillPath}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("DiscoverSkillFiles() = %#v, want %#v", files, want)
	}
}

func TestGetSkillDataReadsFrontmatter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "demo", "SKILL.md")
	writeFile(t, path, skillFile("demo", "Demo skill"))

	skill, err := skills.GetSkillData(path)
	if err != nil {
		t.Fatalf("GetSkillData() error = %v", err)
	}

	if skill.Path != path {
		t.Fatalf("skill.Path = %q, want %q", skill.Path, path)
	}
	if skill.Name != "demo" {
		t.Fatalf("skill.Name = %q, want %q", skill.Name, "demo")
	}
	if skill.Description != "Demo skill" {
		t.Fatalf("skill.Description = %q, want %q", skill.Description, "Demo skill")
	}
}

func TestLoadAllSkillsLoadsSkillsFromAllRoots(t *testing.T) {
	t.Parallel()

	rootA := t.TempDir()
	rootB := t.TempDir()

	pathA := filepath.Join(rootA, "alpha", "SKILL.md")
	pathB := filepath.Join(rootB, "beta", "SKILL.md")
	writeFile(t, pathA, skillFile("alpha", "Alpha skill"))
	writeFile(t, pathB, skillFile("beta", "Beta skill"))

	loaded, err := skills.LoadAllSkills([]string{rootA, rootB})
	if err != nil {
		t.Fatalf("LoadAllSkills() error = %v", err)
	}

	want := []skills.Skill{
		{Path: pathA, Name: "alpha", Description: "Alpha skill"},
		{Path: pathB, Name: "beta", Description: "Beta skill"},
	}
	if !reflect.DeepEqual(loaded, want) {
		t.Fatalf("LoadAllSkills() = %#v, want %#v", loaded, want)
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

func skillFile(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\n# Skill\n"
}
