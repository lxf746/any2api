package chatgpt

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

type chatgptProvider struct {
	client        *http.Client
	requestConfig core.RequestConfig
	baseURL       string
	token         string
}

type chatgptChatRequest struct {
	Model    string         `json:"model"`
	Messages []core.Message `json:"messages"`
	Stream   bool           `json:"stream,omitempty"`
}

type chatgptChatResponse struct {
	Choices []struct {
		Message struct {
			Content interface{} `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error interface{} `json:"error,omitempty"`
}

type chatgptStreamResponse struct {
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
	return NewProviderWithConfig(core.ChatGPTConfig{})
}

func NewProviderWithConfig(cfg core.ChatGPTConfig) core.Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = core.DefaultChatGPTBaseURL
	}
	if cfg.Request.Timeout <= 0 {
		cfg.Request.Timeout = time.Duration(core.DefaultCursorTimeoutSeconds) * time.Second
	}
	if cfg.Request.MaxInputLength <= 0 {
		cfg.Request.MaxInputLength = core.DefaultCursorMaxInputLength
	}
	return &chatgptProvider{
		client:        &http.Client{Timeout: cfg.Request.Timeout},
		requestConfig: cfg.Request,
		baseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		token:         strings.TrimSpace(cfg.Token),
	}
}

func (*chatgptProvider) ID() string { return "chatgpt" }

func (*chatgptProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{OpenAICompatible: true}
}

func (*chatgptProvider) Models() []core.ModelInfo {
	return []core.ModelInfo{{Provider: "chatgpt", PublicModel: "gpt-4.1", UpstreamModel: "gpt-4.1", OwnedBy: "openai"}}
}

func (p *chatgptProvider) BuildUpstreamPreview(req core.UnifiedRequest) map[string]interface{} {
	return map[string]interface{}{
		"url":           p.chatURL(),
		"auth":          "bearer token",
		"live_enabled":  true,
		"configured":    p.baseURL != "" && p.token != "",
		"token_set":     p.token != "",
		"mapped_model":  p.mapModel(req.Model),
		"message_count": len(req.Messages),
	}
}

func (*chatgptProvider) GenerateReply(req core.UnifiedRequest) string {
	if req.Model == "" {
		return "[chatgpt provider] mapped request to ChatGPT upstream"
	}
	return fmt.Sprintf("[chatgpt provider] mapped request to ChatGPT upstream for model=%s", req.Model)
}

func (p *chatgptProvider) CompleteOpenAI(ctx context.Context, req core.UnifiedRequest) (string, error) {
	resp, err := p.doRequest(ctx, req, false)
	if err != nil {
		return "", err
	}
	if isChatGPTSSE(resp) {
		output := make(chan core.TextStreamEvent, 32)
		go p.consumeStream(ctx, resp.Body, output)
		return core.CollectTextStream(ctx, output)
	}
	defer resp.Body.Close()
	return parseChatGPTChatResponse(resp.Body)
}

func (p *chatgptProvider) StreamOpenAI(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	resp, err := p.doRequest(ctx, req, true)
	if err != nil {
		return nil, err
	}
	output := make(chan core.TextStreamEvent, 32)
	if isChatGPTSSE(resp) {
		go p.consumeStream(ctx, resp.Body, output)
		return output, nil
	}
	defer resp.Body.Close()
	text, err := parseChatGPTChatResponse(resp.Body)
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

func (p *chatgptProvider) doRequest(ctx context.Context, req core.UnifiedRequest, stream bool) (*http.Response, error) {
	if p.baseURL == "" {
		return nil, fmt.Errorf("chatgpt base url is not configured")
	}
	if p.token == "" {
		return nil, fmt.Errorf("chatgpt token is not configured")
	}
	body, err := json.Marshal(p.buildChatRequest(req, stream))
	if err != nil {
		return nil, fmt.Errorf("marshal chatgpt payload: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.chatURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build chatgpt request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.token)
	httpReq.Header.Set("Content-Type", "application/json")
	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	} else {
		httpReq.Header.Set("Accept", "application/json")
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chatgpt upstream request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("chatgpt upstream error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func (p *chatgptProvider) buildChatRequest(req core.UnifiedRequest, stream bool) chatgptChatRequest {
	return chatgptChatRequest{
		Model:    p.mapModel(req.Model),
		Messages: core.NormalizeMessages(req, p.requestConfig.SystemPromptInject, p.requestConfig.MaxInputLength),
		Stream:   stream,
	}
}

func (p *chatgptProvider) chatURL() string {
	return p.baseURL + "/v1/chat/completions"
}

func (p *chatgptProvider) mapModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return p.Models()[0].UpstreamModel
	}
	return strings.TrimSpace(model)
}

func (p *chatgptProvider) consumeStream(ctx context.Context, body io.ReadCloser, output chan<- core.TextStreamEvent) {
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
		delta, err := parseChatGPTStreamPayload(payload)
		if err != nil {
			emitChatGPTEvent(ctx, output, core.TextStreamEvent{Err: err})
			return
		}
		if delta == "" {
			continue
		}
		if !emitChatGPTEvent(ctx, output, core.TextStreamEvent{Delta: delta}) {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		emitChatGPTEvent(ctx, output, core.TextStreamEvent{Err: fmt.Errorf("read chatgpt sse response: %w", err)})
	}
}

func parseChatGPTChatResponse(body io.Reader) (string, error) {
	var resp chatgptChatResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return "", fmt.Errorf("decode chatgpt response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("chatgpt upstream error: %s", describeChatGPTError(resp.Error))
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("chatgpt upstream returned no choices")
	}
	text := core.ContentText(resp.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("chatgpt upstream returned empty content")
	}
	return text, nil
}

func parseChatGPTStreamPayload(payload string) (string, error) {
	var resp chatgptStreamResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		return "", fmt.Errorf("decode chatgpt stream payload: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("chatgpt upstream error: %s", describeChatGPTError(resp.Error))
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

func describeChatGPTError(value interface{}) string {
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

func isChatGPTSSE(resp *http.Response) bool {
	return strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream")
}

func emitChatGPTEvent(ctx context.Context, output chan<- core.TextStreamEvent, event core.TextStreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case output <- event:
		return true
	}
}
