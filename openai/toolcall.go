package openai

import "encoding/json"

type ToolCall struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function ToolCallFunction `json:"function,omitempty"`
	Extra    map[string]any   `json:"-"`
}

func (t ToolCall) MarshalJSON() ([]byte, error) {
	type alias ToolCall
	base := map[string]any{}
	raw, err := json.Marshal(alias(t))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, err
	}
	for k, v := range t.Extra {
		base[k] = v
	}
	return json.Marshal(base)
}

func (t *ToolCall) UnmarshalJSON(data []byte) error {
	type alias ToolCall
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*t = ToolCall(a)

	var all map[string]any
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	known := map[string]struct{}{
		"id":       {},
		"type":     {},
		"function": {},
	}
	extra := make(map[string]any)
	for k, v := range all {
		if _, ok := known[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		t.Extra = extra
	}
	return nil
}

type ToolCallFunction struct {
	Name      string         `json:"name,omitempty"`
	Arguments string         `json:"arguments,omitempty"`
	Extra     map[string]any `json:"-"`
}

func (f ToolCallFunction) MarshalJSON() ([]byte, error) {
	type alias ToolCallFunction
	base := map[string]any{}
	raw, err := json.Marshal(alias(f))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, err
	}
	for k, v := range f.Extra {
		base[k] = v
	}
	return json.Marshal(base)
}

func (f *ToolCallFunction) UnmarshalJSON(data []byte) error {
	type alias ToolCallFunction
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*f = ToolCallFunction(a)

	var all map[string]any
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	known := map[string]struct{}{
		"name":      {},
		"arguments": {},
	}
	extra := make(map[string]any)
	for k, v := range all {
		if _, ok := known[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		f.Extra = extra
	}
	return nil
}
