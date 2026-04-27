package main

import (
	"context"
	"log"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/agents"
	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/skills"
	"github.com/mishankov/hrns/tools"
	"github.com/mishankov/hrns/tui"
)

func main() {
	ctx := context.Background()

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

	builtInAgents := []agent.Agent{agents.Builder, agents.Explorer, agents.Planner, agents.Pirate}
	fsAgents, err := agent.LoadAllAgents([]string{agent.DefaultLocalRootPath, agent.DefaultGlobalRootPath})
	if err != nil {
		log.Fatalf("failed to load agents from fs: %v", err)
	}

	agents := append(builtInAgents, fsAgents...)

	tuiapp := tui.New(
		tui.WithTools(map[string]loop.Tool{
			"read_file":   tools.ReadFileTool,
			"list_files":  tools.ListFilesTool,
			"write_file":  tools.WriteFileTool,
			"run_command": tools.CommandTool,
			"web_fetch":   tools.WebFetchTool,
			"load_skill":  loadSkillTool,
		}),
		tui.WithAgents(agents),
	)
	tuiapp.Run(ctx)
}
