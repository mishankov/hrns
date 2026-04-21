package openai

import "encoding/json"

type ChatCompletionRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	N                *int      `json:"n,omitempty"`
	Stop             any       `json:"stop,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	User             string    `json:"user,omitempty"`
	Tools            []Tool    `json:"tools,omitempty"`
	ToolChoice       any       `json:"tool_choice,omitempty"`
	ResponseFormat   any       `json:"response_format,omitempty"`
	Logprobs         *bool     `json:"logprobs,omitempty"`
	TopLogprobs      *int      `json:"top_logprobs,omitempty"`

	Extra map[string]any `json:"-"`
}

func (r ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	type alias ChatCompletionRequest
	base := map[string]any{}
	raw, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, err
	}
	for k, v := range r.Extra {
		base[k] = v
	}
	return json.Marshal(base)
}

type ChatCompletionResponse struct {
	ID                string                     `json:"id,omitempty"`
	Object            string                     `json:"object,omitempty"`
	Created           int64                      `json:"created,omitempty"`
	Model             string                     `json:"model,omitempty"`
	Choices           []ChatCompletionChoice     `json:"choices,omitempty"`
	Usage             map[string]any             `json:"usage,omitempty"`
	SystemFingerprint string                     `json:"system_fingerprint,omitempty"`
	Extra             map[string]any             `json:"-"`
	Raw               map[string]json.RawMessage `json:"-"`
}

func (r *ChatCompletionResponse) UnmarshalJSON(data []byte) error {
	type alias ChatCompletionResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = ChatCompletionResponse(a)

	if err := json.Unmarshal(data, &r.Raw); err != nil {
		return err
	}

	var all map[string]any
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	known := map[string]struct{}{
		"id":                 {},
		"object":             {},
		"created":            {},
		"model":              {},
		"choices":            {},
		"usage":              {},
		"system_fingerprint": {},
	}
	extra := make(map[string]any)
	for k, v := range all {
		if _, ok := known[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		r.Extra = extra
	}
	return nil
}

type ChatCompletionChoice struct {
	Index        int            `json:"index,omitempty"`
	Message      Message        `json:"message,omitempty"`
	Delta        *Message       `json:"delta,omitempty"`
	FinishReason string         `json:"finish_reason,omitempty"`
	Logprobs     any            `json:"logprobs,omitempty"`
	Extra        map[string]any `json:"-"`
}

func (c *ChatCompletionChoice) UnmarshalJSON(data []byte) error {
	type alias ChatCompletionChoice
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = ChatCompletionChoice(a)

	var all map[string]any
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	known := map[string]struct{}{
		"index":         {},
		"message":       {},
		"delta":         {},
		"finish_reason": {},
		"logprobs":      {},
	}
	extra := make(map[string]any)
	for k, v := range all {
		if _, ok := known[k]; !ok {
			extra[k] = v
		}
	}
	if len(extra) > 0 {
		c.Extra = extra
	}
	return nil
}
