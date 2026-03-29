package zai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"any2api-go/internal/core"
)

func TestZAIProviderFetchesAnonymousTokenAndUsesChatEndpoint(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auths/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"anon-zai-token"}`))
		case "/api/chat/completions":
			if got := r.Header.Get("Authorization"); got != "Bearer anon-zai-token" {
				t.Fatalf("expected bearer token, got %q", got)
			}
			if got := r.Header.Get("X-FE-Version"); got != "0.9.1" {
				t.Fatalf("expected x-fe-version, got %q", got)
			}
			var body zaiChatRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode upstream request: %v", err)
			}
			if body.Model != "glm-4.5-flash" {
				t.Fatalf("unexpected model: %q", body.Model)
			}
			if len(body.Messages) != 2 {
				t.Fatalf("expected injected system + latest user message, got %#v", body.Messages)
			}
			if body.Messages[0].Role != "system" || core.ContentText(body.Messages[0].Content) != "keep it concise" {
				t.Fatalf("unexpected first message: %#v", body.Messages[0])
			}
			if body.Messages[1].Role != "user" || core.ContentText(body.Messages[1].Content) != "hi" {
				t.Fatalf("unexpected second message: %#v", body.Messages[1])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello zai"}}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	p := NewProviderWithConfig(core.ZAIConfig{
		BaseURL:   upstream.URL,
		FEVersion: "0.9.1",
		Request:   core.RequestConfig{Timeout: 5 * time.Second, MaxInputLength: 2, SystemPromptInject: "keep it concise"},
	}).(*zaiProvider)

	text, err := p.CompleteOpenAI(context.Background(), core.UnifiedRequest{
		Messages: []core.Message{{Role: "user", Content: "this is too long"}, {Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("CompleteOpenAI returned error: %v", err)
	}
	if text != "hello zai" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestZAIProviderStreamsSSEContent(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auths/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"anon-zai-token"}`))
		case "/api/chat/completions":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"message.delta\",\"data\":{\"delta_content\":\"hello\",\"done\":false}}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"message.delta\",\"data\":{\"delta_content\":\" zai\",\"done\":false}}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"message.delta\",\"data\":{\"delta_content\":\"\",\"done\":true}}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	p := NewProviderWithConfig(core.ZAIConfig{
		BaseURL: upstream.URL,
		Request: core.RequestConfig{Timeout: 5 * time.Second, MaxInputLength: core.DefaultCursorMaxInputLength},
	}).(*zaiProvider)

	events, err := p.StreamOpenAI(context.Background(), core.UnifiedRequest{
		Model:    "glm-4.5-flash",
		Messages: []core.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("StreamOpenAI returned error: %v", err)
	}
	text, err := core.CollectTextStream(context.Background(), events)
	if err != nil {
		t.Fatalf("CollectTextStream returned error: %v", err)
	}
	if text != "hello zai" {
		t.Fatalf("unexpected stream text: %q", text)
	}
}
