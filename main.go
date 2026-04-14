package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/terminal"
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

	messages := []openai.ChatCompletionMessageParamUnion{}

	agnt := agent.New(
		&client,
		"You are a coding assistant that talks like a pirate.",
		"z-ai/glm-5.1",
		map[string]agent.Tool{
			"read_file": agent.NewTool(
				"Reads file from filesystem",
				[]agent.ToolArgument{{Name: "fileName", Type: "string"}},
				func(args map[string]any) string {
					dat, err := os.ReadFile(args["fileName"].(string))
					if err != nil {
						return "ERROR: tools calling error: " + err.Error()
					} else {
						return string(dat)
					}
				},
			),
			"list_files": agent.NewTool(
				"Lists files in directory using glob pattern",
				[]agent.ToolArgument{{Name: "dir", Type: "string"}, {Name: "globPattern", Type: "string"}},
				func(args map[string]any) string {
					root := os.DirFS(args["dir"].(string))

					mdFiles, err := fs.Glob(root, args["globPattern"].(string))

					if err != nil {
						log.Fatal(err)
					}

					var files []string
					for _, v := range mdFiles {
						files = append(files, path.Join(args["dir"].(string), v))
					}

					data, err := json.Marshal(files)
					if err != nil {
						return "ERROR: tools calling error: " + err.Error()
					} else {
						return string(data)
					}
				},
			),
		},
	)

	for {
		reader := bufio.NewReader(os.Stdin)
		terminal.PrintUserInputPrompt()
		messageText, _ := reader.ReadString('\n')
		messages = append(messages, openai.UserMessage(messageText))

		go agnt.RunLoop(ctx, messages)

		for chunk := range agnt.Chunks() {
			toBreak := false

			switch chunk.Type {
			case agent.ChunkTypeMessage:
				terminal.PrintResponse(chunk.Text)
			case agent.ChunkTypeReasoning:
				terminal.PrintReasoning(chunk.Text)
			case agent.ChunkTypeToolCallStart:
				terminal.PrintToolCall(chunk.ToolName, chunk.ToolArgs)
			case agent.ChunkTypeToolCallResult:
			case agent.ChunkTypeError:
				terminal.PrintError(chunk.Text)
			case agent.ChunkTypeEnd:
				toBreak = true
			}

			if toBreak {
				break
			}
		}
	}
}
