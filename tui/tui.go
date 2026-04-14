package tui

import (
	"bufio"
	"context"
	"os"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/terminal"
	"github.com/openai/openai-go/v3"
)

type TUIApp struct {
}

func New() *TUIApp {
	return &TUIApp{}
}

func (app *TUIApp) Run(ctx context.Context, agnt agent.Agent) {
	messages := []openai.ChatCompletionMessageParamUnion{}

	for {
		reader := bufio.NewReader(os.Stdin)
		terminal.PrintUserInputPrompt()
		messageText, _ := reader.ReadString('\n')
		messages = append(messages, openai.UserMessage(messageText))

		go agnt.RunLoop(ctx, messages, "z-ai/glm-5.1")

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
