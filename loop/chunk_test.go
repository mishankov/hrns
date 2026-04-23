package loop_test

import (
	"reflect"
	"testing"

	"github.com/mishankov/hrns/loop"
)

func TestChunkConstructors(t *testing.T) {
	t.Parallel()

	if got := loop.NewChunkMessage("hello"); !reflect.DeepEqual(got, loop.Chunk{
		Type: loop.ChunkTypeMessage,
		Text: "hello",
	}) {
		t.Fatalf("NewChunkMessage() = %#v", got)
	}

	if got := loop.NewChunkReasoning("thinking"); !reflect.DeepEqual(got, loop.Chunk{
		Type: loop.ChunkTypeReasoning,
		Text: "thinking",
	}) {
		t.Fatalf("NewChunkReasoning() = %#v", got)
	}

	if got := loop.NewChunkError("boom"); !reflect.DeepEqual(got, loop.Chunk{
		Type: loop.ChunkTypeError,
		Text: "boom",
	}) {
		t.Fatalf("NewChunkError() = %#v", got)
	}

	args := map[string]any{"name": "demo"}
	if got := loop.NewChunkToolCallStart("load_skill", args); !reflect.DeepEqual(got, loop.Chunk{
		Type:     loop.ChunkTypeToolCallStart,
		ToolName: "load_skill",
		ToolArgs: args,
	}) {
		t.Fatalf("NewChunkToolCallStart() = %#v", got)
	}

	if got := loop.NewChunkToolCallError("load_skill", "missing"); !reflect.DeepEqual(got, loop.Chunk{
		Type:     loop.ChunkTypeToolCallError,
		ToolName: "load_skill",
		Text:     "missing",
	}) {
		t.Fatalf("NewChunkToolCallError() = %#v", got)
	}

	if got := loop.NewChunkToolCallResult("load_skill", "body"); !reflect.DeepEqual(got, loop.Chunk{
		Type:     loop.ChunkTypeToolCallResult,
		ToolName: "load_skill",
		Text:     "body",
	}) {
		t.Fatalf("NewChunkToolCallResult() = %#v", got)
	}

	if got := loop.NewChunkEnd(); !reflect.DeepEqual(got, loop.Chunk{
		Type: loop.ChunkTypeEnd,
	}) {
		t.Fatalf("NewChunkEnd() = %#v", got)
	}
}

func TestChunkIsToolChunk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		chunk loop.Chunk
		want  bool
	}{
		{
			name:  "message",
			chunk: loop.NewChunkMessage("hello"),
			want:  false,
		},
		{
			name:  "reasoning",
			chunk: loop.NewChunkReasoning("thinking"),
			want:  false,
		},
		{
			name:  "error",
			chunk: loop.NewChunkError("boom"),
			want:  false,
		},
		{
			name:  "tool call start",
			chunk: loop.NewChunkToolCallStart("load_skill", map[string]any{}),
			want:  true,
		},
		{
			name:  "tool call error",
			chunk: loop.NewChunkToolCallError("load_skill", "missing"),
			want:  true,
		},
		{
			name:  "tool call result",
			chunk: loop.NewChunkToolCallResult("load_skill", "body"),
			want:  true,
		},
		{
			name:  "end",
			chunk: loop.NewChunkEnd(),
			want:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.chunk.IsToolChunk(); got != tt.want {
				t.Fatalf("IsToolChunk() = %v, want %v", got, tt.want)
			}
		})
	}
}
