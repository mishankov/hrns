package tui

import (
	"bufio"
	"context"
	"os"
	"slices"
	"strings"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/openai"
)

type TUIApp struct {
}

func New() *TUIApp {
	return &TUIApp{}
}

func (app *TUIApp) Run(ctx context.Context, agnt loop.Loop) {
	messages := []openai.Message{}

	model := "kimi-k2.6"

	PrintHarnessMessage("HRNS loop. dev")
	PrintHarnessMessage("Model: " + model)

	for {
		reader := bufio.NewReader(os.Stdin)
		PrintUserInputPrompt()
		messageText, _ := reader.ReadString('\n')
		messageText = strings.TrimSpace(messageText)

		if strings.HasPrefix(messageText, "/") {
			commandSplited := strings.Split(messageText, " ")

			switch commandSplited[0] {
			case "/model":
				model = strings.TrimSpace(commandSplited[1])
				PrintHarnessMessage("Model changed to " + model)
			case "/new":
				messages = []openai.Message{}
				PrintHarnessMessage("New session started")
			case "/help":
				PrintHarnessMessage("Available commands:")
				PrintHarnessMessage("/model <model> - change model")
				PrintHarnessMessage("/new - start new session")
				PrintHarnessMessage("/help - show this help")
			default:
				PrintError("unknown command: " + commandSplited[0])
			}

			continue
		}

		messages = append(messages, openai.UserMessage(messageText))

		go agnt.RunLoop(ctx, messages, model)

		lastChunkType := loop.ChunkType("")
		for chunk := range agnt.Chunks() {
			toBreak := false

			if chunk.Type != lastChunkType {
				if !(slices.Contains([]loop.ChunkType{loop.ChunkTypeToolCallResult, loop.ChunkTypeToolCallError, loop.ChunkTypeToolCallStart}, lastChunkType) && slices.Contains([]loop.ChunkType{loop.ChunkTypeToolCallResult, loop.ChunkTypeToolCallError, loop.ChunkTypeToolCallStart}, chunk.Type)) {
					PrintNewLine()
				}

				if slices.Contains([]loop.ChunkType{loop.ChunkTypeMessage, loop.ChunkTypeReasoning}, lastChunkType) {
					PrintNewLine()
				}
			}

			switch chunk.Type {
			case loop.ChunkTypeMessage:
				PrintResponseChunc(chunk.Text)
			case loop.ChunkTypeReasoning:
				PrintReasoningChunc(chunk.Text)
			case loop.ChunkTypeToolCallStart:
				PrintToolCall(chunk.ToolName, chunk.ToolArgs)
			case loop.ChunkTypeToolCallResult, loop.ChunkTypeToolCallError:
				func() {}()
			case loop.ChunkTypeError:
				PrintError(chunk.Text)
			case loop.ChunkTypeEnd:
				toBreak = true
			default:
				PrintHarnessMessage("other chunk " + string(chunk.Type))
			}

			lastChunkType = chunk.Type

			if toBreak {
				break
			}
		}

		messages = agnt.Messages()
	}
}
