package web

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"any2api-go/internal/core"
)

type webProvider struct {
	client        *http.Client
	requestConfig core.RequestConfig
	baseURL       string
	typeName      string
	apiKey        string
}

type webChatRequest struct {
	Model    string         `json:"model"`
	Messages []core.Message `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
}

type webChatResponse struct {
	Choices []struct {
		Message struct {
			Content interface{} `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error interface{} `json:"error,omitempty"`
}

type webStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content interface{} `json:"content"`
		} `json:"delta"`
		Message struct {
			Content interface{} `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error interface{} `json:"error,omitempty"`
}

func NewProvider() core.Provider {
	return NewProviderWithConfig(core.WebConfig{})
}

func NewProviderWithConfig(cfg core.WebConfig) core.Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = core.DefaultWebBaseURL
	}
	if cfg.Type == "" {
		cfg.Type = core.DefaultWebTypeName
	}
	if cfg.Request.Timeout <= 0 {
		cfg.Request.Timeout = time.Duration(core.DefaultCursorTimeoutSeconds) * time.Second
	}
	if cfg.Request.MaxInputLength <= 0 {
		cfg.Request.MaxInputLength = core.DefaultCursorMaxInputLength
	}
	return &webProvider{
		client:        &http.Client{Timeout: cfg.Request.Timeout},
		requestConfig: cfg.Request,
		baseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		typeName:      strings.Trim(strings.TrimSpace(cfg.Type), "/"),
		apiKey:        strings.TrimSpace(cfg.APIKey),
	}
}

func (*webProvider) ID() string { return "web" }

func (*webProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{OpenAICompatible: true}
}

func (p *webProvider) Models() []core.ModelInfo {
	model := p.defaultModel()
	owner := p.typeName
	if owner == "" {
		owner = "web"
	}
	return []core.ModelInfo{{Provider: "web", PublicModel: model, UpstreamModel: model, OwnedBy: owner}}
}

func (p *webProvider) BuildUpstreamPreview(req core.UnifiedRequest) map[string]interface{} {
	return map[string]interface{}{
		"url":           p.chatURL(),
		"auth":          "bearer api key (optional)",
		"live_enabled":  true,
		"configured":    p.baseURL != "" && p.typeName != "",
		"type":          p.typeName,
		"api_key_set":   p.apiKey != "",
		"mapped_model":  p.mapModel(req.Model),
		"message_count": len(req.Messages),
	}
}

func (*webProvider) GenerateReply(req core.UnifiedRequest) string {
	if req.Model == "" {
		return "[web provider] mapped request to web upstream"
	}
	return fmt.Sprintf("[web provider] mapped request to web upstream for model=%s", req.Model)
}

func (p *webProvider) CompleteOpenAI(ctx context.Context, req core.UnifiedRequest) (string, error) {
	resp, err := p.doRequest(ctx, req, false)
	if err != nil {
		return "", err
	}
	if isWebSSE(resp) {
		output := make(chan core.TextStreamEvent, 32)
		go p.consumeStream(ctx, resp.Body, output)
		return core.CollectTextStream(ctx, output)
	}
	defer resp.Body.Close()
	return parseWebChatResponse(resp.Body)
}

func (p *webProvider) StreamOpenAI(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	resp, err := p.doRequest(ctx, req, true)
	if err != nil {
		return nil, err
	}
	output := make(chan core.TextStreamEvent, 32)
	if isWebSSE(resp) {
		go p.consumeStream(ctx, resp.Body, output)
		return output, nil
	}
	defer resp.Body.Close()
	text, err := parseWebChatResponse(resp.Body)
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

func (p *webProvider) doRequest(ctx context.Context, req core.UnifiedRequest, stream bool) (*http.Response, error) {
	if p.baseURL == "" {
		return nil, fmt.Errorf("web base url is not configured")
	}
	if p.typeName == "" {
		return nil, fmt.Errorf("web type is not configured")
	}
	body, err := json.Marshal(p.buildChatRequest(req, stream))
	if err != nil {
		return nil, fmt.Errorf("marshal web payload: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.chatURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build web request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("web upstream request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("web upstream error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func (p *webProvider) buildChatRequest(req core.UnifiedRequest, stream bool) webChatRequest {
	return webChatRequest{
		Model:    p.mapModel(req.Model),
		Messages: core.NormalizeMessages(req, p.requestConfig.SystemPromptInject, p.requestConfig.MaxInputLength),
		Stream:   stream,
	}
}

func (p *webProvider) chatURL() string {
	return fmt.Sprintf("%s/%s/v1/chat/completions", p.baseURL, url.PathEscape(p.typeName))
}

func (p *webProvider) mapModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return p.defaultModel()
	}
	return strings.TrimSpace(model)
}

func (p *webProvider) defaultModel() string {
	switch strings.ToLower(strings.TrimSpace(p.typeName)) {
	case "claude":
		return "claude-sonnet-4.5"
	case "openai":
		return "gpt-4.1"
	default:
		if p.typeName != "" {
			return p.typeName
		}
		return "claude-sonnet-4.5"
	}
}

func (p *webProvider) consumeStream(ctx context.Context, body io.ReadCloser, output chan<- core.TextStreamEvent) {
	defer body.Close()
	defer close(output)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			return
		}
		delta, err := parseWebStreamPayload(payload)
		if err != nil {
			emitWebEvent(ctx, output, core.TextStreamEvent{Err: err})
			return
		}
		if delta == "" {
			continue
		}
		if !emitWebEvent(ctx, output, core.TextStreamEvent{Delta: delta}) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		emitWebEvent(ctx, output, core.TextStreamEvent{Err: fmt.Errorf("read web sse response: %w", err)})
	}
}

func parseWebChatResponse(body io.Reader) (string, error) {
	var resp webChatResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return "", fmt.Errorf("decode web response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("web upstream error: %s", describeWebError(resp.Error))
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("web upstream returned no choices")
	}
	text := core.ContentText(resp.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("web upstream returned empty content")
	}
	return text, nil
}

func parseWebStreamPayload(payload string) (string, error) {
	var resp webStreamResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		return "", fmt.Errorf("decode web stream payload: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("web upstream error: %s", describeWebError(resp.Error))
	}
	var text strings.Builder
	for _, choice := range resp.Choices {
		text.WriteString(core.ContentText(choice.Delta.Content))
		if choice.Message.Content != nil {
			text.WriteString(core.ContentText(choice.Message.Content))
		}
	}
	return text.String(), nil
}

func describeWebError(value interface{}) string {
	if value == nil {
		return "unknown error"
	}
	if msg, ok := value.(string); ok && strings.TrimSpace(msg) != "" {
		return msg
	}
	if m, ok := value.(map[string]interface{}); ok {
		if msg, _ := m["message"].(string); strings.TrimSpace(msg) != "" {
			return msg
		}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "unknown error"
	}
	return string(encoded)
}

func isWebSSE(resp *http.Response) bool {
	return strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream")
}

func emitWebEvent(ctx context.Context, output chan<- core.TextStreamEvent, event core.TextStreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case output <- event:
		return true
	}
}
