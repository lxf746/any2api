// Package zai_tts implements the Z.ai TTS (Text-to-Speech) API client.
// Endpoint: POST https://audio.z.ai/api/v1/z-audio/tts/create
// Auth: Bearer JWT token.
// Response: SSE stream with {"audio":"<base64 WAV>"} chunks, ending with [DONE].
package zai_tts

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultEndpoint     = "https://audio.z.ai/api/v1/z-audio/tts/create"
	DefaultAuthEndpoint = "https://audio.z.ai/api/v1/z-audio/auth/"
	DefaultTimeout      = 120 * time.Second
)

// TTSClient calls the Z.ai TTS API.
type TTSClient struct {
	Endpoint     string
	AuthEndpoint string
	Token        string
	UserID       string
	Client       *http.Client
}

// NewTTSClient creates a client with the given Bearer token and user ID.
func NewTTSClient(token, userID string) *TTSClient {
	return &TTSClient{
		Endpoint:     DefaultEndpoint,
		AuthEndpoint: DefaultAuthEndpoint,
		Token:        token,
		UserID:       userID,
		Client:       &http.Client{Timeout: DefaultTimeout},
	}
}

// Authenticate exchanges an OAuth code for a token. Auto-sets Token and UserID.
func (c *TTSClient) Authenticate(code string) (*AuthResponse, error) {
	payload, _ := json.Marshal(AuthRequest{Code: code})
	req, err := http.NewRequest(http.MethodPost, c.AuthEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://audio.z.ai")
	req.Header.Set("Referer", "https://audio.z.ai/")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result AuthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse auth response: %w", err)
	}
	if result.Data.AuthToken != "" {
		c.Token = result.Data.AuthToken
	}
	if result.Data.UserID != "" {
		c.UserID = result.Data.UserID
	}
	return &result, nil
}

// Synthesize converts text to speech and returns the complete WAV audio bytes.
// It reads the SSE stream, collects all base64 audio chunks, and concatenates them.
func (c *TTSClient) Synthesize(req TTSRequest) ([]byte, error) {
	if req.UserID == "" {
		req.UserID = c.UserID
	}
	if req.Speed == 0 {
		req.Speed = 1
	}
	if req.Volume == 0 {
		req.Volume = 1
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)
	httpReq.Header.Set("Origin", "https://audio.z.ai")
	httpReq.Header.Set("Referer", "https://audio.z.ai/")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return readSSEAudio(resp.Body)
}

// SynthesizeSimple is a convenience method with default voice settings.
func (c *TTSClient) SynthesizeSimple(text string) ([]byte, error) {
	return c.Synthesize(TTSRequest{
		VoiceName: "通用男声",
		VoiceID:   "system_003",
		InputText: text,
		Speed:     1,
		Volume:    1,
	})
}

// readSSEAudio reads SSE events and collects base64 audio chunks into raw bytes.
func readSSEAudio(r io.Reader) ([]byte, error) {
	var audioData []byte
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 4*1024*1024), 16*1024*1024) // up to 16MB per line

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Audio string `json:"audio"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Audio == "" {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(chunk.Audio)
		if err != nil {
			continue
		}
		audioData = append(audioData, decoded...)
	}

	if err := scanner.Err(); err != nil {
		return audioData, fmt.Errorf("read SSE stream: %w", err)
	}
	return audioData, nil
}
