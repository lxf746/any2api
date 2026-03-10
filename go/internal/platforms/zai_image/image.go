// Package zai_image implements the Z.ai Image Generation API client.
// Endpoint: POST https://image.z.ai/api/proxy/images/generate
// Auth: Cookie-based session JWT (not Bearer header).
package zai_image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	DefaultEndpoint         = "https://image.z.ai/api/proxy/images/generate"
	DefaultAuthEndpoint     = "https://image.z.ai/api/v1/z-image/auth/"
	DefaultCallbackEndpoint = "https://image.z.ai/api/auth/callback"
	DefaultTimeout          = 120 * time.Second
)

// ImageClient calls the Z.ai Image Generation API.
type ImageClient struct {
	Endpoint         string
	AuthEndpoint     string
	CallbackEndpoint string
	SessionToken     string // JWT token set as "session" cookie
	Client           *http.Client
}

// NewImageClient creates a client with the given session JWT token.
func NewImageClient(sessionToken string) *ImageClient {
	return &ImageClient{
		Endpoint:         DefaultEndpoint,
		AuthEndpoint:     DefaultAuthEndpoint,
		CallbackEndpoint: DefaultCallbackEndpoint,
		SessionToken:     sessionToken,
		Client:           &http.Client{Timeout: DefaultTimeout},
	}
}

// Authenticate performs the full auth flow: code → token → session cookie.
// Step 1: POST /api/v1/z-image/auth/ with {"code": "..."} → get auth_token
// Step 2: POST /api/auth/callback with {"token": "..."} → register session
func (c *ImageClient) Authenticate(code string) (*AuthResponse, error) {
	// Step 1: exchange code for token
	authPayload, _ := json.Marshal(AuthRequest{Code: code})
	authReq, err := http.NewRequest(http.MethodPost, c.AuthEndpoint, bytes.NewReader(authPayload))
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	authReq.Header.Set("Content-Type", "application/json")
	authReq.Header.Set("Accept", "*/*")
	authReq.Header.Set("X-Request-ID", randomID(22))
	authReq.Header.Set("Origin", "https://image.z.ai")
	authReq.Header.Set("Referer", "https://image.z.ai/")

	authResp, err := c.Client.Do(authReq)
	if err != nil {
		return nil, fmt.Errorf("auth request: %w", err)
	}
	defer authResp.Body.Close()

	authBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	if authResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth HTTP %d: %s", authResp.StatusCode, string(authBody))
	}

	var result AuthResponse
	if err := json.Unmarshal(authBody, &result); err != nil {
		return nil, fmt.Errorf("parse auth response: %w", err)
	}
	if result.Data.AuthToken == "" {
		return nil, fmt.Errorf("auth returned empty token")
	}

	// Step 2: register token as session cookie
	if err := c.registerCallback(result.Data.AuthToken); err != nil {
		return nil, fmt.Errorf("callback: %w", err)
	}

	c.SessionToken = result.Data.AuthToken
	return &result, nil
}

// registerCallback calls POST /api/auth/callback to register the session.
func (c *ImageClient) registerCallback(token string) error {
	payload, _ := json.Marshal(CallbackRequest{Token: token})
	req, err := http.NewRequest(http.MethodPost, c.CallbackEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://image.z.ai")
	req.Header.Set("Referer", "https://image.z.ai/")

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("callback request: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback HTTP %d", resp.StatusCode)
	}
	return nil
}

// NewImageClientFromCode creates a client by authenticating with an OAuth code.
func NewImageClientFromCode(code string) (*ImageClient, *AuthResponse, error) {
	client := &ImageClient{
		Endpoint:         DefaultEndpoint,
		AuthEndpoint:     DefaultAuthEndpoint,
		CallbackEndpoint: DefaultCallbackEndpoint,
		Client:           &http.Client{Timeout: DefaultTimeout},
	}
	auth, err := client.Authenticate(code)
	if err != nil {
		return nil, nil, err
	}
	return client, auth, nil
}

// Generate creates an image from a text prompt.
func (c *ImageClient) Generate(req ImageRequest) (*ImageResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("X-Request-ID", randomID(22))
	httpReq.Header.Set("Origin", "https://image.z.ai")
	httpReq.Header.Set("Referer", "https://image.z.ai/create")
	httpReq.AddCookie(&http.Cookie{Name: "session", Value: c.SessionToken})

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ImageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GenerateSimple is a convenience method with common defaults.
func (c *ImageClient) GenerateSimple(prompt string) (*ImageResponse, error) {
	return c.Generate(ImageRequest{
		Prompt:           prompt,
		Ratio:            "1:1",
		Resolution:       "1K",
		RmLabelWatermark: true,
	})
}

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomID(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = idChars[r.Intn(len(idChars))]
	}
	return string(b)
}
