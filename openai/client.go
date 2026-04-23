package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Headers    map[string]string
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.APIKey = apiKey
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		BaseURL: "https://api.openai.com/v1",
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.BaseURL+"/models",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    extractErrorMessage(respBody),
			Body:       respBody,
		}
	}

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	models := make([]string, 0, len(out.Data))
	for _, model := range out.Data {
		if model.ID != "" {
			models = append(models, model.ID)
		}
	}
	return models, nil
}

func (c *Client) CreateChatCompletion(
	ctx context.Context,
	reqBody ChatCompletionRequest,
) (*ChatCompletionResponse, error) {
	if reqBody.Model == "" {
		return nil, errors.New("model is required")
	}
	if len(reqBody.Messages) == 0 {
		return nil, errors.New("messages is required")
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.applyHeaders(req)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    extractErrorMessage(respBody),
			Body:       respBody,
		}
	}

	var out ChatCompletionResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &out, nil
}

type StreamEvent struct {
	Data  *ChatCompletionResponse
	Done  bool
	Raw   []byte
	Error error
}

func (c *Client) StreamChatCompletion(
	ctx context.Context,
	reqBody ChatCompletionRequest,
) (<-chan StreamEvent, error) {
	if reqBody.Model == "" {
		return nil, errors.New("model is required")
	}
	if len(reqBody.Messages) == 0 {
		return nil, errors.New("messages is required")
	}

	reqBody.Stream = true

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.applyHeaders(req)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    extractErrorMessage(respBody),
			Body:       respBody,
		}
	}

	ch := make(chan StreamEvent)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		var dataLines []string

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				if !emitSSEEvent(ch, dataLines) {
					return
				}
				dataLines = dataLines[:0]
				continue
			}

			if strings.HasPrefix(line, "data:") {
				payload := strings.TrimPrefix(line, "data:")
				if len(payload) > 0 && payload[0] == ' ' {
					payload = payload[1:]
				}
				dataLines = append(dataLines, payload)
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- StreamEvent{Error: fmt.Errorf("stream scan: %w", err)}
			return
		}

		emitSSEEvent(ch, dataLines)
	}()

	return ch, nil
}

func emitSSEEvent(ch chan<- StreamEvent, dataLines []string) bool {
	if len(dataLines) == 0 {
		return true
	}

	payload := strings.Join(dataLines, "\n")
	if payload == "[DONE]" {
		ch <- StreamEvent{Done: true, Raw: []byte(payload)}
		return false
	}

	var out ChatCompletionResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		ch <- StreamEvent{
			Raw:   []byte(payload),
			Error: fmt.Errorf("unmarshal stream event: %w", err),
		}
		return false
	}

	ch <- StreamEvent{
		Data: &out,
		Raw:  []byte(payload),
	}
	return true
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func extractErrorMessage(body []byte) string {
	var v map[string]any
	if err := json.Unmarshal(body, &v); err != nil {
		return string(body)
	}

	if errObj, ok := v["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
	}

	if msg, ok := v["message"].(string); ok {
		return msg
	}

	return string(body)
}
