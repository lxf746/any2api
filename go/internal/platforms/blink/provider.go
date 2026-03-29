package blink

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"any2api-go/internal/core"
)

const (
	blinkUserAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
	blinkDefaultReferer    = "/"
	blinkProjectVisibility = "public"
	blinkProjectAppType    = "full_stack"
	blinkProPriceID        = "price_1S2oW1IChkSeVZoQl1420r64"
)

type blinkProvider struct {
	mu            sync.RWMutex
	client        *http.Client
	requestConfig core.RequestConfig
	baseURL       string
	refreshURL    string
	refreshToken  string
	idToken       string
	sessionToken  string
	workspaceSlug string
	projectID     string
	proxyURL      string
	cachedAuth    *blinkAuthState
}

type blinkAuthState struct {
	RefreshToken  string
	IDToken       string
	IDTokenUntil  time.Time
	SessionToken  string
	WorkspaceSlug string
	WorkspaceID   string
}

type blinkChatRequest struct {
	ID              string       `json:"id"`
	ProjectID       string       `json:"projectId"`
	Mode            string       `json:"mode"`
	Message         blinkMessage `json:"message"`
	ModelID         string       `json:"modelId"`
	ThinkingEnabled bool         `json:"thinkingEnabled"`
}

type blinkMessage struct {
	Role  string      `json:"role"`
	Parts []blinkPart `json:"parts"`
	ID    string      `json:"id"`
}

type blinkPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type blinkProjectCreateRequest struct {
	Prompt     string `json:"prompt"`
	Visibility string `json:"visibility"`
	AppType    string `json:"app_type"`
}

type blinkProjectCreateResponse struct {
	ID string `json:"id"`
}

type blinkCheckoutRequest struct {
	PriceID        string  `json:"priceId"`
	PlanID         string  `json:"planId"`
	ToltReferralID *string `json:"toltReferralId"`
	WorkspaceID    string  `json:"workspaceId"`
	CancelURL      string  `json:"cancelUrl"`
}

type blinkCheckoutResponse struct {
	SessionID string `json:"sessionId"`
	URL       string `json:"url"`
}

type blinkFirebaseRefreshResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

type blinkSessionDataResponse struct {
	User struct {
		ActiveWorkspaceID string `json:"active_workspace_id"`
	} `json:"user"`
	Workspace struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	} `json:"workspace"`
}

type blinkStreamEvent struct {
	Type  string `json:"type"`
	Delta string `json:"delta"`
	Error string `json:"error"`
}

type blinkRequestOptions struct {
	RefreshToken  string
	IDToken       string
	SessionToken  string
	WorkspaceSlug string
	ProjectID     string
}

type CheckoutRequest struct {
	PlanID         string
	PriceID        string
	WorkspaceID    string
	CancelURL      string
	ToltReferralID string
}

type CheckoutResult struct {
	SessionID     string
	URL           string
	PlanID        string
	PriceID       string
	WorkspaceID   string
	WorkspaceSlug string
}

func NewProvider() core.Provider {
	return NewProviderWithConfig(core.BlinkConfig{})
}

func NewProviderWithConfig(cfg core.BlinkConfig) core.Provider {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = core.DefaultBlinkBaseURL
	}
	if strings.TrimSpace(cfg.FirebaseRefreshURL) == "" {
		cfg.FirebaseRefreshURL = core.DefaultBlinkRefreshURL
	}
	if cfg.Request.Timeout <= 0 {
		cfg.Request.Timeout = 300 * time.Second
	}
	if cfg.Request.MaxInputLength <= 0 {
		cfg.Request.MaxInputLength = core.DefaultCursorMaxInputLength
	}

	transport := &http.Transport{}
	if proxyURL := strings.TrimSpace(cfg.ProxyURL); proxyURL != "" {
		if parsed, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(parsed)
		}
	}

	return &blinkProvider{
		client:        &http.Client{Timeout: cfg.Request.Timeout, Transport: transport},
		requestConfig: cfg.Request,
		baseURL:       strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		refreshURL:    strings.TrimSpace(cfg.FirebaseRefreshURL),
		refreshToken:  strings.TrimSpace(cfg.RefreshToken),
		idToken:       strings.TrimSpace(cfg.IDToken),
		sessionToken:  strings.TrimSpace(cfg.SessionToken),
		workspaceSlug: strings.TrimSpace(cfg.WorkspaceSlug),
		projectID:     strings.TrimSpace(cfg.ProjectID),
		proxyURL:      strings.TrimSpace(cfg.ProxyURL),
	}
}

func CreateCheckout(ctx context.Context, cfg core.BlinkConfig, req CheckoutRequest) (*CheckoutResult, error) {
	provider, ok := NewProviderWithConfig(cfg).(*blinkProvider)
	if !ok {
		return nil, fmt.Errorf("blink provider initialization failed")
	}
	options := provider.effectiveRequestOptions(blinkRequestOptions{})
	authState, err := provider.resolveAuthState(ctx, blinkRequestOptions{}, options)
	if err != nil {
		return nil, err
	}

	planID := normalizeBlinkPlanID(req.PlanID)
	priceID := strings.TrimSpace(req.PriceID)
	if priceID == "" {
		priceID = defaultBlinkPriceID(planID)
	}
	if priceID == "" {
		return nil, fmt.Errorf("unsupported blink plan: %s", planID)
	}

	workspaceID := firstNonEmptyString(strings.TrimSpace(req.WorkspaceID), strings.TrimSpace(authState.WorkspaceID))
	if workspaceID == "" {
		return nil, fmt.Errorf("blink session data missing workspace id")
	}

	cancelURL := strings.TrimSpace(req.CancelURL)
	if cancelURL == "" {
		if strings.TrimSpace(authState.WorkspaceSlug) != "" {
			cancelURL = fmt.Sprintf("%s/%s?showPricing=true", provider.baseURL, strings.TrimSpace(authState.WorkspaceSlug))
		} else {
			cancelURL = provider.baseURL + "/?showPricing=true"
		}
	}

	checkout, err := provider.createCheckout(ctx, authState, blinkCheckoutRequest{
		PriceID:        priceID,
		PlanID:         planID,
		ToltReferralID: optionalStringPtr(req.ToltReferralID),
		WorkspaceID:    workspaceID,
		CancelURL:      cancelURL,
	})
	if err != nil {
		return nil, err
	}

	return &CheckoutResult{
		SessionID:     strings.TrimSpace(checkout.SessionID),
		URL:           strings.TrimSpace(checkout.URL),
		PlanID:        planID,
		PriceID:       priceID,
		WorkspaceID:   workspaceID,
		WorkspaceSlug: strings.TrimSpace(authState.WorkspaceSlug),
	}, nil
}

func (*blinkProvider) ID() string { return "blink" }

func (*blinkProvider) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{
		OpenAICompatible:    true,
		AnthropicCompatible: true,
		Tools:               false,
		Images:              false,
		MultiAccount:        false,
	}
}

func (*blinkProvider) Models() []core.ModelInfo {
	return []core.ModelInfo{
		{Provider: "blink", PublicModel: "claude-sonnet-4.6", UpstreamModel: "anthropic/claude-sonnet-4.6", OwnedBy: "blink"},
		{Provider: "blink", PublicModel: "claude-opus-4.6", UpstreamModel: "anthropic/claude-opus-4.6", OwnedBy: "blink"},
		{Provider: "blink", PublicModel: "claude-opus-4", UpstreamModel: "anthropic/claude-opus-4", OwnedBy: "blink"},
		{Provider: "blink", PublicModel: "gpt-4o", UpstreamModel: "openai/gpt-4o", OwnedBy: "blink"},
		{Provider: "blink", PublicModel: "gpt-4.1", UpstreamModel: "openai/gpt-4.1", OwnedBy: "blink"},
	}
}

func (p *blinkProvider) BuildUpstreamPreview(req core.UnifiedRequest) map[string]interface{} {
	options := p.effectiveRequestOptions(p.requestOptions(req))
	authMode := "missing"
	switch {
	case options.IDToken != "" && options.SessionToken != "":
		authMode = "direct_session"
	case firstNonEmptyString(options.RefreshToken, p.refreshToken) != "":
		authMode = "refresh_token"
	}
	return map[string]interface{}{
		"base_url":             p.baseURL,
		"chat_url":             p.chatURL(),
		"project_url":          p.projectCreateURL(),
		"firebase_refresh_url": p.refreshURL,
		"mapped_model":         mapBlinkModel(req.Model),
		"auth_mode":            authMode,
		"refresh_token_set":    firstNonEmptyString(options.RefreshToken, p.refreshToken) != "",
		"id_token_set":         options.IDToken != "",
		"session_token_set":    options.SessionToken != "",
		"workspace_slug":       options.WorkspaceSlug,
		"project_override":     options.ProjectID,
		"message_count":        len(req.Messages),
	}
}

func (*blinkProvider) GenerateReply(req core.UnifiedRequest) string {
	if strings.TrimSpace(req.Model) == "" {
		return "[blink provider] mapped request to Blink upstream"
	}
	return fmt.Sprintf("[blink provider] mapped request to Blink upstream for model=%s", strings.TrimSpace(req.Model))
}

func (p *blinkProvider) CompleteOpenAI(ctx context.Context, req core.UnifiedRequest) (string, error) {
	return core.CollectTextStream(ctx, p.mustStream(ctx, req))
}

func (p *blinkProvider) StreamOpenAI(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	return p.stream(ctx, req)
}

func (p *blinkProvider) CompleteAnthropic(ctx context.Context, req core.UnifiedRequest) (string, error) {
	return core.CollectTextStream(ctx, p.mustStream(ctx, req))
}

func (p *blinkProvider) StreamAnthropic(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	return p.stream(ctx, req)
}

func (p *blinkProvider) mustStream(ctx context.Context, req core.UnifiedRequest) <-chan core.TextStreamEvent {
	events, err := p.stream(ctx, req)
	if err == nil {
		return events
	}
	output := make(chan core.TextStreamEvent, 1)
	go func() {
		defer close(output)
		select {
		case <-ctx.Done():
			output <- core.TextStreamEvent{Err: ctx.Err()}
		case output <- core.TextStreamEvent{Err: err}:
		}
	}()
	return output
}

func (p *blinkProvider) stream(ctx context.Context, req core.UnifiedRequest) (<-chan core.TextStreamEvent, error) {
	rawOptions := p.requestOptions(req)
	options := p.effectiveRequestOptions(rawOptions)
	authState, err := p.resolveAuthState(ctx, rawOptions, options)
	if err != nil {
		return nil, err
	}

	projectPayload, chatText := buildBlinkPrompt(core.NormalizeMessages(req, p.requestConfig.SystemPromptInject, p.requestConfig.MaxInputLength))
	projectID := strings.TrimSpace(options.ProjectID)
	if projectID == "" {
		projectID, err = p.createProject(ctx, authState, projectPayload)
		if err != nil {
			return nil, fmt.Errorf("blink create project: %w", err)
		}
	}

	resp, err := p.sendChat(ctx, req, authState, projectID, chatText)
	if err != nil {
		return nil, err
	}

	output := make(chan core.TextStreamEvent, 32)
	go p.consumeSSE(ctx, resp.Body, output)
	return output, nil
}

func (p *blinkProvider) resolveAuthState(ctx context.Context, rawOptions blinkRequestOptions, options blinkRequestOptions) (*blinkAuthState, error) {
	if options.IDToken != "" || options.SessionToken != "" {
		if options.IDToken != "" && options.SessionToken != "" {
			return p.resolveDirectAuthState(ctx, options)
		}
		if strings.TrimSpace(options.RefreshToken) == "" {
			return nil, fmt.Errorf("blink direct session auth requires both id token and session token")
		}
	}

	refreshToken := strings.TrimSpace(options.RefreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("blink credentials are not configured: either refresh token or id token + session token is required")
	}

	cacheable := strings.TrimSpace(rawOptions.RefreshToken) == ""
	if cacheable {
		p.mu.RLock()
		if p.cachedAuth != nil && p.cachedAuth.IDToken != "" && time.Now().Before(p.cachedAuth.IDTokenUntil) {
			cached := *p.cachedAuth
			p.mu.RUnlock()
			return &cached, nil
		}
		p.mu.RUnlock()
	}

	idToken, nextRefreshToken, expiresAt, err := p.refreshIDToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("blink refresh firebase token: %w", err)
	}
	sessionToken, err := p.createSession(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("blink create session: %w", err)
	}
	sessionData, err := p.fetchSessionData(ctx, idToken, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("blink fetch session data: %w", err)
	}

	state := &blinkAuthState{
		RefreshToken:  firstNonEmptyString(nextRefreshToken, refreshToken),
		IDToken:       idToken,
		IDTokenUntil:  expiresAt,
		SessionToken:  sessionToken,
		WorkspaceSlug: strings.TrimSpace(sessionData.Workspace.Slug),
		WorkspaceID:   firstNonEmptyString(strings.TrimSpace(sessionData.Workspace.ID), strings.TrimSpace(sessionData.User.ActiveWorkspaceID)),
	}
	if cacheable {
		p.mu.Lock()
		p.cachedAuth = state
		p.mu.Unlock()
	}
	return state, nil
}

func (p *blinkProvider) resolveDirectAuthState(ctx context.Context, options blinkRequestOptions) (*blinkAuthState, error) {
	idToken := strings.TrimSpace(options.IDToken)
	sessionToken := strings.TrimSpace(options.SessionToken)
	if idToken == "" || sessionToken == "" {
		return nil, fmt.Errorf("blink direct session auth requires both id token and session token")
	}

	sessionData, err := p.fetchSessionData(ctx, idToken, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("blink fetch session data: %w", err)
	}

	return &blinkAuthState{
		IDToken:       idToken,
		SessionToken:  sessionToken,
		WorkspaceSlug: firstNonEmptyString(strings.TrimSpace(options.WorkspaceSlug), strings.TrimSpace(sessionData.Workspace.Slug)),
		WorkspaceID:   firstNonEmptyString(strings.TrimSpace(sessionData.Workspace.ID), strings.TrimSpace(sessionData.User.ActiveWorkspaceID)),
	}, nil
}

func (p *blinkProvider) refreshIDToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.refreshURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("build firebase refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", blinkUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("firebase refresh request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", time.Time{}, fmt.Errorf("firebase refresh error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload blinkFirebaseRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", time.Time{}, fmt.Errorf("decode firebase refresh response: %w", err)
	}
	if strings.TrimSpace(payload.IDToken) == "" {
		return "", "", time.Time{}, fmt.Errorf("firebase refresh response missing id_token")
	}

	expiresInSeconds := int64(3600)
	if seconds := strings.TrimSpace(payload.ExpiresIn); seconds != "" {
		if parsed, err := strconv.ParseInt(seconds, 10, 64); err == nil && parsed > 0 {
			expiresInSeconds = parsed
		}
	}
	expiresAt := time.Now().Add(time.Duration(expiresInSeconds-60) * time.Second)
	return strings.TrimSpace(payload.IDToken), strings.TrimSpace(payload.RefreshToken), expiresAt, nil
}

func (p *blinkProvider) createSession(ctx context.Context, idToken string) (string, error) {
	body, _ := json.Marshal(map[string]string{"idToken": idToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.sessionURL(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build blink session request: %w", err)
	}
	for key, value := range p.baseHeaders(p.baseURL+blinkDefaultReferer, "") {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("blink session request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("blink session error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" && strings.TrimSpace(cookie.Value) != "" {
			return strings.TrimSpace(cookie.Value), nil
		}
	}
	return "", fmt.Errorf("blink session response missing session cookie")
}

func (p *blinkProvider) fetchSessionData(ctx context.Context, idToken string, sessionToken string) (*blinkSessionDataResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.sessionDataURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("build blink session-data request: %w", err)
	}
	auth := &blinkAuthState{SessionToken: sessionToken}
	for key, value := range p.baseHeaders(p.baseURL+blinkDefaultReferer, sessionToken) {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+idToken)
	req.Header.Set("Accept", "application/json")
	if cookie := p.cookieHeader(auth); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blink session-data request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("blink session-data error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload blinkSessionDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode blink session-data: %w", err)
	}
	return &payload, nil
}

func (p *blinkProvider) createProject(ctx context.Context, auth *blinkAuthState, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		prompt = "API chat session"
	}
	payload := blinkProjectCreateRequest{
		Prompt:     prompt,
		Visibility: blinkProjectVisibility,
		AppType:    blinkProjectAppType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal blink project payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.projectCreateURL(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build blink project request: %w", err)
	}
	referer := p.baseURL
	if auth.WorkspaceSlug != "" {
		referer = p.baseURL + "/" + auth.WorkspaceSlug
	}
	for key, value := range p.baseHeaders(referer, auth.SessionToken) {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+auth.IDToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if cookie := p.cookieHeader(auth); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("blink project request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("blink project error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var response blinkProjectCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode blink project response: %w", err)
	}
	if strings.TrimSpace(response.ID) == "" {
		return "", fmt.Errorf("blink project response missing id")
	}
	return strings.TrimSpace(response.ID), nil
}

func (p *blinkProvider) createCheckout(ctx context.Context, auth *blinkAuthState, payload blinkCheckoutRequest) (*blinkCheckoutResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal blink checkout payload: %w", err)
	}
	referer := strings.TrimSpace(payload.CancelURL)
	if referer == "" {
		referer = p.baseURL
		if strings.TrimSpace(auth.WorkspaceSlug) != "" {
			referer = fmt.Sprintf("%s/%s?showPricing=true", p.baseURL, strings.TrimSpace(auth.WorkspaceSlug))
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.checkoutURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build blink checkout request: %w", err)
	}
	for key, value := range p.baseHeaders(referer, auth.SessionToken) {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+auth.IDToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if cookie := p.cookieHeader(auth); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blink checkout request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		p.invalidateCachedAuth()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("blink checkout error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response blinkCheckoutResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode blink checkout response: %w", err)
	}
	if strings.TrimSpace(response.URL) == "" {
		return nil, fmt.Errorf("blink checkout response missing url")
	}
	return &response, nil
}

func (p *blinkProvider) sendChat(ctx context.Context, req core.UnifiedRequest, auth *blinkAuthState, projectID string, prompt string) (*http.Response, error) {
	payload := blinkChatRequest{
		ID:        projectID,
		ProjectID: projectID,
		Mode:      "agent",
		Message: blinkMessage{
			Role: "user",
			Parts: []blinkPart{{
				Type: "text",
				Text: prompt,
			}},
			ID: randomHex(8),
		},
		ModelID:         mapBlinkModel(req.Model),
		ThinkingEnabled: false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal blink chat payload: %w", err)
	}
	reqURL := p.chatURL()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build blink chat request: %w", err)
	}
	referer := fmt.Sprintf("%s/project/%s", p.baseURL, projectID)
	for key, value := range p.baseHeaders(referer, auth.SessionToken) {
		httpReq.Header.Set(key, value)
	}
	httpReq.Header.Set("Authorization", "Bearer "+auth.IDToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if cookie := p.cookieHeader(auth); cookie != "" {
		httpReq.Header.Set("Cookie", cookie)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("blink chat request failed: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		p.invalidateCachedAuth()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("blink upstream error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func (p *blinkProvider) consumeSSE(ctx context.Context, body io.ReadCloser, output chan<- core.TextStreamEvent) {
	defer body.Close()
	defer close(output)

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
		var event blinkStreamEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			continue
		}
		switch event.Type {
		case "text-delta":
			if strings.TrimSpace(event.Delta) == "" {
				continue
			}
			if !emitBlinkEvent(ctx, output, core.TextStreamEvent{Delta: event.Delta}) {
				return
			}
		case "error":
			message := strings.TrimSpace(event.Error)
			if message == "" {
				message = "blink upstream returned error event"
			}
			emitBlinkEvent(ctx, output, core.TextStreamEvent{Err: fmt.Errorf(message)})
			return
		}
	}
	if err := scanner.Err(); err != nil {
		emitBlinkEvent(ctx, output, core.TextStreamEvent{Err: fmt.Errorf("read blink sse response: %w", err)})
	}
}

func (p *blinkProvider) requestOptions(req core.UnifiedRequest) blinkRequestOptions {
	if len(req.ProviderOptions) == 0 {
		return blinkRequestOptions{}
	}
	return blinkRequestOptions{
		RefreshToken:  strings.TrimSpace(req.ProviderOptions["blink_refresh_token"]),
		IDToken:       strings.TrimSpace(req.ProviderOptions["blink_id_token"]),
		SessionToken:  strings.TrimSpace(req.ProviderOptions["blink_session_token"]),
		WorkspaceSlug: strings.TrimSpace(req.ProviderOptions["blink_workspace_slug"]),
		ProjectID:     strings.TrimSpace(req.ProviderOptions["blink_project_id"]),
	}
}

func (p *blinkProvider) effectiveRequestOptions(options blinkRequestOptions) blinkRequestOptions {
	return blinkRequestOptions{
		RefreshToken:  firstNonEmptyString(options.RefreshToken, p.refreshToken),
		IDToken:       firstNonEmptyString(options.IDToken, p.idToken),
		SessionToken:  firstNonEmptyString(options.SessionToken, p.sessionToken),
		WorkspaceSlug: firstNonEmptyString(options.WorkspaceSlug, p.workspaceSlug),
		ProjectID:     firstNonEmptyString(options.ProjectID, p.projectID),
	}
}

func (p *blinkProvider) baseHeaders(referer string, sessionToken string) map[string]string {
	headers := map[string]string{
		"Origin":          p.baseURL,
		"Referer":         referer,
		"User-Agent":      blinkUserAgent,
		"Accept-Language": "zh-CN,zh;q=0.9",
	}
	if strings.TrimSpace(sessionToken) != "" {
		headers["X-Session-Token"] = sessionToken
	}
	return headers
}

func (p *blinkProvider) cookieHeader(auth *blinkAuthState) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(auth.SessionToken) != "" {
		parts = append(parts, "session="+strings.TrimSpace(auth.SessionToken))
	}
	if strings.TrimSpace(auth.WorkspaceSlug) != "" {
		parts = append(parts, "workspace_slug="+strings.TrimSpace(auth.WorkspaceSlug))
	}
	return strings.Join(parts, "; ")
}

func (p *blinkProvider) invalidateCachedAuth() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cachedAuth = nil
}

func (p *blinkProvider) chatURL() string {
	return p.baseURL + "/api/chat"
}

func (p *blinkProvider) projectCreateURL() string {
	return p.baseURL + "/api/projects/create"
}

func (p *blinkProvider) sessionURL() string {
	return p.baseURL + "/api/auth/session"
}

func (p *blinkProvider) checkoutURL() string {
	return p.baseURL + "/api/stripe/checkout"
}

func (p *blinkProvider) sessionDataURL() string {
	return p.baseURL + "/api/auth/session-data"
}

func buildBlinkPrompt(messages []core.Message) (string, string) {
	currentRequest := currentBlinkRequest(messages)
	history := formatBlinkHistory(messages)
	systems := collectBlinkSystem(messages)

	if history == "" && len(systems) == 0 {
		return blinkProjectSeed(currentRequest), currentRequest
	}

	sections := make([]string, 0, 3)
	if len(systems) > 0 {
		sections = append(sections, fmt.Sprintf("<system>\n%s\n</system>", strings.Join(systems, "\n\n")))
	}
	if history != "" {
		sections = append(sections, fmt.Sprintf("<conversation_history>\n%s\n</conversation_history>", history))
	}
	sections = append(sections, fmt.Sprintf("<current_request>\n%s\n</current_request>", currentRequest))
	return blinkProjectSeed(currentRequest), strings.Join(sections, "\n\n")
}

func collectBlinkSystem(messages []core.Message) []string {
	out := make([]string, 0, len(messages))
	for _, msg := range messages {
		if !strings.EqualFold(msg.Role, "system") {
			continue
		}
		text := strings.TrimSpace(core.ContentText(msg.Content))
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	return out
}

func currentBlinkRequest(messages []core.Message) string {
	for idx := len(messages) - 1; idx >= 0; idx-- {
		msg := messages[idx]
		if !strings.EqualFold(msg.Role, "user") {
			continue
		}
		text := strings.TrimSpace(core.ContentText(msg.Content))
		if text != "" {
			return text
		}
	}
	for idx := len(messages) - 1; idx >= 0; idx-- {
		text := strings.TrimSpace(core.ContentText(messages[idx].Content))
		if text != "" {
			return text
		}
	}
	return "继续"
}

func formatBlinkHistory(messages []core.Message) string {
	dialogue := make([]core.Message, 0, len(messages))
	for _, msg := range messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		text := strings.TrimSpace(core.ContentText(msg.Content))
		if text == "" {
			continue
		}
		dialogue = append(dialogue, core.Message{Role: role, Content: text})
	}
	if len(dialogue) > 0 && strings.EqualFold(dialogue[len(dialogue)-1].Role, "user") {
		dialogue = dialogue[:len(dialogue)-1]
	}
	parts := make([]string, 0, len(dialogue))
	for idx, msg := range dialogue {
		parts = append(parts, fmt.Sprintf("<turn index=\"%d\" role=\"%s\">\n%s\n</turn>", idx+1, msg.Role, strings.TrimSpace(core.ContentText(msg.Content))))
	}
	return strings.Join(parts, "\n\n")
}

func blinkProjectSeed(current string) string {
	seed := strings.TrimSpace(current)
	if seed == "" {
		seed = "API chat session"
	}
	if len(seed) > 2000 {
		return seed[:2000]
	}
	return seed
}

func mapBlinkModel(model string) string {
	lower := strings.ToLower(strings.TrimSpace(model))
	switch {
	case lower == "", strings.Contains(lower, "sonnet-4.6"):
		return "anthropic/claude-sonnet-4.6"
	case strings.Contains(lower, "opus-4.6"):
		return "anthropic/claude-opus-4.6"
	case strings.Contains(lower, "opus-4"):
		return "anthropic/claude-opus-4"
	case strings.Contains(lower, "opus"):
		return "anthropic/claude-opus-4.6"
	case strings.Contains(lower, "haiku"):
		return "anthropic/claude-3.5-haiku"
	case strings.Contains(lower, "gpt-4.1"):
		return "openai/gpt-4.1"
	case strings.Contains(lower, "gpt-4o"):
		return "openai/gpt-4o"
	default:
		if strings.Contains(lower, "/") {
			return strings.TrimSpace(model)
		}
		return "anthropic/claude-sonnet-4.6"
	}
}

func normalizeBlinkPlanID(planID string) string {
	value := strings.ToLower(strings.TrimSpace(planID))
	if value == "" {
		return "pro"
	}
	return value
}

func defaultBlinkPriceID(planID string) string {
	switch normalizeBlinkPlanID(planID) {
	case "pro":
		return blinkProPriceID
	default:
		return ""
	}
}

func optionalStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func emitBlinkEvent(ctx context.Context, output chan<- core.TextStreamEvent, event core.TextStreamEvent) bool {
	select {
	case <-ctx.Done():
		select {
		case output <- core.TextStreamEvent{Err: ctx.Err()}:
		default:
		}
		return false
	case output <- event:
		return true
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func randomHex(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", n*2)
	}
	return hex.EncodeToString(buf)
}
