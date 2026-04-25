package agent

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

type Agent struct {
	Path        string
	Name        string
	Description string
	Prompt      string
	Tools       map[string]bool
}

const DefaultGlobalRootPath = "~/.hrns/agents"
const DefaultLocalRootPath = "./.hrns/agents"

func DiscoverAgentsFiles(rootPaths []string) ([]string, error) {
	files := []string{}
	for _, rootPath := range rootPaths {
		if strings.HasPrefix(rootPath, "~/") {
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return []string{}, fmt.Errorf("Error getting user home dir: %w", err)
			}
			rootPath = filepath.Join(userHomeDir, rootPath[2:])
		}

		entries, err := os.ReadDir(rootPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Fatal(err)
		}

		for _, entry := range entries {
			path := filepath.Join(rootPath, entry.Name())

			// Skip file in root
			if !entry.IsDir() {
				continue
			}

			// If it's a directory, go one level deep
			subEntries, err := os.ReadDir(path)
			if err != nil {
				return []string{}, fmt.Errorf("Error reading %s: %w\n", path, err)
			}

			for _, subEntry := range subEntries {
				// Only list files at this level, don't recurse further
				if !subEntry.IsDir() {
					if subEntry.Name() == "AGENT.md" {
						files = append(files, filepath.Join(path, subEntry.Name()))
					}

				}
			}
		}
	}

	return files, nil
}

func GetAgentData(path string) (*Agent, error) {
	fileReader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", path, err)
	}
	defer fileReader.Close()

	var agent Agent

	prompt, err := frontmatter.Parse(fileReader, &agent)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s: %w", path, err)
	}

	agent.Prompt = string(prompt)

	return &agent, nil
}

func LoadAllAgents(rootPaths []string) ([]Agent, error) {
	files, err := DiscoverAgentsFiles(rootPaths)
	if err != nil {
		return nil, err
	}

	agents := make([]Agent, 0, len(files))
	for _, file := range files {
		agent, err := GetAgentData(file)
		if err != nil {
			return nil, err
		}
		agents = append(agents, *agent)
	}

	return agents, nil
}
