package zai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"any2api-go/internal/core"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"

type zaiProvider struct {
	client        *http.Client
	requestConfig core.RequestConfig
	authURL       string
	apiURL        string
	token         string
	feVersion     string
}

type zaiChatRequest struct {
	Model    string         `json:"model"`
	Messages []core.Message `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
}

type zaiChatResponse struct {
	Choices []struct {
		Message struct {
			Content interface{} `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error interface{} `json:"error,omitempty"`
}

type zaiStreamResponse struct {
	Type string `json:"type"`
	Data struct {
		DeltaContent string `json:"delta_content"`
		Done         bool   `json:"done"`
	} `json:"data"`
}

func NewProvider() core.Provider {
	return NewProviderWithConfig(core.ZAIConfig{})
}

func NewProviderWithConfig(cfg core.ZAIConfig) core.Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = core.DefaultZAIBaseURL
	}
	if cfg.AuthURL == "" {
		cfg.AuthURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + "/api/v1/auths/"
	}
	if cfg.APIURL == "" {
		cfg.APIURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + "/api/chat/completions"
	}
	if cfg.FEVersion == "" {
		cfg.FEVersion = core.DefaultZAIFEVersion
	}
	if cfg.Request.Timeout <= 0 {
		cfg.Request.Timeout = time.Duration(core.DefaultCursorTimeoutSeconds) * time.Second
	}
	if cfg.Request.MaxInputLength <= 0 {
		cfg.Request.MaxInputLength = core.DefaultCursorMaxInputLength
	}
	return &zaiProvider{
		client:        &http.Client{Timeout: cfg.Request.Timeout},
		requestConfig: cfg.Request,
		authURL:       strings.TrimSpace(cfg.AuthURL),
		apiURL:        strings.TrimSpace(cfg.APIURL),
		token:         strings.TrimSpace(cfg.Token),
		feVersion:     strings.TrimSpace(cfg.FEVersion),
	}
}

func (*zaiProvider) ID() string { return "zai" }

func (*zaiProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{OpenAICompatible: true}
}

func (*zaiProvider) Models() []core.ModelInfo {
	return []core.ModelInfo{{Provider: "zai", PublicModel: "glm-4.5-flash", UpstreamModel: "glm-4.5-flash", OwnedBy: "z.ai"}}
}

func (p *zaiProvider) BuildUpstreamPreview(req core.UnifiedRequest) map[string]interface{} {
	return map[string]interface{}{
		"url":           p.apiURL,
		"auth":          "anonymous token or fixed bearer token",
		"live_enabled":  true,
		"configured":    p.apiURL != "" && (p.authURL != "" || p.token != ""),
		"token_set":     p.token != "",
		"mapped_model":  p.mapModel(req.Model),
		"message_count": len(req.Messages),
	}
}

func (*zaiProvider) GenerateReply(req core.UnifiedRequest) string {
	if req.Model == "" {
		return "[zai provider] mapped request to Z.ai upstream"
	}
	return fmt.Sprintf("[zai provider] mapped request to Z.ai upstream for model=%s", req.Model)
}

func (p *zaiProvider) CompleteOpenAI(ctx context.Context, req core.UnifiedRequest) (string, error) {
	resp, err := p.doRequest(ctx, req, false)
	if err != nil {
		return "", err
	}
	if isZAISSE(resp) {
		output := make(chan core.TextStreamEvent, 32)
		go p.consumeStream(ctx, resp.Body, output)
		return core.CollectTextStream(ctx, output)
	}
	defer resp.Body.Close()
	return parseZAIChatResponse(resp.Body)
}

func (p *zaiProvider) StreamOpenAI(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	resp, err := p.doRequest(ctx, req, true)
	if err != nil {
		return nil, err
	}
	output := make(chan core.TextStreamEvent, 32)
	if isZAISSE(resp) {
		go p.consumeStream(ctx, resp.Body, output)
		return output, nil
	}
	defer resp.Body.Close()
	text, err := parseZAIChatResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(output)
		select {
		case <-ctx.Done():
			output <- core.TextStreamEvent{Err: ctx.Err()}
		case output <- core.TextStreamEvent{Delta: text}:
		}
	}()
	return output, nil
}

func (p *zaiProvider) doRequest(ctx context.Context, req core.UnifiedRequest, stream bool) (*http.Response, error) {
	if p.apiURL == "" {
		return nil, fmt.Errorf("zai api url is not configured")
	}
	token, err := p.resolveToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(zaiChatRequest{
		Model:    p.mapModel(req.Model),
		Messages: core.NormalizeMessages(req, p.requestConfig.SystemPromptInject, p.requestConfig.MaxInputLength),
		Stream:   stream,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal zai payload: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build zai request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", defaultUserAgent)
	httpReq.Header.Set("Origin", "https://chat.z.ai")
	httpReq.Header.Set("Referer", "https://chat.z.ai/")
	if p.feVersion != "" {
		httpReq.Header.Set("X-FE-Version", p.feVersion)
	}
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json, text/event-stream")
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("zai upstream request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("zai upstream error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func (p *zaiProvider) resolveToken(ctx context.Context) (string, error) {
	if p.token != "" {
		return p.token, nil
	}
	if p.authURL == "" {
		return "", fmt.Errorf("zai auth url is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.authURL, nil)
	if err != nil {
		return "", fmt.Errorf("build zai auth request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("zai auth request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("zai auth error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode zai auth response: %w", err)
	}
	if strings.TrimSpace(payload.Token) == "" {
		return "", fmt.Errorf("zai auth returned empty token")
	}
	return strings.TrimSpace(payload.Token), nil
}

func (p *zaiProvider) mapModel(model string) string {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "", "glm-4.5", "glm-4.5-flash", "gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo":
		return "glm-4.5-flash"
	default:
		return strings.TrimSpace(model)
	}
}

func (p *zaiProvider) consumeStream(ctx context.Context, body io.ReadCloser, output chan<- core.TextStreamEvent) {
	defer body.Close()
	defer close(output)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			if payload == "[DONE]" {
				return
			}
			continue
		}
		delta, done, err := parseZAIStreamPayload(payload)
		if err != nil {
			emitZAIEvent(ctx, output, core.TextStreamEvent{Err: err})
			return
		}
		if done {
			continue
		}
		if delta == "" {
			continue
		}
		if !emitZAIEvent(ctx, output, core.TextStreamEvent{Delta: delta}) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		emitZAIEvent(ctx, output, core.TextStreamEvent{Err: fmt.Errorf("read zai sse response: %w", err)})
	}
}

func parseZAIChatResponse(body io.Reader) (string, error) {
	var resp zaiChatResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return "", fmt.Errorf("decode zai response: %w", err)
	}
	if resp.Error != nil {
		encoded, _ := json.Marshal(resp.Error)
		return "", fmt.Errorf("zai upstream error: %s", string(encoded))
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("zai upstream returned no choices")
	}
	text := core.ContentText(resp.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("zai upstream returned empty content")
	}
	return text, nil
}

func parseZAIStreamPayload(payload string) (string, bool, error) {
	var resp zaiStreamResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		return "", false, fmt.Errorf("decode zai stream payload: %w", err)
	}
	return resp.Data.DeltaContent, resp.Data.Done, nil
}

func isZAISSE(resp *http.Response) bool {
	return strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream")
}

func emitZAIEvent(ctx context.Context, output chan<- core.TextStreamEvent, event core.TextStreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case output <- event:
		return true
	}
}
