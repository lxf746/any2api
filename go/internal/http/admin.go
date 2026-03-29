package http

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"any2api-go/internal/core"
)

const adminSessionCookieName = "newplatform2api_admin_session"

const (
	adminBackendLanguage = "go"
	adminBackendVersion  = "dev"
	adminAuthModeSession = "session_cookie"
	adminSessionUserID   = "local-admin"
	adminSessionUserName = "Admin"
	adminSessionUserRole = "admin"
)

//go:embed admin/index.html
var adminIndexHTML string

type adminSessionStore struct {
	mu       sync.Mutex
	sessions map[string]time.Time
}

type adminSettingsResponse struct {
	APIKey                  string `json:"apiKey"`
	DefaultProvider         string `json:"defaultProvider"`
	AdminPasswordConfigured bool   `json:"adminPasswordConfigured"`
}

type adminSettingsRequest struct {
	APIKey          string `json:"apiKey"`
	DefaultProvider string `json:"defaultProvider"`
	AdminPassword   string `json:"adminPassword"`
}

type adminFeatures struct {
	Providers          bool `json:"providers"`
	Credentials        bool `json:"credentials"`
	ProviderState      bool `json:"providerState"`
	Stats              bool `json:"stats"`
	Logs               bool `json:"logs"`
	Users              bool `json:"users"`
	ConfigImportExport bool `json:"configImportExport"`
}

type adminMetaResponse struct {
	Backend struct {
		Language string `json:"language"`
		Version  string `json:"version"`
	} `json:"backend"`
	Auth struct {
		Mode string `json:"mode"`
	} `json:"auth"`
	Features adminFeatures `json:"features"`
}

type adminSessionResponse struct {
	Authenticated bool `json:"authenticated"`
	User          struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"user"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

type adminLoginResponse struct {
	OK    bool   `json:"ok"`
	Token string `json:"token,omitempty"`
}

func newAdminSessionStore() *adminSessionStore {
	return &adminSessionStore{sessions: map[string]time.Time{}}
}

func (s *adminSessionStore) create() (string, error) {
	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = time.Now().UTC().Add(24 * time.Hour)
	return token, nil
}

func (s *adminSessionStore) valid(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.sessions[token]
	if !ok {
		return false
	}
	if time.Now().UTC().After(expiresAt) {
		delete(s.sessions, token)
		return false
	}
	return true
}

func (s *adminSessionStore) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, strings.TrimSpace(token))
}

func (s *adminSessionStore) expiresAt(token string) (time.Time, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return time.Time{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.sessions[token]
	if !ok {
		return time.Time{}, false
	}
	if time.Now().UTC().After(expiresAt) {
		delete(s.sessions, token)
		return time.Time{}, false
	}
	return expiresAt, true
}

func (h *Handler) registerAdminRoutes(mux *stdhttp.ServeMux) {
	mux.HandleFunc("/admin", h.adminPage)
	mux.HandleFunc("/api/admin/meta", h.adminAPI(h.adminMeta))
	mux.HandleFunc("/api/admin/auth/login", h.adminAPI(h.adminLogin))
	mux.HandleFunc("/api/admin/auth/logout", h.adminAPI(h.requireAdmin(h.adminLogout)))
	mux.HandleFunc("/api/admin/auth/session", h.adminAPI(h.requireAdmin(h.adminSession)))
	mux.HandleFunc("/admin/api/login", h.adminAPI(h.adminLogin))
	mux.HandleFunc("/admin/api/logout", h.adminAPI(h.requireAdmin(h.adminLogout)))
	mux.HandleFunc("/admin/api/status", h.adminAPI(h.requireAdmin(h.adminStatus)))
	mux.HandleFunc("/admin/api/settings", h.adminAPI(h.requireAdmin(h.adminSettings)))
	mux.HandleFunc("/admin/api/providers/cursor/config", h.adminAPI(h.requireAdmin(h.adminCursorConfig)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts/list", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts/create", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts/detail/", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts/update/", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/kiro/accounts/delete/", h.adminAPI(h.requireAdmin(h.adminKiroAccounts)))
	mux.HandleFunc("/admin/api/providers/grok/config", h.adminAPI(h.requireAdmin(h.adminGrokConfig)))
	mux.HandleFunc("/admin/api/providers/grok/tokens", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/grok/tokens/list", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/grok/tokens/create", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/grok/tokens/detail/", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/grok/tokens/update/", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/grok/tokens/delete/", h.adminAPI(h.requireAdmin(h.adminGrokTokens)))
	mux.HandleFunc("/admin/api/providers/orchids/config", h.adminAPI(h.requireAdmin(h.adminOrchidsConfig)))
	mux.HandleFunc("/admin/api/providers/web/config", h.adminAPI(h.requireAdmin(h.adminWebConfig)))
	mux.HandleFunc("/admin/api/providers/chatgpt/config", h.adminAPI(h.requireAdmin(h.adminChatGPTConfig)))
	mux.HandleFunc("/admin/api/providers/blink/config", h.adminAPI(h.requireAdmin(h.adminBlinkConfig)))
	mux.HandleFunc("/admin/api/providers/zai/image/config", h.adminAPI(h.requireAdmin(h.adminZAIImageConfig)))
	mux.HandleFunc("/admin/api/providers/zai/tts/config", h.adminAPI(h.requireAdmin(h.adminZAITTSConfig)))
	mux.HandleFunc("/admin/api/providers/zai/ocr/config", h.adminAPI(h.requireAdmin(h.adminZAIOCRConfig)))
}

func (h *Handler) adminPage(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(adminIndexHTML))
}

func (h *Handler) adminMeta(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	response := adminMetaResponse{Features: sharedAdminFeatures()}
	response.Backend.Language = adminBackendLanguage
	response.Backend.Version = adminBackendVersion
	response.Auth.Mode = adminAuthModeSession
	h.writeJSON(w, stdhttp.StatusOK, response)
}

func (h *Handler) adminLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var payload struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(payload.Password) != h.runtime.AdminPassword() {
		h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]string{"error": "invalid admin password"})
		return
	}
	token, err := h.sessions.create()
	if err != nil {
		h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("create session: %v", err)})
		return
	}
	stdhttp.SetCookie(w, &stdhttp.Cookie{
		Name:     adminSessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: stdhttp.SameSiteLaxMode,
		MaxAge:   86400,
	})
	h.writeJSON(w, stdhttp.StatusOK, adminLoginResponse{OK: true, Token: token})
}

func (h *Handler) adminLogout(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if token, ok := h.adminSessionToken(r); ok {
		h.sessions.delete(token)
	}
	stdhttp.SetCookie(w, &stdhttp.Cookie{Name: adminSessionCookieName, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: stdhttp.SameSiteLaxMode})
	h.writeJSON(w, stdhttp.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) adminSession(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet {
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	token, ok := h.adminSessionToken(r)
	if !ok {
		h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]string{"error": "admin login required"})
		return
	}
	expiresAt, ok := h.sessions.expiresAt(token)
	if !ok {
		h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]string{"error": "admin login required"})
		return
	}
	response := adminSessionResponse{Authenticated: true, ExpiresAt: expiresAt.UTC().Format(time.RFC3339)}
	response.User.ID = adminSessionUserID
	response.User.Name = adminSessionUserName
	response.User.Role = adminSessionUserRole
	h.writeJSON(w, stdhttp.StatusOK, response)
}

func (h *Handler) adminStatus(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	data := h.runtime.Snapshot()
	cfg := h.currentConfig()
	blinkReady := blinkConfigured(cfg.Blink)
	h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{
		"project": "any2api-go",
		"settings": adminSettingsResponse{
			APIKey:                  data.Settings.APIKey,
			DefaultProvider:         data.Settings.DefaultProvider,
			AdminPasswordConfigured: strings.TrimSpace(data.Settings.AdminPassword) != "",
		},
		"providers": map[string]interface{}{
			"cursor":   map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.Cursor.Cookie) != ""), "configured": strings.TrimSpace(cfg.Cursor.Cookie) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.Cursor.Cookie) != "")},
			"kiro":     map[string]interface{}{"count": len(data.Providers.KiroAccounts), "configured": strings.TrimSpace(cfg.Kiro.AccessToken) != "", "active": activeKiroID(data.Providers.KiroAccounts)},
			"grok":     map[string]interface{}{"count": len(data.Providers.GrokTokens), "configured": strings.TrimSpace(cfg.Grok.CookieToken) != "", "active": activeGrokID(data.Providers.GrokTokens)},
			"orchids":  map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.Orchids.ClientCookie) != ""), "configured": strings.TrimSpace(cfg.Orchids.ClientCookie) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.Orchids.ClientCookie) != "")},
			"web":      map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.Web.BaseURL) != "" && strings.TrimSpace(cfg.Web.Type) != ""), "configured": strings.TrimSpace(cfg.Web.BaseURL) != "" && strings.TrimSpace(cfg.Web.Type) != "", "active": activeWebType(cfg.Web.Type)},
			"chatgpt":  map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.ChatGPT.Token) != ""), "configured": strings.TrimSpace(cfg.ChatGPT.Token) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.ChatGPT.Token) != "")},
			"blink":    map[string]interface{}{"count": boolCount(blinkReady), "configured": blinkReady, "active": activeBlinkLabel(cfg.Blink)},
			"zaiImage": map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.ZAIImage.SessionToken) != ""), "configured": strings.TrimSpace(cfg.ZAIImage.SessionToken) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.ZAIImage.SessionToken) != "")},
			"zaiTTS":   map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.ZAITTS.Token) != "" && strings.TrimSpace(cfg.ZAITTS.UserID) != ""), "configured": strings.TrimSpace(cfg.ZAITTS.Token) != "" && strings.TrimSpace(cfg.ZAITTS.UserID) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.ZAITTS.Token) != "" && strings.TrimSpace(cfg.ZAITTS.UserID) != "")},
			"zaiOCR":   map[string]interface{}{"count": boolCount(strings.TrimSpace(cfg.ZAIOCR.Token) != ""), "configured": strings.TrimSpace(cfg.ZAIOCR.Token) != "", "active": providerActiveLabel(strings.TrimSpace(cfg.ZAIOCR.Token) != "")},
		},
	})
}

func (h *Handler) adminSettings(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		data := h.runtime.Snapshot()
		h.writeJSON(w, stdhttp.StatusOK, adminSettingsResponse{
			APIKey:                  data.Settings.APIKey,
			DefaultProvider:         data.Settings.DefaultProvider,
			AdminPasswordConfigured: strings.TrimSpace(data.Settings.AdminPassword) != "",
		})
	case stdhttp.MethodPut:
		var payload adminSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.UpdateSettings(payload.APIKey, payload.DefaultProvider, payload.AdminPassword)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, adminSettingsResponse{
			APIKey:                  data.Settings.APIKey,
			DefaultProvider:         data.Settings.DefaultProvider,
			AdminPasswordConfigured: strings.TrimSpace(data.Settings.AdminPassword) != "",
		})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminKiroAccounts(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	const (
		legacyPath = "/admin/api/providers/kiro/accounts"
		listPath   = "/admin/api/providers/kiro/accounts/list"
		createPath = "/admin/api/providers/kiro/accounts/create"
		detailBase = "/admin/api/providers/kiro/accounts/detail/"
		updateBase = "/admin/api/providers/kiro/accounts/update/"
		deleteBase = "/admin/api/providers/kiro/accounts/delete/"
	)
	path := r.URL.Path
	switch {
	case path == legacyPath || path == listPath:
		switch r.Method {
		case stdhttp.MethodGet:
			h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"accounts": h.runtime.Snapshot().Providers.KiroAccounts})
		case stdhttp.MethodPut:
			if path != legacyPath {
				h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			var payload struct {
				Accounts []core.KiroAccount `json:"accounts"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
			data, err := h.runtime.ReplaceKiroAccounts(payload.Accounts)
			if err != nil {
				h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"accounts": data.Providers.KiroAccounts})
		default:
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	case path == createPath:
		if r.Method != stdhttp.MethodPost {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var payload core.KiroAccount
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		account, err := h.runtime.CreateKiroAccount(payload)
		if err != nil {
			status := stdhttp.StatusInternalServerError
			if err.Error() == "invalid kiro account" {
				status = stdhttp.StatusBadRequest
			}
			h.writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"account": account})
	case strings.HasPrefix(path, detailBase):
		if r.Method != stdhttp.MethodGet {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, detailBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		account, found := h.runtime.KiroAccount(id)
		if !found {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"account": account})
	case strings.HasPrefix(path, updateBase):
		if r.Method != stdhttp.MethodPut {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, updateBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		var payload core.KiroAccount
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		account, found, err := h.runtime.UpdateKiroAccount(id, payload)
		if err != nil {
			status := stdhttp.StatusInternalServerError
			if err.Error() == "invalid kiro account" {
				status = stdhttp.StatusBadRequest
			}
			h.writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		if !found {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"account": account})
	case strings.HasPrefix(path, deleteBase):
		if r.Method != stdhttp.MethodDelete {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, deleteBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		deleted, err := h.runtime.DeleteKiroAccount(id)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if !deleted {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]bool{"ok": true})
	default:
		h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func (h *Handler) adminCursorConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.CursorConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.CursorRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceCursorConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.CursorConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminGrokTokens(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	const (
		legacyPath = "/admin/api/providers/grok/tokens"
		listPath   = "/admin/api/providers/grok/tokens/list"
		createPath = "/admin/api/providers/grok/tokens/create"
		detailBase = "/admin/api/providers/grok/tokens/detail/"
		updateBase = "/admin/api/providers/grok/tokens/update/"
		deleteBase = "/admin/api/providers/grok/tokens/delete/"
	)
	path := r.URL.Path
	switch {
	case path == legacyPath || path == listPath:
		switch r.Method {
		case stdhttp.MethodGet:
			h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"tokens": h.runtime.Snapshot().Providers.GrokTokens})
		case stdhttp.MethodPut:
			if path != legacyPath {
				h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			var payload struct {
				Tokens []core.GrokToken `json:"tokens"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
			data, err := h.runtime.ReplaceGrokTokens(payload.Tokens)
			if err != nil {
				h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"tokens": data.Providers.GrokTokens})
		default:
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	case path == createPath:
		if r.Method != stdhttp.MethodPost {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var payload core.GrokToken
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		token, err := h.runtime.CreateGrokToken(payload)
		if err != nil {
			status := stdhttp.StatusInternalServerError
			if err.Error() == "invalid grok token" {
				status = stdhttp.StatusBadRequest
			}
			h.writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"token": token})
	case strings.HasPrefix(path, detailBase):
		if r.Method != stdhttp.MethodGet {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, detailBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		token, found := h.runtime.GrokToken(id)
		if !found {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"token": token})
	case strings.HasPrefix(path, updateBase):
		if r.Method != stdhttp.MethodPut {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, updateBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		var payload core.GrokToken
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		token, found, err := h.runtime.UpdateGrokToken(id, payload)
		if err != nil {
			status := stdhttp.StatusInternalServerError
			if err.Error() == "invalid grok token" {
				status = stdhttp.StatusBadRequest
			}
			h.writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		if !found {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"token": token})
	case strings.HasPrefix(path, deleteBase):
		if r.Method != stdhttp.MethodDelete {
			h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id, ok := adminPathID(path, deleteBase)
		if !ok {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		deleted, err := h.runtime.DeleteGrokToken(id)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if !deleted {
			h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]bool{"ok": true})
	default:
		h.writeJSON(w, stdhttp.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func (h *Handler) adminGrokConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.GrokConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.GrokRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceGrokConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.GrokConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminOrchidsConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.OrchidsConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.OrchidsRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceOrchidsConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.OrchidsConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminWebConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.WebConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.WebRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceWebConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.WebConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminChatGPTConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.ChatGPTConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.ChatGPTRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceChatGPTConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.ChatGPTConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminBlinkConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.BlinkConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.BlinkRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceBlinkConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.BlinkConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminZAIImageConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.ZAIImageConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.ZAIImageRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceZAIImageConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.ZAIImageConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminZAITTSConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.ZAITTSConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.ZAITTSRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceZAITTSConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.ZAITTSConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) adminZAIOCRConfig(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodGet:
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": h.runtime.Snapshot().Providers.ZAIOCRConfig})
	case stdhttp.MethodPut:
		var payload struct {
			Config core.ZAIOCRRuntimeConfig `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			h.writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		data, err := h.runtime.ReplaceZAIOCRConfig(payload.Config)
		if err != nil {
			h.writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		h.writeJSON(w, stdhttp.StatusOK, map[string]interface{}{"config": data.Providers.ZAIOCRConfig})
	default:
		h.writeJSON(w, stdhttp.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Handler) requireAdmin(next stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		token, ok := h.adminSessionToken(r)
		if !ok || !h.sessions.valid(token) {
			h.writeJSON(w, stdhttp.StatusUnauthorized, map[string]string{"error": "admin login required"})
			return
		}
		next(w, r)
	}
}

func (h *Handler) adminAPI(next stdhttp.HandlerFunc) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.writeAdminCORSHeaders(w, r)
		if r.Method == stdhttp.MethodOptions {
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (h *Handler) writeAdminCORSHeaders(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func adminPathID(path string, prefix string) (string, bool) {
	raw := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if raw == "" || strings.Contains(raw, "/") {
		return "", false
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", false
	}
	decoded = strings.TrimSpace(decoded)
	if decoded == "" {
		return "", false
	}
	return decoded, true
}

func (h *Handler) adminSessionToken(r *stdhttp.Request) (string, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token != "" {
			return token, true
		}
	}
	if cookie, err := r.Cookie(adminSessionCookieName); err == nil {
		token := strings.TrimSpace(cookie.Value)
		if token != "" {
			return token, true
		}
	}
	return "", false
}

func activeKiroID(accounts []core.KiroAccount) string {
	for _, account := range accounts {
		if account.Active {
			return account.ID
		}
	}
	return ""
}

func activeGrokID(tokens []core.GrokToken) string {
	for _, token := range tokens {
		if token.Active {
			return token.ID
		}
	}
	return ""
}

func activeWebType(typeName string) string {
	trimmed := strings.TrimSpace(typeName)
	if trimmed == "" {
		return ""
	}
	return trimmed
}

func blinkConfigured(cfg core.BlinkConfig) bool {
	if strings.TrimSpace(cfg.RefreshToken) != "" {
		return true
	}
	return strings.TrimSpace(cfg.IDToken) != "" && strings.TrimSpace(cfg.SessionToken) != ""
}

func activeBlinkLabel(cfg core.BlinkConfig) string {
	if !blinkConfigured(cfg) {
		return ""
	}
	if slug := strings.TrimSpace(cfg.WorkspaceSlug); slug != "" {
		return slug
	}
	if projectID := strings.TrimSpace(cfg.ProjectID); projectID != "" {
		return projectID
	}
	if strings.TrimSpace(cfg.RefreshToken) != "" {
		return "refresh-token"
	}
	return "direct-session"
}

func providerActiveLabel(configured bool) string {
	if configured {
		return "default"
	}
	return ""
}

func boolCount(v bool) int {
	if v {
		return 1
	}
	return 0
}

func sharedAdminFeatures() adminFeatures {
	return adminFeatures{
		Providers:          true,
		Credentials:        true,
		ProviderState:      true,
		Stats:              false,
		Logs:               false,
		Users:              false,
		ConfigImportExport: false,
	}
}
