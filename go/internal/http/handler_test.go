package http

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"any2api-go/internal/core"
	"any2api-go/internal/platforms"
)

func testAppConfig() core.AppConfig {
	cfg := core.LoadAppConfigFromEnv()
	cfg.APIKey = "test-key"
	cfg.DefaultProvider = "cursor"
	return cfg
}

func TestModelsEndpointIncludesSevenProviders(t *testing.T) {
	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var body struct {
		Object string                   `json:"object"`
		Data   []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Object != "list" {
		t.Fatalf("expected object=list, got %q", body.Object)
	}
	providers := map[string]bool{}
	for _, item := range body.Data {
		provider, _ := item["provider"].(string)
		if provider != "" {
			providers[provider] = true
		}
	}
	for _, provider := range []string{"cursor", "kiro", "grok", "orchids", "web", "chatgpt", "blink"} {
		if !providers[provider] {
			t.Fatalf("expected models response to include provider %q, got %#v", provider, providers)
		}
	}
}

func TestOpenAIChatUsesRequestedProvider(t *testing.T) {
	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"kiro","model":"claude-sonnet-4.6","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)
	if rec.Header().Get("X-Newplatform2API-Provider") != "kiro" {
		t.Fatalf("expected provider header to be kiro, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
}

func TestOpenAIChatRejectsProviderWithoutOpenAICompatibility(t *testing.T) {
	cfg := testAppConfig()
	registry := core.NewRegistry("anthropic-only")
	registry.Register(testAnthropicOnlyProvider{})
	h := NewHandler(registry, cfg)
	payload := []byte(`{"provider":"anthropic-only","model":"stub-model","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["error"] != `provider "anthropic-only" does not support OpenAI compatible endpoint` {
		t.Fatalf("unexpected error body: %#v", body)
	}
}

func TestOpenAIChatOrchidsUsesConfiguredUpstream(t *testing.T) {
	clerk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/client":
			cookie, err := r.Cookie("__client")
			if err != nil || cookie.Value != "orchids-client" {
				t.Fatalf("expected orchids client cookie on clerk bootstrap, got %v %v", cookie, err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"response":{"last_active_session_id":"sess_test","sessions":[{"user":{"id":"user_test","email_addresses":[{"email_address":"orchids@example.com"}]}}]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/client/sessions/sess_test/tokens":
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "__client=orchids-client") || !strings.Contains(got, "__client_uat=") {
				t.Fatalf("unexpected orchids token cookie header: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jwt":"orchids-jwt"}`))
		default:
			t.Fatalf("unexpected orchids clerk request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer clerk.Close()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer orchids-jwt" {
			t.Fatalf("expected orchids bearer token, got %q", got)
		}
		var body struct {
			Prompt    string `json:"prompt"`
			ProjectID string `json:"projectId"`
			AgentMode string `json:"agentMode"`
			Email     string `json:"email"`
			UserID    string `json:"userId"`
			Model     string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode orchids upstream request: %v", err)
		}
		if body.Model != "claude-sonnet-4-5" {
			t.Fatalf("unexpected orchids model: %q", body.Model)
		}
		if body.ProjectID != "project-test" || body.AgentMode != "claude-opus-4.5" || body.Email != "orchids@example.com" || body.UserID != "user_test" {
			t.Fatalf("unexpected orchids payload identity: %#v", body)
		}
		if !strings.Contains(body.Prompt, "<user_request>\nhi\n</user_request>") {
			t.Fatalf("unexpected orchids prompt: %s", body.Prompt)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"text-start\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"text-delta\",\"delta\":\"hello \"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"text-delta\",\"delta\":\"orchids\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"finish\",\"finishReason\":\"stop\"}}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_ORCHIDS_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_CLERK_URL", clerk.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_CLIENT_COOKIE", "orchids-client")
	t.Setenv("NEWPLATFORM2API_ORCHIDS_PROJECT_ID", "project-test")
	t.Setenv("NEWPLATFORM2API_ORCHIDS_AGENT_MODE", "claude-opus-4.5")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"orchids","model":"claude-sonnet-4.5","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "orchids" {
		t.Fatalf("expected provider header to be orchids, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode orchids openai response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello orchids" {
		t.Fatalf("unexpected orchids response: %#v", response.Choices)
	}
}

func TestOpenAIChatOrchidsAcceptsRequestCredentialHeader(t *testing.T) {
	clerk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/client":
			cookie, err := r.Cookie("__client")
			if err != nil || cookie.Value != "orchids-header-client" {
				t.Fatalf("expected orchids client cookie on clerk bootstrap, got %v %v", cookie, err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"response":{"last_active_session_id":"sess_header","sessions":[{"user":{"id":"user_header","email_addresses":[{"email_address":"header@example.com"}]}}]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/client/sessions/sess_header/tokens":
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "__client=orchids-header-client") {
				t.Fatalf("unexpected orchids token cookie header: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jwt":"orchids-jwt-header"}`))
		default:
			t.Fatalf("unexpected orchids clerk request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer clerk.Close()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer orchids-jwt-header" {
			t.Fatalf("expected orchids bearer token, got %q", got)
		}
		var body struct {
			ProjectID string `json:"projectId"`
			Email     string `json:"email"`
			UserID    string `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode orchids upstream request: %v", err)
		}
		if body.ProjectID != "project-from-header-test" || body.Email != "header@example.com" || body.UserID != "user_header" {
			t.Fatalf("unexpected orchids payload identity: %#v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"text-delta\",\"delta\":\"hello request header\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"finish\",\"finishReason\":\"stop\"}}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_ORCHIDS_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_CLERK_URL", clerk.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_PROJECT_ID", "project-from-header-test")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"orchids","model":"claude-sonnet-4.5","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("X-Orchids-Client-Cookie", "orchids-header-client")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode orchids openai response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello request header" {
		t.Fatalf("unexpected orchids response: %#v", response.Choices)
	}
}

func TestOpenAIChatBlinkUsesConfiguredUpstream(t *testing.T) {
	firebase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected blink firebase method: %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("unexpected blink firebase content-type: %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse blink firebase form: %v", err)
		}
		if r.Form.Get("refresh_token") != "blink-refresh-token" {
			t.Fatalf("unexpected blink refresh_token: %q", r.Form.Get("refresh_token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id_token":"blink-id-token","refresh_token":"blink-refresh-token-next","expires_in":"3600"}`))
	}))
	defer firebase.Close()

	blink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/session":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink session body: %v", err)
			}
			if body["idToken"] != "blink-id-token" {
				t.Fatalf("unexpected blink session token: %#v", body)
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "blink-session-cookie", Path: "/"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-id-token" {
				t.Fatalf("unexpected blink session-data auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-session-cookie") {
				t.Fatalf("unexpected blink session-data cookie: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"active_workspace_id":"wsp_test"},"workspace":{"id":"wsp_test","slug":"workspace-test"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/projects/create":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-id-token" {
				t.Fatalf("unexpected blink project auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-session-cookie") || !strings.Contains(got, "workspace_slug=workspace-test") {
				t.Fatalf("unexpected blink project cookie: %q", got)
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink project body: %v", err)
			}
			if body["prompt"] != "build app" || body["app_type"] != "full_stack" {
				t.Fatalf("unexpected blink project payload: %#v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"project-blink"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/chat":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-id-token" {
				t.Fatalf("unexpected blink chat auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-session-cookie") || !strings.Contains(got, "workspace_slug=workspace-test") {
				t.Fatalf("unexpected blink chat cookie: %q", got)
			}
			var body struct {
				ID        string `json:"id"`
				ProjectID string `json:"projectId"`
				Message   struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"message"`
				ModelID string `json:"modelId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink chat body: %v", err)
			}
			if body.ProjectID != "project-blink" || body.ID != "project-blink" {
				t.Fatalf("unexpected blink chat project payload: %#v", body)
			}
			if body.ModelID != "anthropic/claude-sonnet-4.6" {
				t.Fatalf("unexpected blink model id: %q", body.ModelID)
			}
			if len(body.Message.Parts) != 1 || !strings.Contains(body.Message.Parts[0].Text, "<conversation_history>") || !strings.Contains(body.Message.Parts[0].Text, "<current_request>\nbuild app\n</current_request>") {
				t.Fatalf("unexpected blink chat prompt: %#v", body.Message.Parts)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"hello \"}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"blink\"}\n\n"))
			_, _ = w.Write([]byte("data: {\"type\":\"finish\",\"finishReason\":\"stop\"}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected blink request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	t.Setenv("NEWPLATFORM2API_BLINK_BASE_URL", blink.URL)
	t.Setenv("NEWPLATFORM2API_BLINK_FIREBASE_REFRESH_URL", firebase.URL)
	t.Setenv("NEWPLATFORM2API_BLINK_REFRESH_TOKEN", "blink-refresh-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"blink","model":"claude-sonnet-4.6","messages":[{"role":"system","content":"be concise"},{"role":"user","content":"hello"},{"role":"assistant","content":"hi"},{"role":"user","content":"build app"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected blink status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "blink" {
		t.Fatalf("expected provider header to be blink, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode blink openai response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello blink" {
		t.Fatalf("unexpected blink response: %#v", response.Choices)
	}
}

func TestOpenAIChatBlinkAcceptsDirectSessionHeaders(t *testing.T) {
	blink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-direct-id-token" {
				t.Fatalf("unexpected blink direct session-data auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-direct-session-token") {
				t.Fatalf("unexpected blink direct session-data cookie: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"active_workspace_id":"wsp_direct"},"workspace":{"id":"wsp_direct","slug":"workspace-direct"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/chat":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-direct-id-token" {
				t.Fatalf("unexpected blink direct chat auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-direct-session-token") || !strings.Contains(got, "workspace_slug=workspace-direct") {
				t.Fatalf("unexpected blink direct chat cookie: %q", got)
			}
			var body struct {
				ID        string `json:"id"`
				ProjectID string `json:"projectId"`
				ModelID   string `json:"modelId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink direct chat body: %v", err)
			}
			if body.ID != "project-direct" || body.ProjectID != "project-direct" {
				t.Fatalf("unexpected blink direct project payload: %#v", body)
			}
			if body.ModelID != "anthropic/claude-sonnet-4.6" {
				t.Fatalf("unexpected blink direct model id: %q", body.ModelID)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"header blink\"}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected blink direct request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	t.Setenv("NEWPLATFORM2API_BLINK_BASE_URL", blink.URL)

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"blink","model":"claude-sonnet-4.6","messages":[{"role":"user","content":"hello direct"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("X-Blink-Id-Token", "blink-direct-id-token")
	req.Header.Set("X-Blink-Session-Token", "blink-direct-session-token")
	req.Header.Set("X-Blink-Project-Id", "project-direct")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected blink direct status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode blink direct response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "header blink" {
		t.Fatalf("unexpected blink direct response: %#v", response.Choices)
	}
}

func TestOpenAIChatBlinkMapsGenericOpusAliasToOpus46(t *testing.T) {
	blink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"active_workspace_id":"wsp_opus"},"workspace":{"id":"wsp_opus","slug":"workspace-opus"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/chat":
			var body struct {
				ID        string `json:"id"`
				ProjectID string `json:"projectId"`
				ModelID   string `json:"modelId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink generic opus chat body: %v", err)
			}
			if body.ID != "project-opus" || body.ProjectID != "project-opus" {
				t.Fatalf("unexpected blink generic opus project payload: %#v", body)
			}
			if body.ModelID != "anthropic/claude-opus-4.6" {
				t.Fatalf("unexpected blink generic opus model id: %q", body.ModelID)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"opus blink\"}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected blink generic opus request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	t.Setenv("NEWPLATFORM2API_BLINK_BASE_URL", blink.URL)

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"blink","model":"opus[1m]","messages":[{"role":"user","content":"hello opus"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("X-Blink-Id-Token", "blink-opus-id-token")
	req.Header.Set("X-Blink-Session-Token", "blink-opus-session-token")
	req.Header.Set("X-Blink-Project-Id", "project-opus")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected blink generic opus status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode blink generic opus response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "opus blink" {
		t.Fatalf("unexpected blink generic opus response: %#v", response.Choices)
	}
}

func TestOpenAIChatBlinkUsesConfiguredDirectSessionUpstream(t *testing.T) {
	blink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-config-id-token" {
				t.Fatalf("unexpected blink config session-data auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-config-session-token") {
				t.Fatalf("unexpected blink config session-data cookie: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"active_workspace_id":"wsp_config"},"workspace":{"id":"wsp_config","slug":"workspace-config"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/chat":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-config-id-token" {
				t.Fatalf("unexpected blink config chat auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-config-session-token") || !strings.Contains(got, "workspace_slug=workspace-config") {
				t.Fatalf("unexpected blink config chat cookie: %q", got)
			}
			var body struct {
				ID        string `json:"id"`
				ProjectID string `json:"projectId"`
				ModelID   string `json:"modelId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink config chat body: %v", err)
			}
			if body.ID != "project-config" || body.ProjectID != "project-config" {
				t.Fatalf("unexpected blink config project payload: %#v", body)
			}
			if body.ModelID != "anthropic/claude-sonnet-4.6" {
				t.Fatalf("unexpected blink config model id: %q", body.ModelID)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"config blink\"}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected blink config request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	t.Setenv("NEWPLATFORM2API_BLINK_BASE_URL", blink.URL)
	t.Setenv("NEWPLATFORM2API_BLINK_ID_TOKEN", "blink-config-id-token")
	t.Setenv("NEWPLATFORM2API_BLINK_SESSION_TOKEN", "blink-config-session-token")
	t.Setenv("NEWPLATFORM2API_BLINK_WORKSPACE_SLUG", "workspace-config")
	t.Setenv("NEWPLATFORM2API_BLINK_PROJECT_ID", "project-config")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"blink","model":"claude-sonnet-4.6","messages":[{"role":"user","content":"hello config"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected blink config status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode blink config response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "config blink" {
		t.Fatalf("unexpected blink config response: %#v", response.Choices)
	}
}

func TestAnthropicMessagesBlinkUsesConfiguredUpstream(t *testing.T) {
	firebase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse blink firebase form: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id_token":"blink-anthropic-id-token","refresh_token":"blink-anthropic-refresh","expires_in":"3600"}`))
	}))
	defer firebase.Close()

	blink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/session":
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "blink-anthropic-session", Path: "/"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"workspace":{"id":"wsp_anthropic","slug":"workspace-anthropic"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/projects/create":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"project-anthropic"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/chat":
			var body struct {
				ModelID string `json:"modelId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink anthropic chat body: %v", err)
			}
			if body.ModelID != "openai/gpt-4o" {
				t.Fatalf("unexpected blink anthropic model id: %q", body.ModelID)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"type\":\"text-delta\",\"delta\":\"anthropic blink\"}\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
		default:
			t.Fatalf("unexpected blink anthropic request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	t.Setenv("NEWPLATFORM2API_BLINK_BASE_URL", blink.URL)
	t.Setenv("NEWPLATFORM2API_BLINK_FIREBASE_REFRESH_URL", firebase.URL)
	t.Setenv("NEWPLATFORM2API_BLINK_REFRESH_TOKEN", "blink-anthropic-refresh")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"blink","model":"gpt-4o","messages":[{"role":"user","content":"hello from anthropic"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected blink anthropic status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode blink anthropic response: %v", err)
	}
	if len(response.Content) != 1 || response.Content[0].Text != "anthropic blink" {
		t.Fatalf("unexpected blink anthropic response: %#v", response.Content)
	}
}

func TestAnthropicMessagesRejectsProviderWithoutAnthropicCompatibility(t *testing.T) {
	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"grok","model":"grok-4","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["error"] != `provider "grok" does not support Anthropic compatible endpoint` {
		t.Fatalf("unexpected error body: %#v", body)
	}
}

func TestAnthropicMessagesAcceptsXAPIKeyHeader(t *testing.T) {
	cfg := testAppConfig()
	registry := core.NewRegistry("anthropic-only")
	registry.Register(testAnthropicOnlyProvider{})
	h := NewHandler(registry, cfg)
	payload := []byte(`{"model":"stub-model","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("X-API-Key", "test-key")
	req.Header.Set("Anthropic-Version", "2023-06-01")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected anthropic x-api-key status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "anthropic-only" {
		t.Fatalf("expected provider header to be anthropic-only, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var body struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode anthropic x-api-key response: %v", err)
	}
	if len(body.Content) != 1 || body.Content[0].Text != "stub" {
		t.Fatalf("unexpected anthropic x-api-key response: %#v", body.Content)
	}
}

func TestProtectedRoutesRequireAPIKey(t *testing.T) {
	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without api key, got %d", rec.Code)
	}
}

func TestOpenAIChatCursorUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Cookie"); !strings.Contains(got, "WorkosCursorSessionToken=test-cookie") {
			t.Errorf("expected cursor cookie header, got %q", got)
		}
		if got := r.Header.Get("X-Is-Human"); got != "test-human-token" {
			t.Errorf("expected x-is-human header, got %q", got)
		}

		var body struct {
			Model    string `json:"model"`
			Trigger  string `json:"trigger"`
			Messages []struct {
				Role  string `json:"role"`
				Parts []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.Model != "anthropic/claude-sonnet-4.6" {
			t.Errorf("expected mapped upstream model, got %q", body.Model)
		}
		if body.Trigger != "submit-message" {
			t.Errorf("expected trigger submit-message, got %q", body.Trigger)
		}
		if len(body.Messages) != 1 || len(body.Messages[0].Parts) != 1 || body.Messages[0].Parts[0].Text != "hi" {
			t.Errorf("unexpected upstream messages: %#v", body.Messages)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\"hello \"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\"world\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"finish\"}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CURSOR_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CURSOR_COOKIE", "WorkosCursorSessionToken=test-cookie")
	t.Setenv("NEWPLATFORM2API_CURSOR_X_IS_HUMAN", "test-human-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"cursor","model":"claude-sonnet-4.6","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "cursor" {
		t.Fatalf("expected provider header to be cursor, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello world" {
		t.Fatalf("unexpected assistant content: %#v", response.Choices)
	}
}

func TestOpenAIChatCursorStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\"hello\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\" framework\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"finish\"}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CURSOR_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CURSOR_COOKIE", "WorkosCursorSessionToken=test-cookie")
	t.Setenv("NEWPLATFORM2API_CURSOR_X_IS_HUMAN", "test-human-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"cursor","model":"claude-sonnet-4.6","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"content":"hello"`) || !strings.Contains(body, `"content":" framework"`) {
		t.Fatalf("expected streamed deltas in response body, got %s", body)
	}
	if !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("expected [DONE] marker, got %s", body)
	}
}

func TestAnthropicMessagesCursorUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[0].Parts[0].Text != "be precise" {
			t.Fatalf("expected system message to be prepended, got %#v", body.Messages)
		}
		if body.Messages[1].Role != "user" || body.Messages[1].Parts[0].Text != "hi" {
			t.Fatalf("unexpected user message: %#v", body.Messages)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\"hello anthropic\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"finish\"}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CURSOR_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CURSOR_COOKIE", "WorkosCursorSessionToken=test-cookie")
	t.Setenv("NEWPLATFORM2API_CURSOR_X_IS_HUMAN", "test-human-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"cursor","model":"claude-sonnet-4.6","system":"be precise","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Content) != 1 || response.Content[0].Text != "hello anthropic" {
		t.Fatalf("unexpected anthropic response: %#v", response.Content)
	}
}

func TestAnthropicMessagesOrchidsUsesConfiguredUpstream(t *testing.T) {
	clerk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/client":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"response":{"last_active_session_id":"sess_test","sessions":[{"user":{"id":"user_test","email_addresses":[{"email_address":"orchids@example.com"}]}}]}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/client/sessions/sess_test/tokens":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jwt":"orchids-jwt"}`))
		default:
			t.Fatalf("unexpected orchids clerk request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer clerk.Close()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode orchids upstream request: %v", err)
		}
		if !strings.Contains(body.Prompt, "<client_system>\nbe precise\n</client_system>") {
			t.Fatalf("expected orchids prompt to include system block, got %s", body.Prompt)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"text-delta\",\"delta\":\"hello orchids anthropic\"}}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"model\",\"event\":{\"type\":\"finish\",\"finishReason\":\"stop\"}}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_ORCHIDS_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_CLERK_URL", clerk.URL)
	t.Setenv("NEWPLATFORM2API_ORCHIDS_CLIENT_COOKIE", "orchids-client")
	t.Setenv("NEWPLATFORM2API_ORCHIDS_PROJECT_ID", "project-test")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"orchids","model":"claude-sonnet-4.5","system":"be precise","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "orchids" {
		t.Fatalf("expected provider header to be orchids, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Content) != 1 || response.Content[0].Text != "hello orchids anthropic" {
		t.Fatalf("unexpected orchids response: %#v", response.Content)
	}
}

func TestAnthropicMessagesCursorStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\"hello\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"delta\",\"delta\":\" anthropic\"}\n\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"finish\"}\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CURSOR_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CURSOR_COOKIE", "WorkosCursorSessionToken=test-cookie")
	t.Setenv("NEWPLATFORM2API_CURSOR_X_IS_HUMAN", "test-human-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"cursor","model":"claude-sonnet-4.6","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: message_start") || !strings.Contains(body, "event: content_block_delta") || !strings.Contains(body, `"text":"hello"`) || !strings.Contains(body, `"text":" anthropic"`) || !strings.Contains(body, "event: message_stop") {
		t.Fatalf("unexpected anthropic stream body: %s", body)
	}
}

func TestOpenAIChatKiroUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer kiro-token" {
			t.Fatalf("expected kiro bearer token, got %q", got)
		}
		if got := r.Header.Get("X-Amz-Target"); got != "AmazonCodeWhispererStreamingService.GenerateAssistantResponse" {
			t.Fatalf("unexpected x-amz-target: %q", got)
		}
		var body struct {
			ConversationState struct {
				CurrentMessage struct {
					UserInputMessage struct {
						Content string `json:"content"`
						ModelID string `json:"modelId"`
					} `json:"userInputMessage"`
				} `json:"currentMessage"`
			} `json:"conversationState"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.ConversationState.CurrentMessage.UserInputMessage.Content != "hi" {
			t.Fatalf("unexpected kiro current content: %q", body.ConversationState.CurrentMessage.UserInputMessage.Content)
		}
		if body.ConversationState.CurrentMessage.UserInputMessage.ModelID != "claude-sonnet-4.6" {
			t.Fatalf("unexpected kiro model id: %q", body.ConversationState.CurrentMessage.UserInputMessage.ModelID)
		}
		_, _ = w.Write(encodeKiroTestStream(kiroTestEvent{eventType: "assistantResponseEvent", payload: map[string]interface{}{"content": "hello kiro"}}))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_KIRO_ACCESS_TOKEN", "kiro-token")
	t.Setenv("NEWPLATFORM2API_KIRO_CODEWHISPERER_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_KIRO_AMAZONQ_URL", upstream.URL)

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"kiro","model":"claude-sonnet-4.6","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello kiro" {
		t.Fatalf("unexpected kiro response: %#v", response.Choices)
	}
}

func TestAnthropicMessagesKiroStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ConversationState struct {
				CurrentMessage struct {
					UserInputMessage struct {
						Content string `json:"content"`
					} `json:"userInputMessage"`
				} `json:"currentMessage"`
			} `json:"conversationState"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.ConversationState.CurrentMessage.UserInputMessage.Content != "be precise\n\nhi" {
			t.Fatalf("unexpected anthropic kiro content: %q", body.ConversationState.CurrentMessage.UserInputMessage.Content)
		}
		_, _ = w.Write(encodeKiroTestStream(
			kiroTestEvent{eventType: "assistantResponseEvent", payload: map[string]interface{}{"content": "hello"}},
			kiroTestEvent{eventType: "assistantResponseEvent", payload: map[string]interface{}{"content": "hello anthropic"}},
		))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_KIRO_ACCESS_TOKEN", "kiro-token")
	t.Setenv("NEWPLATFORM2API_KIRO_CODEWHISPERER_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_KIRO_AMAZONQ_URL", upstream.URL)

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"kiro","model":"claude-sonnet-4.6","system":"be precise","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"text":"hello"`) || !strings.Contains(body, `"text":" anthropic"`) {
		t.Fatalf("unexpected kiro anthropic stream body: %s", body)
	}
}

func TestOpenAIChatGrokUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Cookie"); !strings.Contains(got, "sso=grok-cookie") {
			t.Fatalf("expected grok sso cookie, got %q", got)
		}
		if got := r.Header.Get("Cookie"); !strings.Contains(got, "cf_clearance=fresh-clearance") || strings.Contains(got, "cf_clearance=old") {
			t.Fatalf("expected refreshed cf_clearance cookie, got %q", got)
		}
		var body struct {
			Message   string `json:"message"`
			ModelName string `json:"modelName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.Message != "assistant: earlier\n\nhi" {
			t.Fatalf("unexpected flattened grok message: %q", body.Message)
		}
		if body.ModelName != "grok-4" {
			t.Fatalf("unexpected grok model: %q", body.ModelName)
		}
		_, _ = w.Write([]byte("{\"result\":{\"response\":{\"token\":\"hello\"}}}\n"))
		_, _ = w.Write([]byte("{\"result\":{\"response\":{\"token\":\" grok\"}}}\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_GROK_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_GROK_COOKIE_TOKEN", "grok-cookie")
	t.Setenv("NEWPLATFORM2API_GROK_CF_COOKIES", "theme=dark; cf_clearance=old")
	t.Setenv("NEWPLATFORM2API_GROK_CF_CLEARANCE", "fresh-clearance")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"grok","model":"grok-4","messages":[{"role":"assistant","content":"earlier"},{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello grok" {
		t.Fatalf("unexpected grok response: %#v", response.Choices)
	}
}

func TestOpenAIChatGrokStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{\"result\":{\"response\":{\"token\":\"hello\"}}}\n"))
		_, _ = w.Write([]byte("{\"result\":{\"response\":{\"token\":\" world\"}}}\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_GROK_API_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_GROK_COOKIE_TOKEN", "grok-cookie")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"grok","model":"grok-4","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"content":"hello"`) || !strings.Contains(body, `"content":" world"`) || !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("unexpected grok stream body: %s", body)
	}
}

func TestOpenAIChatWebUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/claude/v1/chat/completions" {
			t.Fatalf("unexpected web path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer web-key" {
			t.Fatalf("expected web bearer api key, got %q", got)
		}
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string      `json:"role"`
				Content interface{} `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.Model != "claude-sonnet-4.5" {
			t.Fatalf("unexpected web model: %q", body.Model)
		}
		if len(body.Messages) != 1 || body.Messages[0].Role != "user" || core.ContentText(body.Messages[0].Content) != "hi" {
			t.Fatalf("unexpected web messages: %#v", body.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello web"}}]}`))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_WEB_BASE_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_WEB_TYPE", "claude")
	t.Setenv("NEWPLATFORM2API_WEB_API_KEY", "web-key")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"web","model":"claude-sonnet-4.5","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "web" {
		t.Fatalf("expected provider header to be web, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello web" {
		t.Fatalf("unexpected web response: %#v", response.Choices)
	}
}

func TestOpenAIChatWebStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/claude/v1/chat/completions" {
			t.Fatalf("unexpected web path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" web\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_WEB_BASE_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_WEB_TYPE", "claude")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"web","model":"claude-sonnet-4.5","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"content":"hello"`) || !strings.Contains(body, `"content":" web"`) || !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("unexpected web stream body: %s", body)
	}
}

func TestOpenAIChatChatGPTUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected chatgpt path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer chatgpt-token" {
			t.Fatalf("expected chatgpt bearer token, got %q", got)
		}
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string      `json:"role"`
				Content interface{} `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		if body.Model != "gpt-4.1" {
			t.Fatalf("unexpected chatgpt model: %q", body.Model)
		}
		if len(body.Messages) != 1 || body.Messages[0].Role != "user" || core.ContentText(body.Messages[0].Content) != "hi" {
			t.Fatalf("unexpected chatgpt messages: %#v", body.Messages)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello chatgpt"}}]}`))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CHATGPT_BASE_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CHATGPT_TOKEN", "chatgpt-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"chatgpt","model":"gpt-4.1","messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "chatgpt" {
		t.Fatalf("expected provider header to be chatgpt, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "hello chatgpt" {
		t.Fatalf("unexpected chatgpt response: %#v", response.Choices)
	}
}

func TestOpenAIChatChatGPTStreamsConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected chatgpt path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer chatgpt-token" {
			t.Fatalf("expected chatgpt bearer token, got %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" chatgpt\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer upstream.Close()

	t.Setenv("NEWPLATFORM2API_CHATGPT_BASE_URL", upstream.URL)
	t.Setenv("NEWPLATFORM2API_CHATGPT_TOKEN", "chatgpt-token")

	cfg := testAppConfig()
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)
	payload := []byte(`{"provider":"chatgpt","model":"gpt-4.1","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"content":"hello"`) || !strings.Contains(body, `"content":" chatgpt"`) || !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("unexpected chatgpt stream body: %s", body)
	}
}

type kiroTestEvent struct {
	eventType string
	payload   interface{}
}

func encodeKiroTestStream(events ...kiroTestEvent) []byte {
	var output bytes.Buffer
	for _, event := range events {
		payload, _ := json.Marshal(event.payload)
		headers := encodeKiroTestHeader(":event-type", event.eventType)
		prelude := make([]byte, 12)
		totalLength := uint32(12 + len(headers) + len(payload) + 4)
		binary.BigEndian.PutUint32(prelude[0:4], totalLength)
		binary.BigEndian.PutUint32(prelude[4:8], uint32(len(headers)))
		output.Write(prelude)
		output.Write(headers)
		output.Write(payload)
		output.Write([]byte{0, 0, 0, 0})
	}
	return output.Bytes()
}

func encodeKiroTestHeader(name string, value string) []byte {
	header := make([]byte, 0, 1+len(name)+1+2+len(value))
	header = append(header, byte(len(name)))
	header = append(header, []byte(name)...)
	header = append(header, 7)
	header = append(header, byte(len(value)>>8), byte(len(value)))
	header = append(header, []byte(value)...)
	return header
}

func TestOpenAIImagesGenerationZAIUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "zai-image-session" {
			t.Fatalf("expected zai image session cookie, got cookie=%v err=%v", cookie, err)
		}
		var body struct {
			Prompt           string `json:"prompt"`
			Ratio            string `json:"ratio"`
			Resolution       string `json:"resolution"`
			RmLabelWatermark bool   `json:"rm_label_watermark"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode zai image request: %v", err)
		}
		if body.Prompt != "draw a cat" || body.Ratio != "9:16" || body.Resolution != "2K" || !body.RmLabelWatermark {
			t.Fatalf("unexpected zai image request: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","timestamp":1710000000,"data":{"image":{"image_id":"img_1","prompt":"draw a cat","size":"1024x1792","ratio":"9:16","resolution":"2K","image_url":"https://cdn.test/cat.png","status":"success","width":1024,"height":1792}}}`))
	}))
	defer upstream.Close()

	cfg := testAppConfig()
	cfg.ZAIImage.SessionToken = "zai-image-session"
	cfg.ZAIImage.APIURL = upstream.URL
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)

	payload := []byte(`{"prompt":"draw a cat","size":"1024x1792"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "zai_image" {
		t.Fatalf("expected provider header zai_image, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var body struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
			Size          string `json:"size"`
			Width         int    `json:"width"`
			Height        int    `json:"height"`
			Ratio         string `json:"ratio"`
			Resolution    string `json:"resolution"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode zai image response: %v", err)
	}
	if body.Created != 1710000000 || len(body.Data) != 1 {
		t.Fatalf("unexpected zai image envelope: %#v", body)
	}
	if item := body.Data[0]; item.URL != "https://cdn.test/cat.png" || item.Size != "1024x1792" || item.Width != 1024 || item.Height != 1792 || item.Ratio != "9:16" || item.Resolution != "2K" {
		t.Fatalf("unexpected zai image item: %#v", item)
	}
}

func TestOpenAIAudioSpeechZAIUsesConfiguredUpstream(t *testing.T) {
	expectedAudio := []byte("wav-bytes")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer zai-tts-token" {
			t.Fatalf("expected bearer token, got %q", got)
		}
		var body struct {
			VoiceName string  `json:"voice_name"`
			VoiceID   string  `json:"voice_id"`
			UserID    string  `json:"user_id"`
			InputText string  `json:"input_text"`
			Speed     float64 `json:"speed"`
			Volume    float64 `json:"volume"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode zai tts request: %v", err)
		}
		if body.VoiceID != "system_003" || body.VoiceName != "通用男声" || body.UserID != "user-123" || body.InputText != "speak now" || body.Speed != 1 || body.Volume != 1 {
			t.Fatalf("unexpected zai tts request: %#v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"audio\":\"" + base64.StdEncoding.EncodeToString(expectedAudio) + "\"}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer upstream.Close()

	cfg := testAppConfig()
	cfg.ZAITTS.Token = "zai-tts-token"
	cfg.ZAITTS.UserID = "user-123"
	cfg.ZAITTS.APIURL = upstream.URL
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)

	payload := []byte(`{"input":"speak now"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "zai_tts" {
		t.Fatalf("expected provider header zai_tts, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "audio/wav" {
		t.Fatalf("expected audio/wav content type, got %q", contentType)
	}
	if !bytes.Equal(rec.Body.Bytes(), expectedAudio) {
		t.Fatalf("unexpected audio bytes: %q", rec.Body.Bytes())
	}
}

func TestOCRUploadZAIUsesConfiguredUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer zai-ocr-token" {
			t.Fatalf("expected bearer token, got %q", got)
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatalf("parse upstream multipart: %v", err)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("expected file field: %v", err)
		}
		defer file.Close()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)
		if header.Filename != "note.txt" || buf.String() != "hello ocr" {
			t.Fatalf("unexpected upstream file: filename=%q body=%q", header.Filename, buf.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"task_id":"task_1","status":"success","file_name":"note.txt","file_size":9,"file_type":"text/plain","file_url":"https://cdn.test/note.txt","created_at":"2025-01-01T00:00:00Z","markdown_content":"# hello","json_content":"{\"md_results\":[],\"layout_details\":[],\"data_info\":[],\"usage\":{\"pages\":1,\"tokens\":2}}","layout":[{"type":"text","sub_type":"paragraph","content":"hello","bbox":[0,0,10,10],"order":1,"page_idx":0}]}}`))
	}))
	defer upstream.Close()

	cfg := testAppConfig()
	cfg.ZAIOCR.Token = "zai-ocr-token"
	cfg.ZAIOCR.APIURL = upstream.URL
	h := NewHandler(platforms.DefaultRegistry(cfg), cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "note.txt")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	_, _ = part.Write([]byte("hello ocr"))
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/ocr", body)
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Newplatform2API-Provider") != "zai_ocr" {
		t.Fatalf("expected provider header zai_ocr, got %q", rec.Header().Get("X-Newplatform2API-Provider"))
	}
	var resp struct {
		ID       string                   `json:"id"`
		Object   string                   `json:"object"`
		Model    string                   `json:"model"`
		Status   string                   `json:"status"`
		Text     string                   `json:"text"`
		Markdown string                   `json:"markdown"`
		JSON     map[string]interface{}   `json:"json"`
		Layout   []map[string]interface{} `json:"layout"`
		File     struct {
			Name string `json:"name"`
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"file"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode ocr response: %v", err)
	}
	if resp.ID != "task_1" || resp.Object != "ocr.result" || resp.Model != "zai-ocr" || resp.Status != "success" || resp.Text != "# hello" || resp.Markdown != "# hello" {
		t.Fatalf("unexpected ocr response: %#v", resp)
	}
	usage, _ := resp.JSON["usage"].(map[string]interface{})
	if usage["pages"] != float64(1) || usage["tokens"] != float64(2) || len(resp.Layout) != 1 || resp.File.Name != "note.txt" || resp.File.Type != "text/plain" || resp.File.URL != "https://cdn.test/note.txt" {
		t.Fatalf("unexpected ocr detail response: %#v", resp)
	}
}

func TestAdminSettingsCanRotateRuntimeAPIKey(t *testing.T) {
	h, cookie := newRuntimeAdminHandler(t)

	settingsPayload := []byte(`{"apiKey":"runtime-key","defaultProvider":"kiro","adminPassword":"new-secret"}`)
	settingsReq := httptest.NewRequest(http.MethodPut, "/admin/api/settings", bytes.NewReader(settingsPayload))
	settingsReq.AddCookie(cookie)
	settingsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(settingsRec, settingsReq)
	if settingsRec.Code != http.StatusOK {
		t.Fatalf("unexpected settings status: %d body=%s", settingsRec.Code, settingsRec.Body.String())
	}

	unauthorized := httptest.NewRecorder()
	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	unauthorizedReq.Header.Set("Authorization", "Bearer test-key")
	h.Routes().ServeHTTP(unauthorized, unauthorizedReq)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected old api key to fail, got %d", unauthorized.Code)
	}

	authorized := httptest.NewRecorder()
	authorizedReq := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	authorizedReq.Header.Set("Authorization", "Bearer runtime-key")
	h.Routes().ServeHTTP(authorized, authorizedReq)
	if authorized.Code != http.StatusOK {
		t.Fatalf("expected new api key to work, got %d body=%s", authorized.Code, authorized.Body.String())
	}
}

func TestAdminProviderEndpointsPersistCountsAndSelections(t *testing.T) {
	h, cookie := newRuntimeAdminHandler(t)

	optionsReq := httptest.NewRequest(http.MethodOptions, "/admin/api/settings", nil)
	optionsReq.Header.Set("Origin", "http://localhost:1420")
	optionsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(optionsRec, optionsReq)
	if optionsRec.Code != http.StatusNoContent {
		t.Fatalf("unexpected options status: %d body=%s", optionsRec.Code, optionsRec.Body.String())
	}
	if allowMethods := optionsRec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(allowMethods, http.MethodDelete) {
		t.Fatalf("expected delete in cors methods, got %q", allowMethods)
	}

	cursorPayload := []byte(`{"config":{"cookie":"cursor-cookie","scriptUrl":"https://cursor.com/_next/static/chunks/pages/_app.js","userAgent":"Mozilla/5.0","referer":"https://cursor.com/"}}`)
	cursorReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/cursor/config", bytes.NewReader(cursorPayload))
	cursorReq.AddCookie(cookie)
	cursorRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(cursorRec, cursorReq)
	if cursorRec.Code != http.StatusOK {
		t.Fatalf("unexpected cursor status: %d body=%s", cursorRec.Code, cursorRec.Body.String())
	}

	kiroCreatePrimaryReq := httptest.NewRequest(http.MethodPost, "/admin/api/providers/kiro/accounts/create", bytes.NewReader([]byte(`{"name":"Primary","accessToken":"kiro-token","machineId":"machine-1","preferredEndpoint":"amazonq","active":true}`)))
	kiroCreatePrimaryReq.AddCookie(cookie)
	kiroCreatePrimaryRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroCreatePrimaryRec, kiroCreatePrimaryReq)
	if kiroCreatePrimaryRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro create status: %d body=%s", kiroCreatePrimaryRec.Code, kiroCreatePrimaryRec.Body.String())
	}
	var kiroPrimaryBody struct {
		Account core.KiroAccount `json:"account"`
	}
	if err := json.NewDecoder(kiroCreatePrimaryRec.Body).Decode(&kiroPrimaryBody); err != nil {
		t.Fatalf("decode kiro primary create: %v", err)
	}

	kiroCreateBackupReq := httptest.NewRequest(http.MethodPost, "/admin/api/providers/kiro/accounts/create", bytes.NewReader([]byte(`{"name":"Backup","accessToken":"kiro-backup","machineId":"machine-2","preferredEndpoint":"codewhisperer","active":false}`)))
	kiroCreateBackupReq.AddCookie(cookie)
	kiroCreateBackupRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroCreateBackupRec, kiroCreateBackupReq)
	if kiroCreateBackupRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro backup create status: %d body=%s", kiroCreateBackupRec.Code, kiroCreateBackupRec.Body.String())
	}
	var kiroBackupBody struct {
		Account core.KiroAccount `json:"account"`
	}
	if err := json.NewDecoder(kiroCreateBackupRec.Body).Decode(&kiroBackupBody); err != nil {
		t.Fatalf("decode kiro backup create: %v", err)
	}

	kiroDetailReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/kiro/accounts/detail/"+kiroPrimaryBody.Account.ID, nil)
	kiroDetailReq.AddCookie(cookie)
	kiroDetailRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroDetailRec, kiroDetailReq)
	if kiroDetailRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro detail status: %d body=%s", kiroDetailRec.Code, kiroDetailRec.Body.String())
	}

	kiroUpdateReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/kiro/accounts/update/"+kiroBackupBody.Account.ID, bytes.NewReader([]byte(`{"name":"Backup Updated","accessToken":"kiro-backup-2","machineId":"machine-2b","preferredEndpoint":"codewhisperer","active":true}`)))
	kiroUpdateReq.AddCookie(cookie)
	kiroUpdateRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroUpdateRec, kiroUpdateReq)
	if kiroUpdateRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro update status: %d body=%s", kiroUpdateRec.Code, kiroUpdateRec.Body.String())
	}
	var kiroUpdatedBody struct {
		Account core.KiroAccount `json:"account"`
	}
	if err := json.NewDecoder(kiroUpdateRec.Body).Decode(&kiroUpdatedBody); err != nil {
		t.Fatalf("decode kiro update: %v", err)
	}
	if !kiroUpdatedBody.Account.Active || kiroUpdatedBody.Account.Name != "Backup Updated" {
		t.Fatalf("unexpected kiro update response: %#v", kiroUpdatedBody.Account)
	}

	kiroDeleteReq := httptest.NewRequest(http.MethodDelete, "/admin/api/providers/kiro/accounts/delete/"+kiroPrimaryBody.Account.ID, nil)
	kiroDeleteReq.AddCookie(cookie)
	kiroDeleteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroDeleteRec, kiroDeleteReq)
	if kiroDeleteRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro delete status: %d body=%s", kiroDeleteRec.Code, kiroDeleteRec.Body.String())
	}

	kiroListReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/kiro/accounts/list", nil)
	kiroListReq.AddCookie(cookie)
	kiroListRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(kiroListRec, kiroListReq)
	if kiroListRec.Code != http.StatusOK {
		t.Fatalf("unexpected kiro list status: %d body=%s", kiroListRec.Code, kiroListRec.Body.String())
	}
	var kiroListBody struct {
		Accounts []core.KiroAccount `json:"accounts"`
	}
	if err := json.NewDecoder(kiroListRec.Body).Decode(&kiroListBody); err != nil {
		t.Fatalf("decode kiro list: %v", err)
	}
	if len(kiroListBody.Accounts) != 1 || kiroListBody.Accounts[0].ID != kiroBackupBody.Account.ID || !kiroListBody.Accounts[0].Active {
		t.Fatalf("unexpected kiro list response: %#v", kiroListBody.Accounts)
	}

	grokConfigPayload := []byte(`{"config":{"apiUrl":"https://grok.test/chat","proxyUrl":"http://127.0.0.1:7890","cfCookies":"theme=dark","cfClearance":"cf-token","userAgent":"Mozilla/Test","origin":"https://grok.test","referer":"https://grok.test/"}}`)
	grokConfigReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/grok/config", bytes.NewReader(grokConfigPayload))
	grokConfigReq.AddCookie(cookie)
	grokConfigRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokConfigRec, grokConfigReq)
	if grokConfigRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok config status: %d body=%s", grokConfigRec.Code, grokConfigRec.Body.String())
	}

	grokCreatePrimaryReq := httptest.NewRequest(http.MethodPost, "/admin/api/providers/grok/tokens/create", bytes.NewReader([]byte(`{"name":"Primary","cookieToken":"grok-cookie-1","active":false}`)))
	grokCreatePrimaryReq.AddCookie(cookie)
	grokCreatePrimaryRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokCreatePrimaryRec, grokCreatePrimaryReq)
	if grokCreatePrimaryRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok create status: %d body=%s", grokCreatePrimaryRec.Code, grokCreatePrimaryRec.Body.String())
	}
	var grokPrimaryBody struct {
		Token core.GrokToken `json:"token"`
	}
	if err := json.NewDecoder(grokCreatePrimaryRec.Body).Decode(&grokPrimaryBody); err != nil {
		t.Fatalf("decode grok primary create: %v", err)
	}

	grokCreateBackupReq := httptest.NewRequest(http.MethodPost, "/admin/api/providers/grok/tokens/create", bytes.NewReader([]byte(`{"name":"Secondary","cookieToken":"grok-cookie-2","active":true}`)))
	grokCreateBackupReq.AddCookie(cookie)
	grokCreateBackupRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokCreateBackupRec, grokCreateBackupReq)
	if grokCreateBackupRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok backup create status: %d body=%s", grokCreateBackupRec.Code, grokCreateBackupRec.Body.String())
	}
	var grokBackupBody struct {
		Token core.GrokToken `json:"token"`
	}
	if err := json.NewDecoder(grokCreateBackupRec.Body).Decode(&grokBackupBody); err != nil {
		t.Fatalf("decode grok backup create: %v", err)
	}

	grokDetailReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/grok/tokens/detail/"+grokPrimaryBody.Token.ID, nil)
	grokDetailReq.AddCookie(cookie)
	grokDetailRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokDetailRec, grokDetailReq)
	if grokDetailRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok detail status: %d body=%s", grokDetailRec.Code, grokDetailRec.Body.String())
	}

	grokUpdateReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/grok/tokens/update/"+grokBackupBody.Token.ID, bytes.NewReader([]byte(`{"name":"Secondary Updated","cookieToken":"grok-cookie-2b","active":true}`)))
	grokUpdateReq.AddCookie(cookie)
	grokUpdateRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokUpdateRec, grokUpdateReq)
	if grokUpdateRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok update status: %d body=%s", grokUpdateRec.Code, grokUpdateRec.Body.String())
	}
	var grokUpdatedBody struct {
		Token core.GrokToken `json:"token"`
	}
	if err := json.NewDecoder(grokUpdateRec.Body).Decode(&grokUpdatedBody); err != nil {
		t.Fatalf("decode grok update: %v", err)
	}
	if !grokUpdatedBody.Token.Active || grokUpdatedBody.Token.Name != "Secondary Updated" {
		t.Fatalf("unexpected grok update response: %#v", grokUpdatedBody.Token)
	}

	grokDeleteReq := httptest.NewRequest(http.MethodDelete, "/admin/api/providers/grok/tokens/delete/"+grokPrimaryBody.Token.ID, nil)
	grokDeleteReq.AddCookie(cookie)
	grokDeleteRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokDeleteRec, grokDeleteReq)
	if grokDeleteRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok delete status: %d body=%s", grokDeleteRec.Code, grokDeleteRec.Body.String())
	}

	grokListReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/grok/tokens/list", nil)
	grokListReq.AddCookie(cookie)
	grokListRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokListRec, grokListReq)
	if grokListRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok list status: %d body=%s", grokListRec.Code, grokListRec.Body.String())
	}
	var grokListBody struct {
		Tokens []core.GrokToken `json:"tokens"`
	}
	if err := json.NewDecoder(grokListRec.Body).Decode(&grokListBody); err != nil {
		t.Fatalf("decode grok list: %v", err)
	}
	if len(grokListBody.Tokens) != 1 || grokListBody.Tokens[0].ID != grokBackupBody.Token.ID || !grokListBody.Tokens[0].Active {
		t.Fatalf("unexpected grok list response: %#v", grokListBody.Tokens)
	}

	orchidsPayload := []byte(`{"config":{"clientCookie":"orchids-cookie","clientUat":"123","sessionId":"sess-1","projectId":"project-1","userId":"user-1","email":"user@example.com","agentMode":"claude-opus-4.5"}}`)
	orchidsReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/orchids/config", bytes.NewReader(orchidsPayload))
	orchidsReq.AddCookie(cookie)
	orchidsRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(orchidsRec, orchidsReq)
	if orchidsRec.Code != http.StatusOK {
		t.Fatalf("unexpected orchids status: %d body=%s", orchidsRec.Code, orchidsRec.Body.String())
	}

	webPayload := []byte(`{"config":{"baseUrl":"http://127.0.0.1:9000","type":"claude","apiKey":"web-key"}}`)
	webReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/web/config", bytes.NewReader(webPayload))
	webReq.AddCookie(cookie)
	webRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(webRec, webReq)
	if webRec.Code != http.StatusOK {
		t.Fatalf("unexpected web status: %d body=%s", webRec.Code, webRec.Body.String())
	}

	chatgptPayload := []byte(`{"config":{"baseUrl":"http://127.0.0.1:5005","token":"chatgpt-token"}}`)
	chatgptReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/chatgpt/config", bytes.NewReader(chatgptPayload))
	chatgptReq.AddCookie(cookie)
	chatgptRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(chatgptRec, chatgptReq)
	if chatgptRec.Code != http.StatusOK {
		t.Fatalf("unexpected chatgpt status: %d body=%s", chatgptRec.Code, chatgptRec.Body.String())
	}

	blinkPayload := []byte(`{"config":{"baseUrl":"https://blink.test","firebaseRefreshUrl":"https://firebase.test/token","refreshToken":"blink-refresh-token","idToken":"blink-id-token","sessionToken":"blink-session-token","workspaceSlug":"workspace-test","projectId":"project-test","proxyUrl":"http://127.0.0.1:7891"}}`)
	blinkReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/blink/config", bytes.NewReader(blinkPayload))
	blinkReq.AddCookie(cookie)
	blinkRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(blinkRec, blinkReq)
	if blinkRec.Code != http.StatusOK {
		t.Fatalf("unexpected blink status: %d body=%s", blinkRec.Code, blinkRec.Body.String())
	}

	zaiImagePayload := []byte(`{"config":{"sessionToken":"zai-image-session","apiUrl":"https://image.test/generate"}}`)
	zaiImageReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/zai/image/config", bytes.NewReader(zaiImagePayload))
	zaiImageReq.AddCookie(cookie)
	zaiImageRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiImageRec, zaiImageReq)
	if zaiImageRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai image status: %d body=%s", zaiImageRec.Code, zaiImageRec.Body.String())
	}

	zaiTTSPayload := []byte(`{"config":{"token":"zai-tts-token","userId":"tts-user","apiUrl":"https://audio.test/tts"}}`)
	zaiTTSReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/zai/tts/config", bytes.NewReader(zaiTTSPayload))
	zaiTTSReq.AddCookie(cookie)
	zaiTTSRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiTTSRec, zaiTTSReq)
	if zaiTTSRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai tts status: %d body=%s", zaiTTSRec.Code, zaiTTSRec.Body.String())
	}

	zaiOCRPayload := []byte(`{"config":{"token":"zai-ocr-token","apiUrl":"https://ocr.test/process"}}`)
	zaiOCRReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/zai/ocr/config", bytes.NewReader(zaiOCRPayload))
	zaiOCRReq.AddCookie(cookie)
	zaiOCRRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiOCRRec, zaiOCRReq)
	if zaiOCRRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai ocr status: %d body=%s", zaiOCRRec.Code, zaiOCRRec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/admin/api/status", nil)
	statusReq.AddCookie(cookie)
	statusRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("unexpected status endpoint status: %d body=%s", statusRec.Code, statusRec.Body.String())
	}
	var body struct {
		Providers struct {
			Cursor struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"cursor"`
			Kiro struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"kiro"`
			Grok struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"grok"`
			Orchids struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"orchids"`
			Web struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"web"`
			ChatGPT struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"chatgpt"`
			Blink struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"blink"`
			ZAIImage struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"zaiImage"`
			ZAITTS struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"zaiTTS"`
			ZAIOCR struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"zaiOCR"`
		} `json:"providers"`
	}
	if err := json.NewDecoder(statusRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode admin status: %v", err)
	}
	if body.Providers.Cursor.Count != 1 || !body.Providers.Cursor.Configured || body.Providers.Cursor.Active == "" {
		t.Fatalf("unexpected cursor status: %#v", body.Providers.Cursor)
	}
	if body.Providers.Kiro.Count != 1 || !body.Providers.Kiro.Configured || body.Providers.Kiro.Active != kiroBackupBody.Account.ID {
		t.Fatalf("unexpected kiro status: %#v", body.Providers.Kiro)
	}
	if body.Providers.Grok.Count != 1 || !body.Providers.Grok.Configured || body.Providers.Grok.Active != grokBackupBody.Token.ID {
		t.Fatalf("unexpected grok status: %#v", body.Providers.Grok)
	}
	if body.Providers.Orchids.Count != 1 || !body.Providers.Orchids.Configured || body.Providers.Orchids.Active == "" {
		t.Fatalf("unexpected orchids status: %#v", body.Providers.Orchids)
	}
	if body.Providers.Web.Count != 1 || !body.Providers.Web.Configured || body.Providers.Web.Active != "claude" {
		t.Fatalf("unexpected web status: %#v", body.Providers.Web)
	}
	if body.Providers.ChatGPT.Count != 1 || !body.Providers.ChatGPT.Configured || body.Providers.ChatGPT.Active == "" {
		t.Fatalf("unexpected chatgpt status: %#v", body.Providers.ChatGPT)
	}
	if body.Providers.Blink.Count != 1 || !body.Providers.Blink.Configured || body.Providers.Blink.Active == "" {
		t.Fatalf("unexpected blink status: %#v", body.Providers.Blink)
	}
	if body.Providers.ZAIImage.Count != 1 || !body.Providers.ZAIImage.Configured || body.Providers.ZAIImage.Active == "" {
		t.Fatalf("unexpected zai image status: %#v", body.Providers.ZAIImage)
	}
	if body.Providers.ZAITTS.Count != 1 || !body.Providers.ZAITTS.Configured || body.Providers.ZAITTS.Active == "" {
		t.Fatalf("unexpected zai tts status: %#v", body.Providers.ZAITTS)
	}
	if body.Providers.ZAIOCR.Count != 1 || !body.Providers.ZAIOCR.Configured || body.Providers.ZAIOCR.Active == "" {
		t.Fatalf("unexpected zai ocr status: %#v", body.Providers.ZAIOCR)
	}

	grokConfigGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/grok/config", nil)
	grokConfigGetReq.AddCookie(cookie)
	grokConfigGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(grokConfigGetRec, grokConfigGetReq)
	if grokConfigGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected grok config get status: %d body=%s", grokConfigGetRec.Code, grokConfigGetRec.Body.String())
	}
	var grokConfigBody struct {
		Config struct {
			APIURL      string `json:"apiUrl"`
			ProxyURL    string `json:"proxyUrl"`
			CFClearance string `json:"cfClearance"`
		} `json:"config"`
	}
	if err := json.NewDecoder(grokConfigGetRec.Body).Decode(&grokConfigBody); err != nil {
		t.Fatalf("decode grok config response: %v", err)
	}
	if grokConfigBody.Config.APIURL != "https://grok.test/chat" || grokConfigBody.Config.ProxyURL != "http://127.0.0.1:7890" || grokConfigBody.Config.CFClearance != "cf-token" {
		t.Fatalf("unexpected grok config response: %#v", grokConfigBody.Config)
	}

	webConfigGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/web/config", nil)
	webConfigGetReq.AddCookie(cookie)
	webConfigGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(webConfigGetRec, webConfigGetReq)
	if webConfigGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected web config get status: %d body=%s", webConfigGetRec.Code, webConfigGetRec.Body.String())
	}
	var webConfigBody struct {
		Config struct {
			BaseURL string `json:"baseUrl"`
			Type    string `json:"type"`
			APIKey  string `json:"apiKey"`
		} `json:"config"`
	}
	if err := json.NewDecoder(webConfigGetRec.Body).Decode(&webConfigBody); err != nil {
		t.Fatalf("decode web config response: %v", err)
	}
	if webConfigBody.Config.BaseURL != "http://127.0.0.1:9000" || webConfigBody.Config.Type != "claude" || webConfigBody.Config.APIKey != "web-key" {
		t.Fatalf("unexpected web config response: %#v", webConfigBody.Config)
	}

	chatgptConfigGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/chatgpt/config", nil)
	chatgptConfigGetReq.AddCookie(cookie)
	chatgptConfigGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(chatgptConfigGetRec, chatgptConfigGetReq)
	if chatgptConfigGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected chatgpt config get status: %d body=%s", chatgptConfigGetRec.Code, chatgptConfigGetRec.Body.String())
	}
	var chatgptConfigBody struct {
		Config struct {
			BaseURL string `json:"baseUrl"`
			Token   string `json:"token"`
		} `json:"config"`
	}
	if err := json.NewDecoder(chatgptConfigGetRec.Body).Decode(&chatgptConfigBody); err != nil {
		t.Fatalf("decode chatgpt config response: %v", err)
	}
	if chatgptConfigBody.Config.BaseURL != "http://127.0.0.1:5005" || chatgptConfigBody.Config.Token != "chatgpt-token" {
		t.Fatalf("unexpected chatgpt config response: %#v", chatgptConfigBody.Config)
	}

	blinkConfigGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/blink/config", nil)
	blinkConfigGetReq.AddCookie(cookie)
	blinkConfigGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(blinkConfigGetRec, blinkConfigGetReq)
	if blinkConfigGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected blink config get status: %d body=%s", blinkConfigGetRec.Code, blinkConfigGetRec.Body.String())
	}
	var blinkConfigBody struct {
		Config struct {
			BaseURL            string `json:"baseUrl"`
			FirebaseRefreshURL string `json:"firebaseRefreshUrl"`
			RefreshToken       string `json:"refreshToken"`
			IDToken            string `json:"idToken"`
			SessionToken       string `json:"sessionToken"`
			WorkspaceSlug      string `json:"workspaceSlug"`
			ProjectID          string `json:"projectId"`
			ProxyURL           string `json:"proxyUrl"`
		} `json:"config"`
	}
	if err := json.NewDecoder(blinkConfigGetRec.Body).Decode(&blinkConfigBody); err != nil {
		t.Fatalf("decode blink config response: %v", err)
	}
	if blinkConfigBody.Config.BaseURL != "https://blink.test" || blinkConfigBody.Config.FirebaseRefreshURL != "https://firebase.test/token" || blinkConfigBody.Config.RefreshToken != "blink-refresh-token" || blinkConfigBody.Config.IDToken != "blink-id-token" || blinkConfigBody.Config.SessionToken != "blink-session-token" || blinkConfigBody.Config.WorkspaceSlug != "workspace-test" || blinkConfigBody.Config.ProjectID != "project-test" || blinkConfigBody.Config.ProxyURL != "http://127.0.0.1:7891" {
		t.Fatalf("unexpected blink config response: %#v", blinkConfigBody.Config)
	}

	zaiImageGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/zai/image/config", nil)
	zaiImageGetReq.AddCookie(cookie)
	zaiImageGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiImageGetRec, zaiImageGetReq)
	if zaiImageGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai image config get status: %d body=%s", zaiImageGetRec.Code, zaiImageGetRec.Body.String())
	}
	var zaiImageConfigBody struct {
		Config struct {
			SessionToken string `json:"sessionToken"`
			APIURL       string `json:"apiUrl"`
		} `json:"config"`
	}
	if err := json.NewDecoder(zaiImageGetRec.Body).Decode(&zaiImageConfigBody); err != nil {
		t.Fatalf("decode zai image config response: %v", err)
	}
	if zaiImageConfigBody.Config.SessionToken != "zai-image-session" || zaiImageConfigBody.Config.APIURL != "https://image.test/generate" {
		t.Fatalf("unexpected zai image config response: %#v", zaiImageConfigBody.Config)
	}

	zaiTTSGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/zai/tts/config", nil)
	zaiTTSGetReq.AddCookie(cookie)
	zaiTTSGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiTTSGetRec, zaiTTSGetReq)
	if zaiTTSGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai tts config get status: %d body=%s", zaiTTSGetRec.Code, zaiTTSGetRec.Body.String())
	}
	var zaiTTSConfigBody struct {
		Config struct {
			Token  string `json:"token"`
			UserID string `json:"userId"`
			APIURL string `json:"apiUrl"`
		} `json:"config"`
	}
	if err := json.NewDecoder(zaiTTSGetRec.Body).Decode(&zaiTTSConfigBody); err != nil {
		t.Fatalf("decode zai tts config response: %v", err)
	}
	if zaiTTSConfigBody.Config.Token != "zai-tts-token" || zaiTTSConfigBody.Config.UserID != "tts-user" || zaiTTSConfigBody.Config.APIURL != "https://audio.test/tts" {
		t.Fatalf("unexpected zai tts config response: %#v", zaiTTSConfigBody.Config)
	}

	zaiOCRGetReq := httptest.NewRequest(http.MethodGet, "/admin/api/providers/zai/ocr/config", nil)
	zaiOCRGetReq.AddCookie(cookie)
	zaiOCRGetRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(zaiOCRGetRec, zaiOCRGetReq)
	if zaiOCRGetRec.Code != http.StatusOK {
		t.Fatalf("unexpected zai ocr config get status: %d body=%s", zaiOCRGetRec.Code, zaiOCRGetRec.Body.String())
	}
	var zaiOCRConfigBody struct {
		Config struct {
			Token  string `json:"token"`
			APIURL string `json:"apiUrl"`
		} `json:"config"`
	}
	if err := json.NewDecoder(zaiOCRGetRec.Body).Decode(&zaiOCRConfigBody); err != nil {
		t.Fatalf("decode zai ocr config response: %v", err)
	}
	if zaiOCRConfigBody.Config.Token != "zai-ocr-token" || zaiOCRConfigBody.Config.APIURL != "https://ocr.test/process" {
		t.Fatalf("unexpected zai ocr config response: %#v", zaiOCRConfigBody.Config)
	}
}

func TestAdminStatusMarksBlinkConfiguredForDirectSessionConfig(t *testing.T) {
	h, cookie := newRuntimeAdminHandler(t)

	blinkPayload := []byte(`{"config":{"baseUrl":"https://blink.test","idToken":"blink-direct-id-token","sessionToken":"blink-direct-session-token","workspaceSlug":"workspace-direct","projectId":"project-direct"}}`)
	blinkReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/blink/config", bytes.NewReader(blinkPayload))
	blinkReq.AddCookie(cookie)
	blinkRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(blinkRec, blinkReq)
	if blinkRec.Code != http.StatusOK {
		t.Fatalf("unexpected blink direct config status: %d body=%s", blinkRec.Code, blinkRec.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/admin/api/status", nil)
	statusReq.AddCookie(cookie)
	statusRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("unexpected status response: %d body=%s", statusRec.Code, statusRec.Body.String())
	}

	var body struct {
		Providers struct {
			Blink struct {
				Count      int    `json:"count"`
				Configured bool   `json:"configured"`
				Active     string `json:"active"`
			} `json:"blink"`
		} `json:"providers"`
	}
	if err := json.NewDecoder(statusRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if body.Providers.Blink.Count != 1 || !body.Providers.Blink.Configured || body.Providers.Blink.Active != "workspace-direct" {
		t.Fatalf("unexpected blink direct status: %#v", body.Providers.Blink)
	}
}

func TestAdminBlinkCheckoutURLUsesOfficialBlinkCheckoutEndpoint(t *testing.T) {
	var blink *httptest.Server
	blink = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/auth/session-data":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-checkout-id-token" {
				t.Fatalf("unexpected blink checkout session-data auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-checkout-session-token") {
				t.Fatalf("unexpected blink checkout session-data cookie: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"active_workspace_id":"wsp_checkout"},"workspace":{"id":"wsp_checkout","slug":"workspace-checkout"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/stripe/checkout":
			if got := r.Header.Get("Authorization"); got != "Bearer blink-checkout-id-token" {
				t.Fatalf("unexpected blink checkout auth: %q", got)
			}
			if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=blink-checkout-session-token") || !strings.Contains(got, "workspace_slug=workspace-checkout") {
				t.Fatalf("unexpected blink checkout cookie: %q", got)
			}
			var body struct {
				PriceID     string  `json:"priceId"`
				PlanID      string  `json:"planId"`
				WorkspaceID string  `json:"workspaceId"`
				CancelURL   string  `json:"cancelUrl"`
				ToltRef     *string `json:"toltReferralId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode blink checkout body: %v", err)
			}
			if body.PriceID != "price_1S2oW1IChkSeVZoQl1420r64" {
				t.Fatalf("unexpected blink checkout price id: %q", body.PriceID)
			}
			if body.PlanID != "pro" || body.WorkspaceID != "wsp_checkout" {
				t.Fatalf("unexpected blink checkout payload: %#v", body)
			}
			if body.CancelURL != blink.URL+"/workspace-checkout?showPricing=true" {
				t.Fatalf("unexpected blink checkout cancel url: %q", body.CancelURL)
			}
			if body.ToltRef != nil {
				t.Fatalf("expected empty tolt referral id, got %#v", body.ToltRef)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sessionId":"cs_live_test","url":"https://checkout.stripe.com/c/pay/cs_live_test"}`))
		default:
			t.Fatalf("unexpected blink checkout request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer blink.Close()

	h, cookie := newRuntimeAdminHandler(t)

	blinkPayload := []byte(`{"config":{"baseUrl":"` + blink.URL + `","idToken":"blink-checkout-id-token","sessionToken":"blink-checkout-session-token"}}`)
	blinkReq := httptest.NewRequest(http.MethodPut, "/admin/api/providers/blink/config", bytes.NewReader(blinkPayload))
	blinkReq.AddCookie(cookie)
	blinkRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(blinkRec, blinkReq)
	if blinkRec.Code != http.StatusOK {
		t.Fatalf("unexpected blink config status: %d body=%s", blinkRec.Code, blinkRec.Body.String())
	}

	checkoutReq := httptest.NewRequest(http.MethodPost, "/admin/api/providers/blink/checkout-url", bytes.NewReader([]byte(`{}`)))
	checkoutReq.AddCookie(cookie)
	checkoutRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(checkoutRec, checkoutReq)
	if checkoutRec.Code != http.StatusOK {
		t.Fatalf("unexpected blink checkout status: %d body=%s", checkoutRec.Code, checkoutRec.Body.String())
	}

	var body struct {
		Data struct {
			URL           string `json:"url"`
			SessionID     string `json:"sessionId"`
			PlanID        string `json:"planId"`
			PriceID       string `json:"priceId"`
			WorkspaceID   string `json:"workspaceId"`
			WorkspaceSlug string `json:"workspaceSlug"`
		} `json:"data"`
	}
	if err := json.NewDecoder(checkoutRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode blink checkout response: %v", err)
	}
	if body.Data.URL != "https://checkout.stripe.com/c/pay/cs_live_test" || body.Data.SessionID != "cs_live_test" {
		t.Fatalf("unexpected blink checkout response data: %#v", body.Data)
	}
	if body.Data.PlanID != "pro" || body.Data.PriceID != "price_1S2oW1IChkSeVZoQl1420r64" {
		t.Fatalf("unexpected blink checkout plan response: %#v", body.Data)
	}
	if body.Data.WorkspaceID != "wsp_checkout" || body.Data.WorkspaceSlug != "workspace-checkout" {
		t.Fatalf("unexpected blink checkout workspace response: %#v", body.Data)
	}
}

func TestAdminSharedContractMetaEndpoint(t *testing.T) {
	h := newRuntimeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/meta", nil)
	req.Header.Set("Origin", "http://tauri.localhost")
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected meta status: %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://tauri.localhost" {
		t.Fatalf("expected cors origin to be echoed, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	var body struct {
		Backend struct {
			Language string `json:"language"`
			Version  string `json:"version"`
		} `json:"backend"`
		Auth struct {
			Mode string `json:"mode"`
		} `json:"auth"`
		Features struct {
			Providers          bool `json:"providers"`
			Credentials        bool `json:"credentials"`
			ProviderState      bool `json:"providerState"`
			Stats              bool `json:"stats"`
			Logs               bool `json:"logs"`
			Users              bool `json:"users"`
			ConfigImportExport bool `json:"configImportExport"`
		} `json:"features"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode meta response: %v", err)
	}
	if body.Backend.Language != "go" {
		t.Fatalf("expected backend language go, got %q", body.Backend.Language)
	}
	if strings.TrimSpace(body.Backend.Version) == "" {
		t.Fatal("expected backend version to be set")
	}
	if body.Auth.Mode != "session_cookie" {
		t.Fatalf("expected auth mode session_cookie, got %q", body.Auth.Mode)
	}
	if !body.Features.Providers || !body.Features.Credentials || !body.Features.ProviderState {
		t.Fatalf("expected core admin features enabled, got %#v", body.Features)
	}
	if body.Features.Stats || body.Features.Logs || body.Features.Users || body.Features.ConfigImportExport {
		t.Fatalf("expected optional features disabled by default, got %#v", body.Features)
	}
}

func TestAdminSharedContractSessionLifecycle(t *testing.T) {
	h := newRuntimeHandler(t)
	unauthorizedRec := httptest.NewRecorder()
	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/api/admin/auth/session", nil)
	h.Routes().ServeHTTP(unauthorizedRec, unauthorizedReq)
	if unauthorizedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated session check to return 401, got %d body=%s", unauthorizedRec.Code, unauthorizedRec.Body.String())
	}

	loginPayload := []byte(`{"password":"admin-secret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", bytes.NewReader(loginPayload))
	loginRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("unexpected contract login status: %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	var loginBody struct {
		OK    bool   `json:"ok"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(bytes.NewReader(loginRec.Body.Bytes())).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if !loginBody.OK || strings.TrimSpace(loginBody.Token) == "" {
		t.Fatalf("expected login token in response, got %#v", loginBody)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range loginRec.Result().Cookies() {
		if cookie.Name == adminSessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("missing session cookie after contract login")
	}

	sessionRec := httptest.NewRecorder()
	sessionReq := httptest.NewRequest(http.MethodGet, "/api/admin/auth/session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	h.Routes().ServeHTTP(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("unexpected contract session status: %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var sessionBody struct {
		Authenticated bool   `json:"authenticated"`
		ExpiresAt     string `json:"expiresAt"`
		User          struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Role string `json:"role"`
		} `json:"user"`
	}
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	if !sessionBody.Authenticated {
		t.Fatal("expected authenticated session")
	}
	if sessionBody.User.Role != "admin" || sessionBody.User.ID == "" || sessionBody.User.Name == "" {
		t.Fatalf("unexpected session user: %#v", sessionBody.User)
	}
	if strings.TrimSpace(sessionBody.ExpiresAt) == "" {
		t.Fatal("expected session expiry to be present")
	}
	if sessionCookie == nil {
		t.Fatal("expected cookie login to remain available for web admin")
	}
}

func TestAdminSharedContractLogoutInvalidatesSession(t *testing.T) {
	h := newRuntimeHandler(t)
	loginPayload := []byte(`{"password":"admin-secret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", bytes.NewReader(loginPayload))
	loginRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("unexpected contract login status: %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(bytes.NewReader(loginRec.Body.Bytes())).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if strings.TrimSpace(loginBody.Token) == "" {
		t.Fatal("expected bearer token from login response")
	}
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/admin/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	logoutRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("unexpected contract logout status: %d body=%s", logoutRec.Code, logoutRec.Body.String())
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/admin/auth/session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	sessionRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected logged out session to return 401, got %d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
}

func newRuntimeHandler(t *testing.T) *Handler {
	t.Helper()
	cfg := testAppConfig()
	cfg.AdminPassword = "admin-secret"
	runtime, err := core.NewRuntimeManager(filepath.Join(t.TempDir(), "admin.json"), cfg)
	if err != nil {
		t.Fatalf("new runtime manager: %v", err)
	}
	return NewHandlerWithRuntime(runtime)
}

func newRuntimeAdminHandler(t *testing.T) (*Handler, *http.Cookie) {
	t.Helper()
	h := newRuntimeHandler(t)
	loginPayload := []byte(`{"password":"admin-secret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader(loginPayload))
	loginRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("unexpected admin login status: %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	res := loginRec.Result()
	defer res.Body.Close()
	for _, cookie := range res.Cookies() {
		if cookie.Name == adminSessionCookieName {
			return h, cookie
		}
	}
	t.Fatal("missing admin session cookie")
	return nil, nil
}

type testAnthropicOnlyProvider struct{}

func (testAnthropicOnlyProvider) ID() string { return "anthropic-only" }

func (testAnthropicOnlyProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{AnthropicCompatible: true}
}

func (testAnthropicOnlyProvider) Models() []core.ModelInfo {
	return []core.ModelInfo{{Provider: "anthropic-only", PublicModel: "stub-model", UpstreamModel: "stub-model", OwnedBy: "tests"}}
}

func (testAnthropicOnlyProvider) BuildUpstreamPreview(core.UnifiedRequest) map[string]interface{} {
	return map[string]interface{}{"live_enabled": false}
}

func (testAnthropicOnlyProvider) GenerateReply(core.UnifiedRequest) string { return "stub" }
