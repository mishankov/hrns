package openai

import "fmt"

type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("chat completion API error: status=%d message=%s",
		e.StatusCode, e.Message)
}
