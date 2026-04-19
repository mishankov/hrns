package skills

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

const DefaultRootPath = "~/.agents/skills"

func DisocoverSkillFile(rootPath string) ([]string, error) {
	if strings.HasPrefix(rootPath, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return []string{}, fmt.Errorf("Error getting user home dir: %w", err)
		}
		rootPath = filepath.Join(userHomeDir, rootPath[2:])
	}
	files := []string{}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
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
				if subEntry.Name() == "SKILL.md" {
					files = append(files, filepath.Join(path, subEntry.Name()))
				}

			}
		}
	}

	return files, nil
}

type Skill struct {
	Path        string
	Name        string
	Description string
}

type skillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func GetSkillData(path string) (*Skill, error) {
	fileReader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", path, err)
	}

	var metadata skillMetadata
	_, err = frontmatter.Parse(fileReader, &metadata)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s: %w", path, err)
	}

	return &Skill{
		Path:        path,
		Name:        metadata.Name,
		Description: metadata.Description,
	}, nil
}
