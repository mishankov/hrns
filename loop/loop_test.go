package loop_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/openai"
)

func TestRunLoopStreamsAssistantResponseAndStoresMessages(t *testing.T) {
	t.Parallel()

	initialMessages := []openai.Message{
		openai.SystemMessage("system prompt"),
		openai.UserMessage("Hi"),
	}

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Role: "assistant", Content: "Hello"}),
					streamEvent(0, openai.Message{Content: " world"}),
				},
			},
		},
	}

	agent := loop.New(streamer, map[string]loop.Tool{
		"echo": loop.NewSimpleTool(
			"Echoes text",
			[]loop.ToolArgument{{Name: "value", Type: "string"}},
			func(args map[string]any) string { return args["value"].(string) },
		),
	})

	chunks := runLoop(t, agent, initialMessages, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeMessage,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].Text != "Hello" || chunks[1].Text != " world" {
		t.Fatalf("message chunks = %#v, want Hello / world", chunks[:2])
	}

	if len(streamer.calls) != 1 {
		t.Fatalf("stream calls = %d, want 1", len(streamer.calls))
	}
	req := streamer.calls[0]
	if req.Model != "test-model" {
		t.Fatalf("request model = %q, want %q", req.Model, "test-model")
	}
	if len(req.Messages) != 2 {
		t.Fatalf("request messages len = %d, want 2", len(req.Messages))
	}
	if req.Messages[0].Role != "system" || openai.MessageText(req.Messages[0].Content) != "system prompt" {
		t.Fatalf("request system message = %#v", req.Messages[0])
	}
	if req.Messages[1].Role != "user" || openai.MessageText(req.Messages[1].Content) != "Hi" {
		t.Fatalf("request user message = %#v", req.Messages[1])
	}
	if len(req.Tools) != 1 {
		t.Fatalf("request tools len = %d, want 1", len(req.Tools))
	}
	if req.Tools[0].Type != "function" {
		t.Fatalf("request tool type = %q, want function", req.Tools[0].Type)
	}
	if req.Tools[0].Function["name"] != "echo" {
		t.Fatalf("request tool name = %#v, want %q", req.Tools[0].Function["name"], "echo")
	}

	messages := agent.Messages()
	if len(messages) != 3 {
		t.Fatalf("stored messages len = %d, want 3", len(messages))
	}
	if openai.MessageText(messages[2].Content) != "Hello world" {
		t.Fatalf("stored assistant message = %#v, want Hello world", messages[2])
	}
}

func TestRunLoopEmitsReasoningChunks(t *testing.T) {
	t.Parallel()

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Extra: map[string]any{"reasoning": "thinking"}}),
					streamEvent(0, openai.Message{Role: "assistant", Content: "done"}),
				},
			},
		},
	}

	agent := loop.New(streamer, nil)

	chunks := runLoop(t, agent, []openai.Message{openai.UserMessage("Hi")}, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeReasoning,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].Text != "thinking" {
		t.Fatalf("reasoning chunk = %#v, want thinking", chunks[0])
	}
	if chunks[1].Text != "done" {
		t.Fatalf("message chunk = %#v, want done", chunks[1])
	}
}

func TestRunLoopExecutesToolCallsAndContinuesConversation(t *testing.T) {
	t.Parallel()

	initialMessages := []openai.Message{
		openai.SystemMessage("system prompt"),
		openai.UserMessage("Hi"),
	}

	tool := &spyTool{
		description: "Echoes text",
		arguments:   []loop.ToolArgument{{Name: "value", Type: "string"}},
		result:      "tool-result",
	}
	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{
						Role: "assistant",
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: openai.ToolCallFunction{
									Name:      "echo",
									Arguments: `{"value":"hel`,
								},
							},
						},
					}),
					streamEvent(0, openai.Message{
						ToolCalls: []openai.ToolCall{
							{
								Function: openai.ToolCallFunction{
									Arguments: `lo"}`,
								},
							},
						},
					}),
				},
			},
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Role: "assistant", Content: "final answer"}),
				},
			},
		},
	}

	agent := loop.New(streamer, map[string]loop.Tool{"echo": tool})

	chunks := runLoop(t, agent, initialMessages, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeToolCallStart,
		loop.ChunkTypeToolCallResult,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].ToolName != "echo" || !reflect.DeepEqual(chunks[0].ToolArgs, map[string]any{"value": "hello"}) {
		t.Fatalf("tool start chunk = %#v", chunks[0])
	}
	if chunks[1].Text != "tool-result" {
		t.Fatalf("tool result chunk = %#v, want tool-result", chunks[1])
	}
	if chunks[2].Text != "final answer" {
		t.Fatalf("message chunk = %#v, want final answer", chunks[2])
	}

	if tool.callCount != 1 {
		t.Fatalf("tool call count = %d, want 1", tool.callCount)
	}
	if !reflect.DeepEqual(tool.lastArgs, map[string]any{"value": "hello"}) {
		t.Fatalf("tool args = %#v, want %#v", tool.lastArgs, map[string]any{"value": "hello"})
	}

	if len(streamer.calls) != 2 {
		t.Fatalf("stream calls = %d, want 2", len(streamer.calls))
	}
	secondReq := streamer.calls[1]
	if len(secondReq.Messages) != 4 {
		t.Fatalf("second request messages len = %d, want 4", len(secondReq.Messages))
	}
	if secondReq.Messages[2].Role != "assistant" || len(secondReq.Messages[2].ToolCalls) != 1 {
		t.Fatalf("second request assistant tool message = %#v", secondReq.Messages[2])
	}
	if secondReq.Messages[3].Role != "tool" || openai.MessageText(secondReq.Messages[3].Content) != "tool-result" {
		t.Fatalf("second request tool message = %#v", secondReq.Messages[3])
	}
}

func TestRunLoopPassesBackReasoningContentForToolContinuation(t *testing.T) {
	t.Parallel()

	tool := &spyTool{
		description: "Echoes text",
		arguments:   []loop.ToolArgument{{Name: "value", Type: "string"}},
		result:      "tool-result",
	}
	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{
						Role:  "assistant",
						Extra: map[string]any{"reasoning_content": "need "},
					}),
					streamEvent(0, openai.Message{
						Extra: map[string]any{"reasoning_content": "tool"},
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: openai.ToolCallFunction{
									Name:      "echo",
									Arguments: `{"value":"hello"}`,
								},
							},
						},
					}),
				},
			},
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Role: "assistant", Content: "final answer"}),
				},
			},
		},
	}

	agent := loop.New(streamer, map[string]loop.Tool{"echo": tool})

	chunks := runLoop(t, agent, []openai.Message{openai.UserMessage("Hi")}, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeReasoning,
		loop.ChunkTypeReasoning,
		loop.ChunkTypeToolCallStart,
		loop.ChunkTypeToolCallResult,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if len(streamer.calls) != 2 {
		t.Fatalf("stream calls = %d, want 2", len(streamer.calls))
	}

	assistantMessage := streamer.calls[1].Messages[1]
	if assistantMessage.Role != "assistant" || len(assistantMessage.ToolCalls) != 1 {
		t.Fatalf("second request assistant message = %#v", assistantMessage)
	}
	if assistantMessage.Content == nil || openai.MessageText(assistantMessage.Content) != "" {
		t.Fatalf("second request assistant content = %#v, want empty string", assistantMessage.Content)
	}
	if assistantMessage.Extra["reasoning_content"] != "need tool" {
		t.Fatalf("second request reasoning_content = %#v, want %q", assistantMessage.Extra["reasoning_content"], "need tool")
	}
}

func TestRunLoopReportsUnknownToolErrors(t *testing.T) {
	t.Parallel()

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{
						Role: "assistant",
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: openai.ToolCallFunction{
									Name:      "missing_tool",
									Arguments: `{"value":"hello"}`,
								},
							},
						},
					}),
				},
			},
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Role: "assistant", Content: "fallback"}),
				},
			},
		},
	}

	agent := loop.New(streamer, nil)

	chunks := runLoop(t, agent, []openai.Message{openai.UserMessage("Hi")}, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeToolCallError,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].ToolName != "missing_tool" || chunks[0].Text != "tool not found" {
		t.Fatalf("tool error chunk = %#v", chunks[0])
	}

	if len(streamer.calls) != 2 {
		t.Fatalf("stream calls = %d, want 2", len(streamer.calls))
	}
	secondReq := streamer.calls[1]
	lastMessage := secondReq.Messages[len(secondReq.Messages)-1]
	if lastMessage.Role != "tool" || openai.MessageText(lastMessage.Content) != "ERROR: tool not found" {
		t.Fatalf("second request last message = %#v", lastMessage)
	}
}

func TestRunLoopReportsInvalidToolArguments(t *testing.T) {
	t.Parallel()

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{
						Role: "assistant",
						ToolCalls: []openai.ToolCall{
							{
								ID:   "call_1",
								Type: "function",
								Function: openai.ToolCallFunction{
									Name:      "echo",
									Arguments: `{"value":`,
								},
							},
						},
					}),
				},
			},
			{
				events: []openai.StreamEvent{
					streamEvent(0, openai.Message{Role: "assistant", Content: "fallback"}),
				},
			},
		},
	}

	agent := loop.New(streamer, map[string]loop.Tool{
		"echo": loop.NewSimpleTool("Echoes text", []loop.ToolArgument{{Name: "value", Type: "string"}}, func(args map[string]any) string {
			return "unused"
		}),
	})

	chunks := runLoop(t, agent, []openai.Message{openai.UserMessage("Hi")}, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeToolCallError,
		loop.ChunkTypeMessage,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].ToolName != "echo" || chunks[0].Text == "" {
		t.Fatalf("tool error chunk = %#v", chunks[0])
	}
	if len(streamer.calls) != 2 {
		t.Fatalf("stream calls = %d, want 2", len(streamer.calls))
	}
	lastMessage := streamer.calls[1].Messages[len(streamer.calls[1].Messages)-1]
	if lastMessage.Role != "tool" || lastMessage.ToolCallID != "call_1" {
		t.Fatalf("second request last message = %#v", lastMessage)
	}
}

func TestRunLoopEmitsErrorWhenStreamSetupFails(t *testing.T) {
	t.Parallel()

	initialMessages := []openai.Message{
		openai.SystemMessage("system prompt"),
		openai.UserMessage("Hi"),
	}

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{err: errors.New("boom")},
		},
	}

	agent := loop.New(streamer, nil)

	chunks := runLoop(t, agent, initialMessages, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeError,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].Text != "error in LLM stream: boom" {
		t.Fatalf("error chunk = %#v", chunks[0])
	}

	messages := agent.Messages()
	if len(messages) != 2 {
		t.Fatalf("stored messages len = %d, want 2", len(messages))
	}
	if messages[0].Role != "system" || messages[1].Role != "user" {
		t.Fatalf("stored messages = %#v", messages)
	}
}

func TestRunLoopEmitsErrorWhenStreamEventFails(t *testing.T) {
	t.Parallel()

	initialMessages := []openai.Message{
		openai.SystemMessage("system prompt"),
		openai.UserMessage("Hi"),
	}

	streamer := &scriptedStreamer{
		scripts: []streamScript{
			{
				events: []openai.StreamEvent{
					{Error: errors.New("stream broke")},
				},
			},
		},
	}

	agent := loop.New(streamer, nil)

	chunks := runLoop(t, agent, initialMessages, "test-model")

	if got := chunkTypes(chunks); !reflect.DeepEqual(got, []loop.ChunkType{
		loop.ChunkTypeError,
		loop.ChunkTypeEnd,
	}) {
		t.Fatalf("chunk types = %#v", got)
	}
	if chunks[0].Text != "error in LLM stream: stream broke" {
		t.Fatalf("error chunk = %#v", chunks[0])
	}

	if len(agent.Messages()) != 2 {
		t.Fatalf("stored messages len = %d, want 2", len(agent.Messages()))
	}
}

type streamScript struct {
	events []openai.StreamEvent
	err    error
}

type scriptedStreamer struct {
	mu      sync.Mutex
	calls   []openai.ChatCompletionRequest
	scripts []streamScript
}

func (s *scriptedStreamer) StreamChatCompletion(_ context.Context, req openai.ChatCompletionRequest) (<-chan openai.StreamEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, req)
	if len(s.scripts) == 0 {
		return nil, errors.New("unexpected StreamChatCompletion call")
	}

	script := s.scripts[0]
	s.scripts = s.scripts[1:]
	if script.err != nil {
		return nil, script.err
	}

	ch := make(chan openai.StreamEvent, len(script.events))
	for _, event := range script.events {
		ch <- event
	}
	close(ch)
	return ch, nil
}

type spyTool struct {
	description string
	arguments   []loop.ToolArgument
	result      string
	callCount   int
	lastArgs    map[string]any
}

func (t *spyTool) Description() string {
	return t.description
}

func (t *spyTool) Arguments() []loop.ToolArgument {
	return t.arguments
}

func (t *spyTool) Call(args map[string]any) string {
	t.callCount++
	t.lastArgs = args
	return t.result
}

func streamEvent(index int, delta openai.Message) openai.StreamEvent {
	return openai.StreamEvent{
		Data: &openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Index: index,
					Delta: &delta,
				},
			},
		},
	}
}

func runLoop(t *testing.T, agent *loop.Loop, messages []openai.Message, model string) []loop.Chunk {
	t.Helper()

	done := make(chan struct{})
	go func() {
		agent.RunLoop(context.Background(), messages, model)
		close(done)
	}()

	var chunks []loop.Chunk
	for {
		select {
		case chunk := <-agent.Chunks():
			chunks = append(chunks, chunk)
			if chunk.Type == loop.ChunkTypeEnd {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("timed out waiting for RunLoop to return")
				}
				return chunks
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for chunk")
		}
	}
}

func chunkTypes(chunks []loop.Chunk) []loop.ChunkType {
	types := make([]loop.ChunkType, 0, len(chunks))
	for _, chunk := range chunks {
		types = append(types, chunk.Type)
	}
	return types
}
