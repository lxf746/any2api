package chatgpt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"any2api-go/internal/core"
)

func TestChatGPTProviderUsesBearerTokenAndOpenAIPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer chatgpt-token" {
			t.Fatalf("expected bearer token, got %q", got)
		}
		var body chatgptChatRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.Model != "gpt-4.1" {
			t.Fatalf("unexpected model: %q", body.Model)
		}
		if len(body.Messages) != 2 {
			t.Fatalf("expected injected system + latest user message, got %#v", body.Messages)
		}
		if body.Messages[0].Role != "system" || core.ContentText(body.Messages[0].Content) != "keep it concise" {
			t.Fatalf("unexpected first message: %#v", body.Messages[0])
		}
		if body.Messages[1].Role != "user" || core.ContentText(body.Messages[1].Content) != "hi" {
			t.Fatalf("unexpected truncated user message: %#v", body.Messages[1])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello chatgpt"}}]}`))
	}))
	defer upstream.Close()

	p := NewProviderWithConfig(core.ChatGPTConfig{
		BaseURL: upstream.URL,
		Token:   "chatgpt-token",
		Request: core.RequestConfig{Timeout: 5 * time.Second, MaxInputLength: 2, SystemPromptInject: "keep it concise"},
	}).(*chatgptProvider)

	text, err := p.CompleteOpenAI(context.Background(), core.UnifiedRequest{
		Messages: []core.Message{
			{Role: "user", Content: "this is too long"},
			{Role: "user", Content: "hi"},
		},
	})
	if err != nil {
		t.Fatalf("CompleteOpenAI returned error: %v", err)
	}
	if text != "hello chatgpt" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestChatGPTProviderStreamsJSONFallbackResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello chatgpt"}}]}`))
	}))
	defer upstream.Close()

	p := NewProviderWithConfig(core.ChatGPTConfig{
		BaseURL: upstream.URL,
		Token:   "chatgpt-token",
		Request: core.RequestConfig{Timeout: 5 * time.Second, MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*chatgptProvider)

	events, err := p.StreamOpenAI(context.Background(), core.UnifiedRequest{
		Model:    "gpt-4.1",
		Messages: []core.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("StreamOpenAI returned error: %v", err)
	}
	text, err := core.CollectTextStream(context.Background(), events)
	if err != nil {
		t.Fatalf("CollectTextStream returned error: %v", err)
	}
	if text != "hello chatgpt" {
		t.Fatalf("unexpected stream text: %q", text)
	}
}
