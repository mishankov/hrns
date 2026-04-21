package openai

import "encoding/json"

type Tool struct {
	Type     string         `json:"type"`
	Function map[string]any `json:"function,omitempty"`
	Extra    map[string]any `json:"-"`
}

func (t Tool) MarshalJSON() ([]byte, error) {
	type alias Tool
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
