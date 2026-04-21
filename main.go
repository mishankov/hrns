package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/openai"
	"github.com/mishankov/hrns/skills"
	"github.com/mishankov/hrns/tools"
	"github.com/mishankov/hrns/tui"
)

func main() {
	ctx := context.Background()

	key := os.Getenv("HRNS_KEY")
	baseUrl := os.Getenv("HRNS_BASE_URL")
	skipVerify := os.Getenv("HRNS_SKIP_VERIFY") == "true"
	client := openai.NewClient(
		openai.WithBaseURL(baseUrl),
		openai.WithAPIKey(key),
		openai.WithHTTPClient(
			&http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipVerify,
				},
			}}),
	)

	loadedSkills, err := skills.LoadAllSkills([]string{skills.DefaultGlobalRootPath, skills.DefaultLocalRootPath})
	if err != nil {
		log.Fatalf("failed to load skills: %v", err)
	}
	loadSkillTool := skills.NewLoadSkillTool(loadedSkills)

	systemPrompt := "You are a coding assistant that talks like a pirate."
	if len(loadedSkills) > 0 {
		systemPrompt += "\n\nYou have access to the following skills:"
		for _, skill := range loadedSkills {
			systemPrompt += "\n- " + skill.Name + ": " + skill.Description
		}

		systemPrompt += "\n You can load them with the `load_skill` tool."
	}

	agnt := loop.New(
		client,
		systemPrompt,
		map[string]loop.Tool{
			"read_file":   tools.ReadFileTool,
			"list_files":  tools.ListFilesTool,
			"write_file":  tools.WriteFileTool,
			"run_command": tools.CommandTool,
			"web_fetch":   tools.WebFetchTool,
			"load_skill":  loadSkillTool,
		},
	)

	tuiapp := tui.New()

	tuiapp.Run(ctx, *agnt)
}
