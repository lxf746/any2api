// Package zai_ocr implements the Z.ai OCR API client.
// Endpoint: POST https://ocr.z.ai/api/v1/z-ocr/tasks/process
// Auth: Bearer JWT token, no signature required.
// Request: multipart/form-data with "file" field.
package zai_ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultEndpoint     = "https://ocr.z.ai/api/v1/z-ocr/tasks/process"
	DefaultAuthEndpoint = "https://ocr.z.ai/api/v1/z-ocr/auth/"
	DefaultTimeout      = 120 * time.Second
)

// OCRClient calls the Z.ai OCR API.
type OCRClient struct {
	Endpoint     string
	AuthEndpoint string
	Token        string
	Client       *http.Client
}

// NewOCRClient creates a client with the given Bearer token.
func NewOCRClient(token string) *OCRClient {
	return &OCRClient{
		Endpoint:     DefaultEndpoint,
		AuthEndpoint: DefaultAuthEndpoint,
		Token:        token,
		Client:       &http.Client{Timeout: DefaultTimeout},
	}
}

// Authenticate exchanges an OAuth code for an auth token.
// The code comes from the OAuth redirect URL parameter.
func (c *OCRClient) Authenticate(code string) (*AuthResponse, error) {
	payload, _ := json.Marshal(AuthRequest{Code: code})
	req, err := http.NewRequest(http.MethodPost, c.AuthEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Request-ID", generateUUID())
	req.Header.Set("Origin", "https://ocr.z.ai")
	req.Header.Set("Referer", "https://ocr.z.ai/")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result AuthResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse auth response: %w", err)
	}
	// Auto-set the token for subsequent OCR calls
	if result.Data.AuthToken != "" {
		c.Token = result.Data.AuthToken
	}
	return &result, nil
}

// NewOCRClientFromCode creates a client by first authenticating with an OAuth code.
func NewOCRClientFromCode(code string) (*OCRClient, *AuthResponse, error) {
	client := &OCRClient{
		Endpoint:     DefaultEndpoint,
		AuthEndpoint: DefaultAuthEndpoint,
		Client:       &http.Client{Timeout: DefaultTimeout},
	}
	auth, err := client.Authenticate(code)
	if err != nil {
		return nil, nil, err
	}
	return client, auth, nil
}

// ProcessFile uploads a local file and returns the parsed OCR response.
func (c *OCRClient) ProcessFile(filePath string) (*OCRResponse, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return c.ProcessReader(f, filepath.Base(filePath))
}

// ProcessReader uploads from an io.Reader and returns the parsed OCR response.
func (c *OCRClient) ProcessReader(r io.Reader, filename string) (*OCRResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err = io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("copy file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, c.Endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Request-ID", generateUUID())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://ocr.z.ai")
	req.Header.Set("Referer", "https://ocr.z.ai/")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result OCRResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	// json_content is a stringified JSON — parse it into the structured field
	if result.Data.JsonContentRaw != "" {
		var jc JsonContent
		if err := json.Unmarshal([]byte(result.Data.JsonContentRaw), &jc); err == nil {
			result.Data.JsonContent = &jc
		}
	}
	return &result, nil
}

func generateUUID() string {
	b := make([]byte, 16)
	// simple pseudo-random UUID v4 using current time as seed
	now := time.Now().UnixNano()
	for i := range b {
		b[i] = byte(now >> (i * 4))
		now = now*6364136223846793005 + 1
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
