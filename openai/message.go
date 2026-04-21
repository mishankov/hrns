package openai

import "encoding/json"

type Message struct {
	Role       string     `json:"role"`
	Content    any        `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Refusal    string     `json:"refusal,omitempty"`

	Extra map[string]any `json:"-"`
}

func (m Message) MarshalJSON() ([]byte, error) {
	type alias Message
	base := map[string]any{}
	raw, err := json.Marshal(alias(m))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, err
	}
	for k, v := range m.Extra {
		base[k] = v
	}
	return json.Marshal(base)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type alias Message
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = Message(a)

	var all map[string]any
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	known := map[string]struct{}{
		"role":         {},
		"content":      {},
		"name":         {},
		"tool_calls":   {},
		"tool_call_id": {},
		"refusal":      {},
	}
	extra := make(map[string]any)
	for k, v := range all {
		if _, ok := known[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		m.Extra = extra
	}
	return nil
}

func SystemMessage(content string) Message {
	return Message{
		Role:    "system",
		Content: content,
	}
}

func UserMessage(content string) Message {
	return Message{
		Role:    "user",
		Content: content,
	}
}

func ToolMessage(content, toolCallID string) Message {
	return Message{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolCallID,
	}
}
