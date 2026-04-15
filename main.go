package main

import (
	"context"
	"os"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/tools"
	"github.com/mishankov/hrns/tui"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	ctx := context.Background()

	key, _ := os.LookupEnv("HRNS_KEY")
	baseUrl, _ := os.LookupEnv("HRNS_BASE_URL")
	client := openai.NewClient(
		option.WithAPIKey(key),
		option.WithBaseURL(baseUrl),
	)

	agnt := agent.New(
		&client,
		"You are a coding assistant that talks like a pirate.",
		map[string]agent.Tool{
			"read_file":   tools.ReadFileTool,
			"list_files":  tools.ListFilesTool,
			"write_file":  tools.WriteFileTool,
			"run_command": tools.CommandTool,
		},
	)

	tuiapp := tui.New()

	tuiapp.Run(ctx, *agnt)
}
