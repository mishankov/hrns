package tui

import (
	"bufio"
	"context"
	"os"
	"slices"
	"strings"

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

	model := "glm-5.1"

	terminal.PrintHarnessMessage("HRNS agent. dev")
	terminal.PrintHarnessMessage("Model: " + model)

	for {
		reader := bufio.NewReader(os.Stdin)
		terminal.PrintUserInputPrompt()
		messageText, _ := reader.ReadString('\n')
		messageText = strings.TrimSpace(messageText)

		if strings.HasPrefix(messageText, "/") {
			commandSplited := strings.Split(messageText, " ")

			switch commandSplited[0] {
			case "/model":
				model = strings.TrimSpace(commandSplited[1])
				terminal.PrintHarnessMessage("Model changed to " + model)
			default:
				terminal.PrintError("unknown command: " + commandSplited[0])
			}

			continue
		}

		messages = append(messages, openai.UserMessage(messageText))

		go agnt.RunLoop(ctx, messages, model)

		lastChunkType := agent.ChunkType("")
		for chunk := range agnt.Chunks() {
			toBreak := false

			if chunk.Type != lastChunkType {
				if !(slices.Contains([]agent.ChunkType{agent.ChunkTypeToolCallResult, agent.ChunkTypeToolCallError, agent.ChunkTypeToolCallStart}, lastChunkType) && slices.Contains([]agent.ChunkType{agent.ChunkTypeToolCallResult, agent.ChunkTypeToolCallError, agent.ChunkTypeToolCallStart}, chunk.Type)) {
					terminal.PrintNewLine()
				}

				if slices.Contains([]agent.ChunkType{agent.ChunkTypeMessage, agent.ChunkTypeReasoning}, lastChunkType) {
					terminal.PrintNewLine()
				}
			}

			switch chunk.Type {
			case agent.ChunkTypeMessage:
				terminal.PrintResponseChunc(chunk.Text)
			case agent.ChunkTypeReasoning:
				terminal.PrintReasoningChunc(chunk.Text)
			case agent.ChunkTypeToolCallStart:
				terminal.PrintToolCall(chunk.ToolName, chunk.ToolArgs)
			case agent.ChunkTypeToolCallResult, agent.ChunkTypeToolCallError:
				func() {}()
			case agent.ChunkTypeError:
				terminal.PrintError(chunk.Text)
			case agent.ChunkTypeEnd:
				toBreak = true
			default:
				terminal.PrintHarnessMessage("other chunk " + string(chunk.Type))
			}

			lastChunkType = chunk.Type

			if toBreak {
				break
			}
		}

		messages = agnt.Messages()
	}
}
