package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AdminSettings struct {
	AdminPassword   string `json:"adminPassword,omitempty"`
	APIKey          string `json:"apiKey"`
	DefaultProvider string `json:"defaultProvider"`
}

type KiroAccount struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	AccessToken       string    `json:"accessToken"`
	MachineID         string    `json:"machineId"`
	PreferredEndpoint string    `json:"preferredEndpoint,omitempty"`
	Active            bool      `json:"active"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type GrokToken struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CookieToken string    `json:"cookieToken"`
	Active      bool      `json:"active"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GrokRuntimeConfig struct {
	APIURL      string `json:"apiUrl"`
	ProxyURL    string `json:"proxyUrl"`
	CFCookies   string `json:"cfCookies"`
	CFClearance string `json:"cfClearance"`
	UserAgent   string `json:"userAgent"`
	Origin      string `json:"origin"`
	Referer     string `json:"referer"`
}

type CursorRuntimeConfig struct {
	APIURL        string `json:"apiUrl"`
	ScriptURL     string `json:"scriptUrl"`
	Cookie        string `json:"cookie"`
	XIsHuman      string `json:"xIsHuman"`
	UserAgent     string `json:"userAgent"`
	Referer       string `json:"referer"`
	WebGLVendor   string `json:"webglVendor"`
	WebGLRenderer string `json:"webglRenderer"`
}

type OrchidsRuntimeConfig struct {
	APIURL       string `json:"apiUrl"`
	ClerkURL     string `json:"clerkUrl"`
	ClientCookie string `json:"clientCookie"`
	ClientUAT    string `json:"clientUat"`
	SessionID    string `json:"sessionId"`
	ProjectID    string `json:"projectId"`
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	AgentMode    string `json:"agentMode"`
}

type WebRuntimeConfig struct {
	BaseURL string `json:"baseUrl"`
	Type    string `json:"type"`
	APIKey  string `json:"apiKey"`
}

type ChatGPTRuntimeConfig struct {
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
}

type BlinkRuntimeConfig struct {
	BaseURL            string `json:"baseUrl"`
	FirebaseRefreshURL string `json:"firebaseRefreshUrl"`
	RefreshToken       string `json:"refreshToken"`
	IDToken            string `json:"idToken"`
	SessionToken       string `json:"sessionToken"`
	WorkspaceSlug      string `json:"workspaceSlug"`
	ProjectID          string `json:"projectId"`
	ProxyURL           string `json:"proxyUrl"`
}

type ZAIImageRuntimeConfig struct {
	SessionToken string `json:"sessionToken"`
	APIURL       string `json:"apiUrl"`
}

type ZAITTSRuntimeConfig struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
	APIURL string `json:"apiUrl"`
}

type ZAIOCRRuntimeConfig struct {
	Token  string `json:"token"`
	APIURL string `json:"apiUrl"`
}

type ProviderState struct {
	Enabled        bool            `json:"enabled"`
	DisabledModels map[string]bool `json:"disabledModels,omitempty"`
}

type ProviderStore struct {
	CursorConfig   CursorRuntimeConfig      `json:"cursorConfig,omitempty"`
	KiroAccounts   []KiroAccount            `json:"kiroAccounts"`
	GrokConfig     GrokRuntimeConfig        `json:"grokConfig,omitempty"`
	GrokTokens     []GrokToken              `json:"grokTokens"`
	OrchidsConfig  OrchidsRuntimeConfig     `json:"orchidsConfig,omitempty"`
	WebConfig      WebRuntimeConfig         `json:"webConfig,omitempty"`
	ChatGPTConfig  ChatGPTRuntimeConfig     `json:"chatgptConfig,omitempty"`
	BlinkConfig    BlinkRuntimeConfig       `json:"blinkConfig,omitempty"`
	ZAIImageConfig ZAIImageRuntimeConfig    `json:"zaiImageConfig,omitempty"`
	ZAITTSConfig   ZAITTSRuntimeConfig      `json:"zaiTTSConfig,omitempty"`
	ZAIOCRConfig   ZAIOCRRuntimeConfig      `json:"zaiOCRConfig,omitempty"`
	ProviderStates map[string]ProviderState `json:"providerStates,omitempty"`
}

type UserRecord struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AdminConfig struct {
	Settings  AdminSettings `json:"settings"`
	Providers ProviderStore `json:"providers"`
	Users     []UserRecord  `json:"users,omitempty"`
}

type RuntimeManager struct {
	mu   sync.RWMutex
	path string
	base AppConfig
	data AdminConfig
}

func NewRuntimeManager(path string, base AppConfig) (*RuntimeManager, error) {
	mgr := &RuntimeManager{path: path, base: base, data: defaultAdminConfig(base)}
	if err := mgr.load(); err != nil {
		return nil, err
	}
	return mgr, nil
}

func (m *RuntimeManager) Snapshot() AdminConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneAdminConfig(m.data)
}

func (m *RuntimeManager) CurrentAppConfig() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return applyAdminConfig(m.base, m.data)
}

func (m *RuntimeManager) AdminPassword() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return strings.TrimSpace(m.data.Settings.AdminPassword)
}

func (m *RuntimeManager) UpdateSettings(apiKey string, defaultProvider string, adminPassword string) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Settings.APIKey = strings.TrimSpace(apiKey)
	defaultProvider = normalizeProviderID(defaultProvider)
	if defaultProvider == "" {
		defaultProvider = normalizeProviderID(m.base.DefaultProvider)
	}
	m.data.Settings.DefaultProvider = defaultProvider
	if trimmed := strings.TrimSpace(adminPassword); trimmed != "" {
		m.data.Settings.AdminPassword = trimmed
	}
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceKiroAccounts(accounts []KiroAccount) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.KiroAccounts = normalizeKiroAccounts(accounts)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) KiroAccount(id string) (KiroAccount, bool) {
	data := m.Snapshot()
	account, _, ok := findKiroAccount(data.Providers.KiroAccounts, id)
	return account, ok
}

func (m *RuntimeManager) CreateKiroAccount(account KiroAccount) (KiroAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, _, exists := findKiroAccount(m.data.Providers.KiroAccounts, account.ID); exists {
		account.ID = ""
	}
	prepared := normalizeKiroAccounts([]KiroAccount{account})
	if len(prepared) == 0 {
		return KiroAccount{}, errors.New("invalid kiro account")
	}
	next := append(append([]KiroAccount(nil), m.data.Providers.KiroAccounts...), prepared[0])
	if prepared[0].Active {
		for idx := range next {
			if strings.TrimSpace(next[idx].ID) != strings.TrimSpace(prepared[0].ID) {
				next[idx].Active = false
			}
		}
	}
	next = normalizeKiroAccounts(next)
	created, _, ok := findKiroAccount(next, prepared[0].ID)
	if !ok {
		return KiroAccount{}, errors.New("invalid kiro account")
	}
	m.data.Providers.KiroAccounts = next
	if err := m.persistLocked(); err != nil {
		return KiroAccount{}, err
	}
	return created, nil
}

func (m *RuntimeManager) UpdateKiroAccount(id string, account KiroAccount) (KiroAccount, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, idx, ok := findKiroAccount(m.data.Providers.KiroAccounts, id)
	if !ok {
		return KiroAccount{}, false, nil
	}
	account.ID = strings.TrimSpace(id)
	account.UpdatedAt = time.Now().UTC()
	next := append([]KiroAccount(nil), m.data.Providers.KiroAccounts...)
	if account.Active {
		for i := range next {
			next[i].Active = false
		}
	}
	next[idx] = account
	next = normalizeKiroAccounts(next)
	updated, _, ok := findKiroAccount(next, id)
	if !ok {
		return KiroAccount{}, false, errors.New("invalid kiro account")
	}
	m.data.Providers.KiroAccounts = next
	if err := m.persistLocked(); err != nil {
		return KiroAccount{}, false, err
	}
	return updated, true, nil
}

func (m *RuntimeManager) DeleteKiroAccount(id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, idx, ok := findKiroAccount(m.data.Providers.KiroAccounts, id)
	if !ok {
		return false, nil
	}
	next := append([]KiroAccount(nil), m.data.Providers.KiroAccounts...)
	next = append(next[:idx], next[idx+1:]...)
	m.data.Providers.KiroAccounts = normalizeKiroAccounts(next)
	if err := m.persistLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func (m *RuntimeManager) ReplaceCursorConfig(cfg CursorRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.CursorConfig = normalizeCursorConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceGrokTokens(tokens []GrokToken) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.GrokTokens = normalizeGrokTokens(tokens)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) GrokToken(id string) (GrokToken, bool) {
	data := m.Snapshot()
	token, _, ok := findGrokToken(data.Providers.GrokTokens, id)
	return token, ok
}

func (m *RuntimeManager) CreateGrokToken(token GrokToken) (GrokToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, _, exists := findGrokToken(m.data.Providers.GrokTokens, token.ID); exists {
		token.ID = ""
	}
	prepared := normalizeGrokTokens([]GrokToken{token})
	if len(prepared) == 0 {
		return GrokToken{}, errors.New("invalid grok token")
	}
	next := append(append([]GrokToken(nil), m.data.Providers.GrokTokens...), prepared[0])
	if prepared[0].Active {
		for idx := range next {
			if strings.TrimSpace(next[idx].ID) != strings.TrimSpace(prepared[0].ID) {
				next[idx].Active = false
			}
		}
	}
	next = normalizeGrokTokens(next)
	created, _, ok := findGrokToken(next, prepared[0].ID)
	if !ok {
		return GrokToken{}, errors.New("invalid grok token")
	}
	m.data.Providers.GrokTokens = next
	if err := m.persistLocked(); err != nil {
		return GrokToken{}, err
	}
	return created, nil
}

func (m *RuntimeManager) UpdateGrokToken(id string, token GrokToken) (GrokToken, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, idx, ok := findGrokToken(m.data.Providers.GrokTokens, id)
	if !ok {
		return GrokToken{}, false, nil
	}
	token.ID = strings.TrimSpace(id)
	token.UpdatedAt = time.Now().UTC()
	next := append([]GrokToken(nil), m.data.Providers.GrokTokens...)
	if token.Active {
		for i := range next {
			next[i].Active = false
		}
	}
	next[idx] = token
	next = normalizeGrokTokens(next)
	updated, _, ok := findGrokToken(next, id)
	if !ok {
		return GrokToken{}, false, errors.New("invalid grok token")
	}
	m.data.Providers.GrokTokens = next
	if err := m.persistLocked(); err != nil {
		return GrokToken{}, false, err
	}
	return updated, true, nil
}

func (m *RuntimeManager) DeleteGrokToken(id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, idx, ok := findGrokToken(m.data.Providers.GrokTokens, id)
	if !ok {
		return false, nil
	}
	next := append([]GrokToken(nil), m.data.Providers.GrokTokens...)
	next = append(next[:idx], next[idx+1:]...)
	m.data.Providers.GrokTokens = normalizeGrokTokens(next)
	if err := m.persistLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func (m *RuntimeManager) ReplaceGrokConfig(cfg GrokRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.GrokConfig = normalizeGrokConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceOrchidsConfig(cfg OrchidsRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.OrchidsConfig = normalizeOrchidsConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceWebConfig(cfg WebRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.WebConfig = normalizeWebConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceChatGPTConfig(cfg ChatGPTRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.ChatGPTConfig = normalizeChatGPTConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceBlinkConfig(cfg BlinkRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.BlinkConfig = normalizeBlinkConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceZAIImageConfig(cfg ZAIImageRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.ZAIImageConfig = normalizeZAIImageConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceZAITTSConfig(cfg ZAITTSRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.ZAITTSConfig = normalizeZAITTSConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) ReplaceZAIOCRConfig(cfg ZAIOCRRuntimeConfig) (AdminConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data.Providers.ZAIOCRConfig = normalizeZAIOCRConfig(cfg)
	if err := m.persistLocked(); err != nil {
		return AdminConfig{}, err
	}
	return cloneAdminConfig(m.data), nil
}

func (m *RuntimeManager) load() error {
	content, err := os.ReadFile(m.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read admin config: %w", err)
		}
		return m.persistInitial()
	}
	if strings.TrimSpace(string(content)) == "" {
		return m.persistInitial()
	}
	loaded := defaultAdminConfig(m.base)
	if err := json.Unmarshal(content, &loaded); err != nil {
		return fmt.Errorf("decode admin config: %w", err)
	}
	loaded.Settings.DefaultProvider = normalizeProviderID(loaded.Settings.DefaultProvider)
	if loaded.Settings.DefaultProvider == "" {
		loaded.Settings.DefaultProvider = normalizeProviderID(m.base.DefaultProvider)
	}
	if strings.TrimSpace(loaded.Settings.AdminPassword) == "" {
		loaded.Settings.AdminPassword = m.base.AdminPassword
	}
	loaded.Providers.CursorConfig = normalizeCursorConfig(loaded.Providers.CursorConfig)
	loaded.Providers.KiroAccounts = normalizeKiroAccounts(loaded.Providers.KiroAccounts)
	loaded.Providers.GrokConfig = normalizeGrokConfig(loaded.Providers.GrokConfig)
	loaded.Providers.GrokTokens = normalizeGrokTokens(loaded.Providers.GrokTokens)
	loaded.Providers.OrchidsConfig = normalizeOrchidsConfig(loaded.Providers.OrchidsConfig)
	loaded.Providers.WebConfig = normalizeWebConfig(loaded.Providers.WebConfig)
	loaded.Providers.ChatGPTConfig = normalizeChatGPTConfig(loaded.Providers.ChatGPTConfig)
	loaded.Providers.BlinkConfig = normalizeBlinkConfig(loaded.Providers.BlinkConfig)
	loaded.Providers.ZAIImageConfig = normalizeZAIImageConfig(loaded.Providers.ZAIImageConfig)
	loaded.Providers.ZAITTSConfig = normalizeZAITTSConfig(loaded.Providers.ZAITTSConfig)
	loaded.Providers.ZAIOCRConfig = normalizeZAIOCRConfig(loaded.Providers.ZAIOCRConfig)
	m.data = loaded
	return nil
}

func (m *RuntimeManager) persistInitial() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = defaultAdminConfig(m.base)
	return m.persistLocked()
}

func (m *RuntimeManager) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return fmt.Errorf("create admin config dir: %w", err)
	}
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal admin config: %w", err)
	}
	if err := os.WriteFile(m.path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write admin config: %w", err)
	}
	return nil
}

func defaultAdminConfig(base AppConfig) AdminConfig {
	return AdminConfig{
		Settings: AdminSettings{
			AdminPassword:   base.AdminPassword,
			APIKey:          base.APIKey,
			DefaultProvider: normalizeProviderID(base.DefaultProvider),
		},
		Providers: ProviderStore{
			CursorConfig:   defaultCursorConfig(base),
			KiroAccounts:   normalizeKiroAccounts(defaultKiroAccounts(base)),
			GrokConfig:     defaultGrokConfig(base),
			GrokTokens:     normalizeGrokTokens(defaultGrokTokens(base)),
			OrchidsConfig:  normalizeOrchidsConfig(defaultOrchidsConfig(base)),
			WebConfig:      normalizeWebConfig(defaultWebConfig(base)),
			ChatGPTConfig:  normalizeChatGPTConfig(defaultChatGPTConfig(base)),
			BlinkConfig:    normalizeBlinkConfig(defaultBlinkConfig(base)),
			ZAIImageConfig: normalizeZAIImageConfig(defaultZAIImageConfig(base)),
			ZAITTSConfig:   normalizeZAITTSConfig(defaultZAITTSConfig(base)),
			ZAIOCRConfig:   normalizeZAIOCRConfig(defaultZAIOCRConfig(base)),
		},
	}
}

func applyAdminConfig(base AppConfig, admin AdminConfig) AppConfig {
	cfg := base
	cfg.APIKey = admin.Settings.APIKey
	cfg.DefaultProvider = normalizeProviderID(admin.Settings.DefaultProvider)
	cfg.Cursor.APIURL = admin.Providers.CursorConfig.APIURL
	cfg.Cursor.ScriptURL = admin.Providers.CursorConfig.ScriptURL
	cfg.Cursor.Cookie = admin.Providers.CursorConfig.Cookie
	cfg.Cursor.XIsHuman = admin.Providers.CursorConfig.XIsHuman
	cfg.Cursor.UserAgent = admin.Providers.CursorConfig.UserAgent
	cfg.Cursor.Referer = admin.Providers.CursorConfig.Referer
	cfg.Cursor.Fingerprint.WebGLVendor = admin.Providers.CursorConfig.WebGLVendor
	cfg.Cursor.Fingerprint.WebGLRenderer = admin.Providers.CursorConfig.WebGLRenderer
	for _, account := range admin.Providers.KiroAccounts {
		if !account.Active {
			continue
		}
		cfg.Kiro.AccessToken = account.AccessToken
		cfg.Kiro.MachineID = account.MachineID
		cfg.Kiro.PreferredEndpoint = account.PreferredEndpoint
		break
	}
	cfg.Grok.APIURL = admin.Providers.GrokConfig.APIURL
	cfg.Grok.ProxyURL = admin.Providers.GrokConfig.ProxyURL
	cfg.Grok.CFCookies = admin.Providers.GrokConfig.CFCookies
	cfg.Grok.CFClearance = admin.Providers.GrokConfig.CFClearance
	cfg.Grok.UserAgent = admin.Providers.GrokConfig.UserAgent
	cfg.Grok.Origin = admin.Providers.GrokConfig.Origin
	cfg.Grok.Referer = admin.Providers.GrokConfig.Referer
	for _, token := range admin.Providers.GrokTokens {
		if !token.Active {
			continue
		}
		cfg.Grok.CookieToken = token.CookieToken
		break
	}
	cfg.Orchids.APIURL = admin.Providers.OrchidsConfig.APIURL
	cfg.Orchids.ClerkURL = admin.Providers.OrchidsConfig.ClerkURL
	cfg.Orchids.ClientCookie = admin.Providers.OrchidsConfig.ClientCookie
	cfg.Orchids.ClientUAT = admin.Providers.OrchidsConfig.ClientUAT
	cfg.Orchids.SessionID = admin.Providers.OrchidsConfig.SessionID
	cfg.Orchids.ProjectID = admin.Providers.OrchidsConfig.ProjectID
	cfg.Orchids.UserID = admin.Providers.OrchidsConfig.UserID
	cfg.Orchids.Email = admin.Providers.OrchidsConfig.Email
	cfg.Orchids.AgentMode = admin.Providers.OrchidsConfig.AgentMode
	cfg.Web.BaseURL = admin.Providers.WebConfig.BaseURL
	cfg.Web.Type = admin.Providers.WebConfig.Type
	cfg.Web.APIKey = admin.Providers.WebConfig.APIKey
	cfg.ChatGPT.BaseURL = admin.Providers.ChatGPTConfig.BaseURL
	cfg.ChatGPT.Token = admin.Providers.ChatGPTConfig.Token
	cfg.Blink.BaseURL = admin.Providers.BlinkConfig.BaseURL
	cfg.Blink.FirebaseRefreshURL = admin.Providers.BlinkConfig.FirebaseRefreshURL
	cfg.Blink.RefreshToken = admin.Providers.BlinkConfig.RefreshToken
	cfg.Blink.IDToken = admin.Providers.BlinkConfig.IDToken
	cfg.Blink.SessionToken = admin.Providers.BlinkConfig.SessionToken
	cfg.Blink.WorkspaceSlug = admin.Providers.BlinkConfig.WorkspaceSlug
	cfg.Blink.ProjectID = admin.Providers.BlinkConfig.ProjectID
	cfg.Blink.ProxyURL = admin.Providers.BlinkConfig.ProxyURL
	cfg.ZAIImage.SessionToken = admin.Providers.ZAIImageConfig.SessionToken
	cfg.ZAIImage.APIURL = admin.Providers.ZAIImageConfig.APIURL
	cfg.ZAITTS.Token = admin.Providers.ZAITTSConfig.Token
	cfg.ZAITTS.UserID = admin.Providers.ZAITTSConfig.UserID
	cfg.ZAITTS.APIURL = admin.Providers.ZAITTSConfig.APIURL
	cfg.ZAIOCR.Token = admin.Providers.ZAIOCRConfig.Token
	cfg.ZAIOCR.APIURL = admin.Providers.ZAIOCRConfig.APIURL
	return cfg
}

func cloneAdminConfig(input AdminConfig) AdminConfig {
	clone := input
	clone.Providers.KiroAccounts = append([]KiroAccount(nil), input.Providers.KiroAccounts...)
	clone.Providers.GrokTokens = append([]GrokToken(nil), input.Providers.GrokTokens...)
	if input.Providers.ProviderStates != nil {
		clone.Providers.ProviderStates = make(map[string]ProviderState, len(input.Providers.ProviderStates))
		for k, v := range input.Providers.ProviderStates {
			state := v
			if v.DisabledModels != nil {
				state.DisabledModels = make(map[string]bool, len(v.DisabledModels))
				for mk, mv := range v.DisabledModels {
					state.DisabledModels[mk] = mv
				}
			}
			clone.Providers.ProviderStates[k] = state
		}
	}
	clone.Users = append([]UserRecord(nil), input.Users...)
	return clone
}

func defaultCursorConfig(base AppConfig) CursorRuntimeConfig {
	return normalizeCursorConfig(CursorRuntimeConfig{
		APIURL:        base.Cursor.APIURL,
		ScriptURL:     base.Cursor.ScriptURL,
		Cookie:        base.Cursor.Cookie,
		XIsHuman:      base.Cursor.XIsHuman,
		UserAgent:     base.Cursor.UserAgent,
		Referer:       base.Cursor.Referer,
		WebGLVendor:   base.Cursor.Fingerprint.WebGLVendor,
		WebGLRenderer: base.Cursor.Fingerprint.WebGLRenderer,
	})
}

func defaultKiroAccounts(base AppConfig) []KiroAccount {
	if strings.TrimSpace(base.Kiro.AccessToken) == "" && strings.TrimSpace(base.Kiro.MachineID) == "" {
		return nil
	}
	return []KiroAccount{{
		ID:                "kiro-env-default",
		Name:              "Env Default Kiro",
		AccessToken:       strings.TrimSpace(base.Kiro.AccessToken),
		MachineID:         strings.TrimSpace(base.Kiro.MachineID),
		PreferredEndpoint: strings.TrimSpace(strings.ToLower(base.Kiro.PreferredEndpoint)),
		Active:            true,
		UpdatedAt:         time.Now().UTC(),
	}}
}

func defaultGrokTokens(base AppConfig) []GrokToken {
	if strings.TrimSpace(base.Grok.CookieToken) == "" {
		return nil
	}
	return []GrokToken{{
		ID:          "grok-env-default",
		Name:        "Env Default Grok",
		CookieToken: strings.TrimSpace(base.Grok.CookieToken),
		Active:      true,
		UpdatedAt:   time.Now().UTC(),
	}}
}

func defaultGrokConfig(base AppConfig) GrokRuntimeConfig {
	return normalizeGrokConfig(GrokRuntimeConfig{
		APIURL:      base.Grok.APIURL,
		ProxyURL:    base.Grok.ProxyURL,
		CFCookies:   base.Grok.CFCookies,
		CFClearance: base.Grok.CFClearance,
		UserAgent:   base.Grok.UserAgent,
		Origin:      base.Grok.Origin,
		Referer:     base.Grok.Referer,
	})
}

func defaultOrchidsConfig(base AppConfig) OrchidsRuntimeConfig {
	return normalizeOrchidsConfig(OrchidsRuntimeConfig{
		APIURL:       base.Orchids.APIURL,
		ClerkURL:     base.Orchids.ClerkURL,
		ClientCookie: base.Orchids.ClientCookie,
		ClientUAT:    base.Orchids.ClientUAT,
		SessionID:    base.Orchids.SessionID,
		ProjectID:    base.Orchids.ProjectID,
		UserID:       base.Orchids.UserID,
		Email:        base.Orchids.Email,
		AgentMode:    base.Orchids.AgentMode,
	})
}

func defaultWebConfig(base AppConfig) WebRuntimeConfig {
	return normalizeWebConfig(WebRuntimeConfig{
		BaseURL: base.Web.BaseURL,
		Type:    base.Web.Type,
		APIKey:  base.Web.APIKey,
	})
}

func defaultChatGPTConfig(base AppConfig) ChatGPTRuntimeConfig {
	return normalizeChatGPTConfig(ChatGPTRuntimeConfig{
		BaseURL: base.ChatGPT.BaseURL,
		Token:   base.ChatGPT.Token,
	})
}

func defaultBlinkConfig(base AppConfig) BlinkRuntimeConfig {
	return normalizeBlinkConfig(BlinkRuntimeConfig{
		BaseURL:            base.Blink.BaseURL,
		FirebaseRefreshURL: base.Blink.FirebaseRefreshURL,
		RefreshToken:       base.Blink.RefreshToken,
		IDToken:            base.Blink.IDToken,
		SessionToken:       base.Blink.SessionToken,
		WorkspaceSlug:      base.Blink.WorkspaceSlug,
		ProjectID:          base.Blink.ProjectID,
		ProxyURL:           base.Blink.ProxyURL,
	})
}

func defaultZAIImageConfig(base AppConfig) ZAIImageRuntimeConfig {
	return normalizeZAIImageConfig(ZAIImageRuntimeConfig{
		SessionToken: base.ZAIImage.SessionToken,
		APIURL:       base.ZAIImage.APIURL,
	})
}

func defaultZAITTSConfig(base AppConfig) ZAITTSRuntimeConfig {
	return normalizeZAITTSConfig(ZAITTSRuntimeConfig{
		Token:  base.ZAITTS.Token,
		UserID: base.ZAITTS.UserID,
		APIURL: base.ZAITTS.APIURL,
	})
}

func defaultZAIOCRConfig(base AppConfig) ZAIOCRRuntimeConfig {
	return normalizeZAIOCRConfig(ZAIOCRRuntimeConfig{
		Token:  base.ZAIOCR.Token,
		APIURL: base.ZAIOCR.APIURL,
	})
}

func normalizeCursorConfig(cfg CursorRuntimeConfig) CursorRuntimeConfig {
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	cfg.ScriptURL = strings.TrimSpace(cfg.ScriptURL)
	cfg.Cookie = strings.TrimSpace(cfg.Cookie)
	cfg.XIsHuman = strings.TrimSpace(cfg.XIsHuman)
	cfg.UserAgent = strings.TrimSpace(cfg.UserAgent)
	cfg.Referer = strings.TrimSpace(cfg.Referer)
	cfg.WebGLVendor = strings.TrimSpace(cfg.WebGLVendor)
	cfg.WebGLRenderer = strings.TrimSpace(cfg.WebGLRenderer)
	return cfg
}

func normalizeKiroAccounts(accounts []KiroAccount) []KiroAccount {
	normalized := make([]KiroAccount, 0, len(accounts))
	activeSet := false
	now := time.Now().UTC()
	for idx, account := range accounts {
		account.AccessToken = strings.TrimSpace(account.AccessToken)
		account.MachineID = strings.TrimSpace(account.MachineID)
		account.PreferredEndpoint = strings.TrimSpace(strings.ToLower(account.PreferredEndpoint))
		if account.AccessToken == "" && account.MachineID == "" {
			continue
		}
		account.ID = normalizedID(account.ID, "kiro", idx)
		account.Name = strings.TrimSpace(account.Name)
		if account.Name == "" {
			account.Name = fmt.Sprintf("Kiro Account %d", len(normalized)+1)
		}
		if activeSet {
			account.Active = false
		} else if account.Active {
			activeSet = true
		}
		if account.UpdatedAt.IsZero() {
			account.UpdatedAt = now
		}
		normalized = append(normalized, account)
	}
	if !activeSet && len(normalized) > 0 {
		normalized[0].Active = true
	}
	return normalized
}

func normalizeGrokTokens(tokens []GrokToken) []GrokToken {
	normalized := make([]GrokToken, 0, len(tokens))
	activeSet := false
	now := time.Now().UTC()
	for idx, token := range tokens {
		token.CookieToken = strings.TrimSpace(token.CookieToken)
		if token.CookieToken == "" {
			continue
		}
		token.ID = normalizedID(token.ID, "grok", idx)
		token.Name = strings.TrimSpace(token.Name)
		if token.Name == "" {
			token.Name = fmt.Sprintf("Grok Token %d", len(normalized)+1)
		}
		if activeSet {
			token.Active = false
		} else if token.Active {
			activeSet = true
		}
		if token.UpdatedAt.IsZero() {
			token.UpdatedAt = now
		}
		normalized = append(normalized, token)
	}
	if !activeSet && len(normalized) > 0 {
		normalized[0].Active = true
	}
	return normalized
}

func normalizeGrokConfig(cfg GrokRuntimeConfig) GrokRuntimeConfig {
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	cfg.ProxyURL = strings.TrimSpace(cfg.ProxyURL)
	cfg.CFCookies = strings.TrimSpace(cfg.CFCookies)
	cfg.CFClearance = strings.TrimSpace(cfg.CFClearance)
	cfg.UserAgent = strings.TrimSpace(cfg.UserAgent)
	cfg.Origin = strings.TrimSpace(cfg.Origin)
	cfg.Referer = strings.TrimSpace(cfg.Referer)
	return cfg
}

func normalizeOrchidsConfig(cfg OrchidsRuntimeConfig) OrchidsRuntimeConfig {
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	cfg.ClerkURL = strings.TrimSpace(cfg.ClerkURL)
	cfg.ClientCookie = strings.TrimSpace(cfg.ClientCookie)
	cfg.ClientUAT = strings.TrimSpace(cfg.ClientUAT)
	cfg.SessionID = strings.TrimSpace(cfg.SessionID)
	cfg.ProjectID = strings.TrimSpace(cfg.ProjectID)
	cfg.UserID = strings.TrimSpace(cfg.UserID)
	cfg.Email = strings.TrimSpace(cfg.Email)
	cfg.AgentMode = strings.TrimSpace(cfg.AgentMode)
	return cfg
}

func normalizeWebConfig(cfg WebRuntimeConfig) WebRuntimeConfig {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.Type = strings.Trim(strings.TrimSpace(cfg.Type), "/")
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	return cfg
}

func normalizeChatGPTConfig(cfg ChatGPTRuntimeConfig) ChatGPTRuntimeConfig {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.Token = strings.TrimSpace(cfg.Token)
	return cfg
}

func normalizeBlinkConfig(cfg BlinkRuntimeConfig) BlinkRuntimeConfig {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.FirebaseRefreshURL = strings.TrimSpace(cfg.FirebaseRefreshURL)
	cfg.RefreshToken = strings.TrimSpace(cfg.RefreshToken)
	cfg.IDToken = strings.TrimSpace(cfg.IDToken)
	cfg.SessionToken = strings.TrimSpace(cfg.SessionToken)
	cfg.WorkspaceSlug = strings.TrimSpace(cfg.WorkspaceSlug)
	cfg.ProjectID = strings.TrimSpace(cfg.ProjectID)
	cfg.ProxyURL = strings.TrimSpace(cfg.ProxyURL)
	return cfg
}

func normalizeZAIImageConfig(cfg ZAIImageRuntimeConfig) ZAIImageRuntimeConfig {
	cfg.SessionToken = strings.TrimSpace(cfg.SessionToken)
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	return cfg
}

func normalizeZAITTSConfig(cfg ZAITTSRuntimeConfig) ZAITTSRuntimeConfig {
	cfg.Token = strings.TrimSpace(cfg.Token)
	cfg.UserID = strings.TrimSpace(cfg.UserID)
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	return cfg
}

func normalizeZAIOCRConfig(cfg ZAIOCRRuntimeConfig) ZAIOCRRuntimeConfig {
	cfg.Token = strings.TrimSpace(cfg.Token)
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	return cfg
}

func normalizedID(value string, prefix string, idx int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		return trimmed
	}
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), idx)
}

func findKiroAccount(accounts []KiroAccount, id string) (KiroAccount, int, bool) {
	target := strings.TrimSpace(id)
	if target == "" {
		return KiroAccount{}, -1, false
	}
	for idx, account := range accounts {
		if strings.TrimSpace(account.ID) == target {
			return account, idx, true
		}
	}
	return KiroAccount{}, -1, false
}

func findGrokToken(tokens []GrokToken, id string) (GrokToken, int, bool) {
	target := strings.TrimSpace(id)
	if target == "" {
		return GrokToken{}, -1, false
	}
	for idx, token := range tokens {
		if strings.TrimSpace(token.ID) == target {
			return token, idx, true
		}
	}
	return GrokToken{}, -1, false
}
