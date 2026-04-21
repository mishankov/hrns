package openai

type ChatCompletionAccumulator struct {
	choices map[int]*ChatCompletionChoice
}

func (a *ChatCompletionAccumulator) AddChunk(chunk ChatCompletionResponse) {
	if a.choices == nil {
		a.choices = map[int]*ChatCompletionChoice{}
	}

	for i := range chunk.Choices {
		a.accumulateChoice(chunk.Choices[i])
	}
}

func (a *ChatCompletionAccumulator) Choices() []ChatCompletionChoice {
	return orderedChoices(a.choices)
}

func (a *ChatCompletionAccumulator) accumulateChoice(incoming ChatCompletionChoice) {
	choice, ok := a.choices[incoming.Index]
	if !ok {
		choice = &ChatCompletionChoice{Index: incoming.Index}
		a.choices[incoming.Index] = choice
	}

	if incoming.FinishReason != "" {
		choice.FinishReason = incoming.FinishReason
	}
	if incoming.Logprobs != nil {
		choice.Logprobs = incoming.Logprobs
	}
	if len(incoming.Extra) > 0 {
		if choice.Extra == nil {
			choice.Extra = map[string]any{}
		}
		for k, v := range incoming.Extra {
			choice.Extra[k] = v
		}
	}

	if incoming.Delta == nil {
		if incoming.Message.Role != "" || incoming.Message.Content != nil || len(incoming.Message.ToolCalls) > 0 {
			mergeMessage(&choice.Message, incoming.Message)
		}
		return
	}

	mergeMessage(&choice.Message, *incoming.Delta)
}

func mergeMessage(dst *Message, delta Message) {
	if delta.Role != "" {
		dst.Role = delta.Role
	}

	if content := MessageText(delta.Content); content != "" {
		dst.Content = MessageText(dst.Content) + content
	} else if dst.Content == nil && delta.Content != nil {
		dst.Content = delta.Content
	}

	if delta.Name != "" {
		dst.Name = delta.Name
	}
	if delta.ToolCallID != "" {
		dst.ToolCallID = delta.ToolCallID
	}
	if delta.Refusal != "" {
		dst.Refusal = delta.Refusal
	}
	if len(delta.Extra) > 0 {
		if dst.Extra == nil {
			dst.Extra = map[string]any{}
		}
		for k, v := range delta.Extra {
			dst.Extra[k] = v
		}
	}

	for i, toolCallDelta := range delta.ToolCalls {
		if len(dst.ToolCalls) <= i {
			dst.ToolCalls = append(dst.ToolCalls, ToolCall{})
		}
		mergeToolCall(&dst.ToolCalls[i], toolCallDelta)
	}
}

func mergeToolCall(dst *ToolCall, delta ToolCall) {
	if delta.ID != "" {
		dst.ID = delta.ID
	}
	if delta.Type != "" {
		dst.Type = delta.Type
	}
	if delta.Function.Name != "" {
		dst.Function.Name = delta.Function.Name
	}
	if delta.Function.Arguments != "" {
		dst.Function.Arguments += delta.Function.Arguments
	}
	if len(delta.Extra) > 0 {
		if dst.Extra == nil {
			dst.Extra = map[string]any{}
		}
		for k, v := range delta.Extra {
			dst.Extra[k] = v
		}
	}
	if len(delta.Function.Extra) > 0 {
		if dst.Function.Extra == nil {
			dst.Function.Extra = map[string]any{}
		}
		for k, v := range delta.Function.Extra {
			dst.Function.Extra[k] = v
		}
	}
}

func orderedChoices(acc map[int]*ChatCompletionChoice) []ChatCompletionChoice {
	if len(acc) == 0 {
		return nil
	}

	maxIndex := -1
	for index := range acc {
		if index > maxIndex {
			maxIndex = index
		}
	}

	choices := make([]ChatCompletionChoice, 0, len(acc))
	for i := 0; i <= maxIndex; i++ {
		if choice, ok := acc[i]; ok {
			choices = append(choices, *choice)
		}
	}

	return choices
}

func MessageText(content any) string {
	text, _ := content.(string)
	return text
}
