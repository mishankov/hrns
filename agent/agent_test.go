package agent_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mishankov/hrns/agent"
)

func TestDiscoverAgentsFilesFindsOnlyTopLevelAgentFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "ignored.md"), "not an agent")
	writeFile(t, filepath.Join(root, "alpha", "AGENT.md"), agentFile("alpha", "Alpha agent", nil))
	writeFile(t, filepath.Join(root, "alpha", "notes.md"), "notes")
	writeFile(t, filepath.Join(root, "beta", "AGENT.md"), agentFile("beta", "Beta agent", nil))
	writeFile(t, filepath.Join(root, "beta", "nested", "AGENT.md"), agentFile("nested", "Nested agent", nil))

	files, err := agent.DiscoverAgentsFiles([]string{root})
	if err != nil {
		t.Fatalf("DiscoverAgentsFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(root, "alpha", "AGENT.md"),
		filepath.Join(root, "beta", "AGENT.md"),
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("DiscoverAgentsFiles() = %#v, want %#v", files, want)
	}
}

func TestDiscoverAgentsFilesIgnoresMissingRoots(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")

	files, err := agent.DiscoverAgentsFiles([]string{root})
	if err != nil {
		t.Fatalf("DiscoverAgentsFiles() error = %v, want nil", err)
	}
	if len(files) != 0 {
		t.Fatalf("DiscoverAgentsFiles() = %#v, want empty result", files)
	}
}

func TestDiscoverAgentsFilesExpandsHomeDirectory(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	dirName := filepath.Join(".hrns-test-agents", t.Name(), "demo")
	agentPath := filepath.Join(homeDir, dirName, "AGENT.md")
	writeFile(t, agentPath, agentFile("home-demo", "Home agent", nil))
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(homeDir, ".hrns-test-agents", t.Name()))
	})

	files, err := agent.DiscoverAgentsFiles([]string{"~/" + filepath.Dir(dirName)})
	if err != nil {
		t.Fatalf("DiscoverAgentsFiles() error = %v", err)
	}

	want := []string{agentPath}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("DiscoverAgentsFiles() = %#v, want %#v", files, want)
	}
}

func TestDiscoverAgentsFilesSkipsFilesAtRootLevel(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "AGENT.md"), agentFile("root", "Root agent", nil))
	writeFile(t, filepath.Join(root, "gamma", "AGENT.md"), agentFile("gamma", "Gamma agent", nil))

	files, err := agent.DiscoverAgentsFiles([]string{root})
	if err != nil {
		t.Fatalf("DiscoverAgentsFiles() error = %v", err)
	}

	want := []string{filepath.Join(root, "gamma", "AGENT.md")}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("DiscoverAgentsFiles() = %#v, want %#v", files, want)
	}
}

func TestGetAgentDataReadsFrontmatter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "demo", "AGENT.md")
	tools := map[string]bool{"tool1": true, "tool2": false}
	writeFile(t, path, agentFile("demo", "Demo agent", tools))

	ag, err := agent.GetAgentData(path)
	if err != nil {
		t.Fatalf("GetAgentData() error = %v", err)
	}

	if ag.Name != "demo" {
		t.Fatalf("ag.Name = %q, want %q", ag.Name, "demo")
	}
	if ag.Description != "Demo agent" {
		t.Fatalf("ag.Description = %q, want %q", ag.Description, "Demo agent")
	}
	if !reflect.DeepEqual(ag.Tools, tools) {
		t.Fatalf("ag.Tools = %#v, want %#v", ag.Tools, tools)
	}
	if ag.Prompt != "\n# Agent Prompt\n" {
		t.Fatalf("ag.Prompt = %q, want %q", ag.Prompt, "\n# Agent Prompt\n")
	}
}

func TestGetAgentDataReturnsErrorForMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "AGENT.md")

	_, err := agent.GetAgentData(path)
	if err == nil {
		t.Fatal("GetAgentData() error = nil, want non-nil")
	}
}

func TestGetAgentDataReturnsErrorForInvalidFrontmatter(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad", "AGENT.md")
	writeFile(t, path, "---\n{not yaml\n---\nbody\n")

	_, err := agent.GetAgentData(path)
	if err == nil {
		t.Fatal("GetAgentData() error = nil, want non-nil")
	}
}

func TestLoadAllAgentsLoadsAgentsFromAllRoots(t *testing.T) {
	t.Parallel()

	rootA := t.TempDir()
	rootB := t.TempDir()

	pathA := filepath.Join(rootA, "alpha", "AGENT.md")
	pathB := filepath.Join(rootB, "beta", "AGENT.md")
	writeFile(t, pathA, agentFile("alpha", "Alpha agent", nil))
	writeFile(t, pathB, agentFile("beta", "Beta agent", nil))

	loaded, err := agent.LoadAllAgents([]string{rootA, rootB})
	if err != nil {
		t.Fatalf("LoadAllAgents() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("len(loaded) = %d, want 2", len(loaded))
	}

	byName := make(map[string]string)
	for _, ag := range loaded {
		byName[ag.Name] = ag.Description
	}

	if desc, ok := byName["alpha"]; !ok || desc != "Alpha agent" {
		t.Fatalf("loaded agents missing alpha or wrong description: %#v", loaded)
	}
	if desc, ok := byName["beta"]; !ok || desc != "Beta agent" {
		t.Fatalf("loaded agents missing beta or wrong description: %#v", loaded)
	}
}

func TestLoadAllAgentsReturnsErrorForInvalidAgentFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "bad", "AGENT.md")
	writeFile(t, path, "---\n{not yaml\n---\n")

	_, err := agent.LoadAllAgents([]string{root})
	if err == nil {
		t.Fatal("LoadAllAgents() error = nil, want non-nil")
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

func agentFile(name, description string, tools map[string]bool) string {
	var toolsYAML string
	if len(tools) > 0 {
		toolsYAML = "\ntools:\n"
		for k, v := range tools {
			if v {
				toolsYAML += "  " + k + ": true\n"
			} else {
				toolsYAML += "  " + k + ": false\n"
			}
		}
	}
	return "---\nname: " + name + "\ndescription: " + description + toolsYAML + "\n---\n\n# Agent Prompt\n"
}
