package agent

type ChunkType string

const (
	ChunkTypeMessage        ChunkType = "message"
	ChunkTypeReasoning      ChunkType = "reasoning"
	ChunkTypeError          ChunkType = "error"
	ChunkTypeToolCallStart  ChunkType = "tool_call_start"
	ChunkTypeToolCallError  ChunkType = "tool_call_error"
	ChunkTypeToolCallResult ChunkType = "tool_call_result"
	ChunkTypeEnd            ChunkType = "end"
)

type Chunk struct {
	Type     ChunkType
	Text     string
	ToolName string
	ToolArgs map[string]any
}

func NewChunkMessage(text string) Chunk {
	return Chunk{
		Type: ChunkTypeMessage,
		Text: text,
	}
}

func NewChunkReasoning(text string) Chunk {
	return Chunk{
		Type: ChunkTypeReasoning,
		Text: text,
	}
}

func NewChunkError(text string) Chunk {
	return Chunk{
		Type: ChunkTypeError,
		Text: text,
	}
}

func NewChunkToolCallStart(name string, args map[string]any) Chunk {
	return Chunk{
		Type:     ChunkTypeToolCallStart,
		ToolName: name,
		ToolArgs: args,
	}
}

func NewChunkToolCallError(name string, error string) Chunk {
	return Chunk{
		Type:     ChunkTypeToolCallError,
		ToolName: name,
		Text:     error,
	}
}

func NewChunkToolCallResult(name string, result string) Chunk {
	return Chunk{
		Type:     ChunkTypeToolCallResult,
		ToolName: name,
		Text:     result,
	}
}

func NewChunkEnd() Chunk {
	return Chunk{
		Type: ChunkTypeEnd,
	}
}
