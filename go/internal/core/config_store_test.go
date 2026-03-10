package core

import (
	"path/filepath"
	"testing"
)

func TestRuntimeManagerPersistsAndAppliesProviderSelections(t *testing.T) {
	base := DefaultAppConfig()
	base.APIKey = "base-key"
	base.DefaultProvider = "cursor"
	base.AdminPassword = "secret"
	configPath := filepath.Join(t.TempDir(), "admin.json")

	store, err := NewRuntimeManager(configPath, base)
	if err != nil {
		t.Fatalf("new runtime manager: %v", err)
	}

	if _, err := store.UpdateSettings("runtime-key", "kiro", "updated-secret"); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if _, err := store.ReplaceKiroAccounts([]KiroAccount{{
		Name:              "Primary",
		AccessToken:       "kiro-token",
		MachineID:         "machine-1",
		PreferredEndpoint: "amazonq",
		Active:            true,
	}}); err != nil {
		t.Fatalf("replace kiro accounts: %v", err)
	}
	if _, err := store.ReplaceGrokTokens([]GrokToken{{Name: "Main", CookieToken: "cookie-1", Active: true}}); err != nil {
		t.Fatalf("replace grok tokens: %v", err)
	}
	if _, err := store.ReplaceGrokConfig(GrokRuntimeConfig{
		APIURL:      "https://grok.test/chat",
		ProxyURL:    "http://127.0.0.1:7890",
		CFCookies:   "theme=dark",
		CFClearance: "cf-1",
		UserAgent:   "Mozilla/Test",
		Origin:      "https://grok.test",
		Referer:     "https://grok.test/",
	}); err != nil {
		t.Fatalf("replace grok config: %v", err)
	}
	if _, err := store.ReplaceWebConfig(WebRuntimeConfig{
		BaseURL: "http://127.0.0.1:9000",
		Type:    "claude",
		APIKey:  "web-key",
	}); err != nil {
		t.Fatalf("replace web config: %v", err)
	}
	if _, err := store.ReplaceChatGPTConfig(ChatGPTRuntimeConfig{
		BaseURL: "http://127.0.0.1:5005",
		Token:   "chatgpt-token",
	}); err != nil {
		t.Fatalf("replace chatgpt config: %v", err)
	}
	if _, err := store.ReplaceZAIImageConfig(ZAIImageRuntimeConfig{
		SessionToken: "zai-image-session",
		APIURL:       "https://image.example.com/generate",
	}); err != nil {
		t.Fatalf("replace zai image config: %v", err)
	}
	if _, err := store.ReplaceZAITTSConfig(ZAITTSRuntimeConfig{
		Token:  "zai-tts-token",
		UserID: "zai-user-1",
		APIURL: "https://audio.example.com/tts",
	}); err != nil {
		t.Fatalf("replace zai tts config: %v", err)
	}
	if _, err := store.ReplaceZAIOCRConfig(ZAIOCRRuntimeConfig{
		Token:  "zai-ocr-token",
		APIURL: "https://ocr.example.com/process",
	}); err != nil {
		t.Fatalf("replace zai ocr config: %v", err)
	}

	cfg := store.CurrentAppConfig()
	if cfg.APIKey != "runtime-key" {
		t.Fatalf("expected runtime api key, got %q", cfg.APIKey)
	}
	if cfg.DefaultProvider != "kiro" {
		t.Fatalf("expected default provider kiro, got %q", cfg.DefaultProvider)
	}
	if cfg.Kiro.AccessToken != "kiro-token" || cfg.Kiro.MachineID != "machine-1" {
		t.Fatalf("expected selected kiro account to be applied, got %#v", cfg.Kiro)
	}
	if cfg.Grok.CookieToken != "cookie-1" {
		t.Fatalf("expected selected grok token to be applied, got %q", cfg.Grok.CookieToken)
	}
	if cfg.Grok.APIURL != "https://grok.test/chat" || cfg.Grok.ProxyURL != "http://127.0.0.1:7890" || cfg.Grok.CFClearance != "cf-1" {
		t.Fatalf("expected grok runtime config to be applied, got %#v", cfg.Grok)
	}
	if cfg.Web.BaseURL != "http://127.0.0.1:9000" || cfg.Web.Type != "claude" || cfg.Web.APIKey != "web-key" {
		t.Fatalf("expected web runtime config to be applied, got %#v", cfg.Web)
	}
	if cfg.ChatGPT.BaseURL != "http://127.0.0.1:5005" || cfg.ChatGPT.Token != "chatgpt-token" {
		t.Fatalf("expected chatgpt runtime config to be applied, got %#v", cfg.ChatGPT)
	}
	if cfg.ZAIImage.SessionToken != "zai-image-session" || cfg.ZAIImage.APIURL != "https://image.example.com/generate" {
		t.Fatalf("expected zai image runtime config to be applied, got %#v", cfg.ZAIImage)
	}
	if cfg.ZAITTS.Token != "zai-tts-token" || cfg.ZAITTS.UserID != "zai-user-1" || cfg.ZAITTS.APIURL != "https://audio.example.com/tts" {
		t.Fatalf("expected zai tts runtime config to be applied, got %#v", cfg.ZAITTS)
	}
	if cfg.ZAIOCR.Token != "zai-ocr-token" || cfg.ZAIOCR.APIURL != "https://ocr.example.com/process" {
		t.Fatalf("expected zai ocr runtime config to be applied, got %#v", cfg.ZAIOCR)
	}
	if store.AdminPassword() != "updated-secret" {
		t.Fatalf("expected updated admin password, got %q", store.AdminPassword())
	}

	reloaded, err := NewRuntimeManager(configPath, base)
	if err != nil {
		t.Fatalf("reload runtime manager: %v", err)
	}
	reloadedCfg := reloaded.CurrentAppConfig()
	if reloadedCfg.Kiro.AccessToken != "kiro-token" || reloadedCfg.Grok.CookieToken != "cookie-1" {
		t.Fatalf("expected persisted provider settings after reload, got %#v / %#v", reloadedCfg.Kiro, reloadedCfg.Grok)
	}
	if reloadedCfg.Grok.APIURL != "https://grok.test/chat" || reloadedCfg.Grok.ProxyURL != "http://127.0.0.1:7890" {
		t.Fatalf("expected persisted grok runtime config after reload, got %#v", reloadedCfg.Grok)
	}
	if reloadedCfg.Web.BaseURL != "http://127.0.0.1:9000" || reloadedCfg.Web.Type != "claude" || reloadedCfg.Web.APIKey != "web-key" {
		t.Fatalf("expected persisted web runtime config after reload, got %#v", reloadedCfg.Web)
	}
	if reloadedCfg.ChatGPT.BaseURL != "http://127.0.0.1:5005" || reloadedCfg.ChatGPT.Token != "chatgpt-token" {
		t.Fatalf("expected persisted chatgpt runtime config after reload, got %#v", reloadedCfg.ChatGPT)
	}
	if reloadedCfg.ZAIImage.SessionToken != "zai-image-session" || reloadedCfg.ZAIImage.APIURL != "https://image.example.com/generate" {
		t.Fatalf("expected persisted zai image runtime config after reload, got %#v", reloadedCfg.ZAIImage)
	}
	if reloadedCfg.ZAITTS.Token != "zai-tts-token" || reloadedCfg.ZAITTS.UserID != "zai-user-1" || reloadedCfg.ZAITTS.APIURL != "https://audio.example.com/tts" {
		t.Fatalf("expected persisted zai tts runtime config after reload, got %#v", reloadedCfg.ZAITTS)
	}
	if reloadedCfg.ZAIOCR.Token != "zai-ocr-token" || reloadedCfg.ZAIOCR.APIURL != "https://ocr.example.com/process" {
		t.Fatalf("expected persisted zai ocr runtime config after reload, got %#v", reloadedCfg.ZAIOCR)
	}
}
