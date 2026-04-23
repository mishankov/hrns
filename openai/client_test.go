package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/mishankov/hrns/openai"
)

func TestClientNewClientDefaults(t *testing.T) {
	t.Parallel()

	client := openai.NewClient()

	if client.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL, "https://api.openai.com/v1")
	}
	if client.APIKey != "" {
		t.Fatalf("APIKey = %q, want empty", client.APIKey)
	}
	if client.HTTPClient == nil {
		t.Fatal("HTTPClient = nil, want non-nil")
	}
}

func TestClientListModelsUsesConfiguredBaseURLAndAPIKey(t *testing.T) {
	t.Parallel()

	type recordedRequest struct {
		Method        string
		Path          string
		Authorization string
	}

	got := make(chan recordedRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		got <- recordedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"object":"list",
			"data":[
				{"id":"gpt-4.1","object":"model"},
				{"id":"","object":"model"},
				{"id":"gpt-4.1-mini","object":"model"}
			]
		}`))
	}))
	defer server.Close()

	client := openai.NewClient(
		openai.WithBaseURL(server.URL+"/"),
		openai.WithAPIKey("secret"),
	)

	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels returned error: %v", err)
	}

	request := <-got
	if request.Method != http.MethodGet {
		t.Fatalf("method = %q, want %q", request.Method, http.MethodGet)
	}
	if request.Path != "/models" {
		t.Fatalf("path = %q, want %q", request.Path, "/models")
	}
	if request.Authorization != "Bearer secret" {
		t.Fatalf("authorization = %q, want %q", request.Authorization, "Bearer secret")
	}

	want := []string{"gpt-4.1", "gpt-4.1-mini"}
	if !slices.Equal(models, want) {
		t.Fatalf("models = %#v, want %#v", models, want)
	}
}

func TestClientListModelsReturnsAPIErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"models endpoint unavailable"}}`))
	}))
	defer server.Close()

	client := openai.NewClient(openai.WithBaseURL(server.URL))

	_, err := client.ListModels(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *openai.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *openai.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Message != "models endpoint unavailable" {
		t.Fatalf("message = %q, want %q", apiErr.Message, "models endpoint unavailable")
	}
}

func TestClientCreateChatCompletionUsesConfiguredBaseURLAndAPIKey(t *testing.T) {
	t.Parallel()

	type recordedRequest struct {
		Method        string
		Path          string
		Authorization string
		ContentType   string
		Model         string
		MessageCount  int
	}

	got := make(chan recordedRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var body struct {
			Model    string            `json:"model"`
			Messages []json.RawMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		got <- recordedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
			ContentType:   r.Header.Get("Content-Type"),
			Model:         body.Model,
			MessageCount:  len(body.Messages),
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"resp_123",
			"model":"test-model",
			"choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}]
		}`))
	}))
	defer server.Close()

	client := openai.NewClient(
		openai.WithBaseURL(server.URL+"/"),
		openai.WithAPIKey("secret"),
	)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "test-model",
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion returned error: %v", err)
	}

	request := <-got
	if request.Method != http.MethodPost {
		t.Fatalf("method = %q, want %q", request.Method, http.MethodPost)
	}
	if request.Path != "/chat/completions" {
		t.Fatalf("path = %q, want %q", request.Path, "/chat/completions")
	}
	if request.Authorization != "Bearer secret" {
		t.Fatalf("authorization = %q, want %q", request.Authorization, "Bearer secret")
	}
	if request.ContentType != "application/json" {
		t.Fatalf("content type = %q, want %q", request.ContentType, "application/json")
	}
	if request.Model != "test-model" {
		t.Fatalf("request model = %q, want %q", request.Model, "test-model")
	}
	if request.MessageCount != 1 {
		t.Fatalf("message count = %d, want 1", request.MessageCount)
	}

	if resp.ID != "resp_123" {
		t.Fatalf("response id = %q, want %q", resp.ID, "resp_123")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("choices len = %d, want 1", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Fatalf("response role = %q, want %q", resp.Choices[0].Message.Role, "assistant")
	}
	if openai.MessageText(resp.Choices[0].Message.Content) != "hello" {
		t.Fatalf("response content = %q, want %q", openai.MessageText(resp.Choices[0].Message.Content), "hello")
	}
}

func TestClientCreateChatCompletionReturnsValidationAndAPIErrors(t *testing.T) {
	t.Parallel()

	client := openai.NewClient()

	_, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err == nil || err.Error() != "model is required" {
		t.Fatalf("error = %v, want model is required", err)
	}

	_, err = client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "test-model",
	})
	if err == nil || err.Error() != "messages is required" {
		t.Fatalf("error = %v, want messages is required", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client = openai.NewClient(openai.WithBaseURL(server.URL))

	_, err = client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "test-model",
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *openai.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *openai.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
	if apiErr.Message != "bad key" {
		t.Fatalf("message = %q, want %q", apiErr.Message, "bad key")
	}
}

func TestClientStreamChatCompletionStreamsEvents(t *testing.T) {
	t.Parallel()

	type recordedStreamRequest struct {
		Accept       string
		Stream       bool
		MessageCount int
	}

	got := make(chan recordedStreamRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var body struct {
			Stream   bool              `json:"stream"`
			Messages []json.RawMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		got <- recordedStreamRequest{
			Accept:       r.Header.Get("Accept"),
			Stream:       body.Stream,
			MessageCount: len(body.Messages),
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not implement http.Flusher")
			return
		}

		_, _ = w.Write([]byte("data: " + `{"choices":[{"index":0,"delta":{"role":"assistant","content":"Hel"}}]}` + "\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("data: " + `{"choices":[{"index":0,"delta":{"content":"lo"}}]}` + "\n\n"))
		flusher.Flush()
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := openai.NewClient(openai.WithBaseURL(server.URL))

	stream, err := client.StreamChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "test-model",
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err != nil {
		t.Fatalf("StreamChatCompletion returned error: %v", err)
	}

	request := <-got
	if request.Accept != "text/event-stream" {
		t.Fatalf("accept = %q, want %q", request.Accept, "text/event-stream")
	}
	if !request.Stream {
		t.Fatal("stream flag = false, want true")
	}
	if request.MessageCount != 1 {
		t.Fatalf("message count = %d, want 1", request.MessageCount)
	}

	var events []openai.StreamEvent
	for event := range stream {
		events = append(events, event)
	}

	if len(events) != 3 {
		t.Fatalf("events len = %d, want 3", len(events))
	}
	if events[0].Error != nil {
		t.Fatalf("first event error = %v, want nil", events[0].Error)
	}
	if events[0].Done {
		t.Fatal("first event Done = true, want false")
	}
	if events[0].Data == nil || len(events[0].Data.Choices) != 1 {
		t.Fatalf("first event data = %#v, want one choice", events[0].Data)
	}
	if openai.MessageText(events[0].Data.Choices[0].Delta.Content) != "Hel" {
		t.Fatalf("first chunk content = %q, want %q", openai.MessageText(events[0].Data.Choices[0].Delta.Content), "Hel")
	}
	if openai.MessageText(events[1].Data.Choices[0].Delta.Content) != "lo" {
		t.Fatalf("second chunk content = %q, want %q", openai.MessageText(events[1].Data.Choices[0].Delta.Content), "lo")
	}
	if !events[2].Done {
		t.Fatal("final event Done = false, want true")
	}
	if string(events[2].Raw) != "[DONE]" {
		t.Fatalf("final raw payload = %q, want %q", string(events[2].Raw), "[DONE]")
	}
}

func TestClientStreamChatCompletionReturnsEventAndAPIErrors(t *testing.T) {
	t.Parallel()

	malformedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("response writer does not implement http.Flusher")
			return
		}

		_, _ = w.Write([]byte("data: " + `{"choices":[` + "\n\n"))
		flusher.Flush()
	}))
	defer malformedServer.Close()

	client := openai.NewClient(openai.WithBaseURL(malformedServer.URL))

	stream, err := client.StreamChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "test-model",
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err != nil {
		t.Fatalf("StreamChatCompletion returned error: %v", err)
	}

	var events []openai.StreamEvent
	for event := range stream {
		events = append(events, event)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].Error == nil {
		t.Fatal("event error = nil, want non-nil")
	}
	if events[0].Done {
		t.Fatal("event Done = true, want false")
	}

	apiErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream failed"}}`))
	}))
	defer apiErrorServer.Close()

	client = openai.NewClient(openai.WithBaseURL(apiErrorServer.URL))

	_, err = client.StreamChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    "test-model",
		Messages: []openai.Message{openai.UserMessage("hi")},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *openai.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *openai.APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status code = %d, want %d", apiErr.StatusCode, http.StatusBadGateway)
	}
	if apiErr.Message != "upstream failed" {
		t.Fatalf("message = %q, want %q", apiErr.Message, "upstream failed")
	}
}
