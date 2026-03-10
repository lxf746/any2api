/* ═══════════════════════════════════════════════════════════════
   Any2API Desktop Admin — Main Application
   ═══════════════════════════════════════════════════════════════ */

const DEFAULT_BACKEND_URL = "http://127.0.0.1:8099";
// Preserve legacy localStorage keys so existing desktop settings continue to work.
const BACKEND_STORAGE_KEY = "newplatform2api.desktop.backendUrl";
const TOKEN_STORAGE_PREFIX = "newplatform2api.desktop.adminToken";
const THEME_STORAGE_KEY = "newplatform2api.desktop.theme";
const SIDEBAR_STORAGE_KEY = "newplatform2api.desktop.sidebarCollapsed";

// ── Types ──

type AdminFeatures = {
  providers: boolean;
  credentials: boolean;
  providerState: boolean;
  stats: boolean;
  logs: boolean;
  users: boolean;
  configImportExport: boolean;
};

type AdminMeta = {
  backend: { language: string; version: string };
  auth: { mode: string };
  features: Partial<AdminFeatures>;
};

type AdminSession = {
  authenticated: boolean;
  expiresAt?: string;
  user: { id: string; name: string; role: string };
};

type AdminLoginResponse = { ok: boolean; token?: string };
type AdminSettings = { apiKey: string; defaultProvider: string; adminPasswordConfigured: boolean };
type ProviderStatus = { count: number; configured: boolean; active?: string };
type AdminStatus = {
  providers: {
    cursor?: ProviderStatus;
    kiro?: ProviderStatus;
    grok?: ProviderStatus;
    orchids?: ProviderStatus;
    web?: ProviderStatus;
    chatgpt?: ProviderStatus;
    zaiImage?: ProviderStatus;
    zaiTTS?: ProviderStatus;
    zaiOCR?: ProviderStatus;
  };
};

type CursorConfig = {
  apiUrl?: string; scriptUrl?: string; cookie?: string; xIsHuman?: string;
  userAgent?: string; referer?: string; webglVendor?: string; webglRenderer?: string;
};

type KiroAccount = {
  id?: string; name?: string; accessToken?: string; machineId?: string;
  preferredEndpoint?: string; active?: boolean;
};

type ImportedKiroAccount = {
  id?: string; name?: string; email?: string; accessToken?: string;
  machineId?: string; machineID?: string; preferredEndpoint?: string;
  credentials?: { accessToken?: string; machineId?: string };
  active?: boolean;
};

type GrokToken = { id?: string; name?: string; cookieToken?: string; active?: boolean };

type ImportedGrokToken = {
  id?: string; name?: string; cookieToken?: string; token?: string; value?: string; active?: boolean;
};

type GrokConfig = {
  apiUrl?: string; proxyUrl?: string; cfCookies?: string; cfClearance?: string;
  userAgent?: string; origin?: string; referer?: string;
};

type OrchidsConfig = {
  apiUrl?: string; clerkUrl?: string; clientCookie?: string; clientUat?: string;
  sessionId?: string; projectId?: string; userId?: string; email?: string; agentMode?: string;
};

type WebConfig = { baseUrl?: string; type?: string; apiKey?: string };
type ChatGPTConfig = { baseUrl?: string; token?: string };
type ZaiImageConfig = { sessionToken?: string; apiUrl?: string };
type ZaiTTSConfig = { token?: string; userId?: string; apiUrl?: string };
type ZaiOCRConfig = { token?: string; apiUrl?: string };

type APIOptions = { method?: string; body?: string; headers?: Record<string, string> };
type ModalProvider = "cursor" | "grok" | "orchids" | "claude" | "chatgpt";
type EntryModalProvider = "kiro" | "grok";
type NavSection = "management" | "dialog" | "multimedia" | "system";

const defaultFeatures: AdminFeatures = {
  providers: false, credentials: false, providerState: false,
  stats: false, logs: false, users: false, configImportExport: false,
};

const ADMIN_TABS = ["overview", "providers", "logs", "users", "cursor", "kiro", "grok", "orchids", "claude", "chatgpt", "zai", "zai-image", "zai-tts", "zai-ocr", "settings"] as const;
type AdminTab = typeof ADMIN_TABS[number];

const NAV_SECTIONS: Record<NavSection, AdminTab[]> = {
  management: ["overview", "providers", "logs", "users"],
  dialog: ["cursor", "kiro", "grok", "orchids", "claude", "chatgpt", "zai"],
  multimedia: ["zai-image", "zai-tts", "zai-ocr"],
  system: ["settings"],
};

// ── Utilities ──

function normalizeBaseUrl(input: string): string {
  const trimmed = input.trim().replace(/\/+$/, "").replace(/\/admin$/, "");
  return trimmed || DEFAULT_BACKEND_URL;
}

function tokenStorageKey(baseUrl: string): string {
  return `${TOKEN_STORAGE_PREFIX}:${normalizeBaseUrl(baseUrl)}`;
}

function escapeHtml(value: unknown): string {
  return String(value ?? "").replace(/[&<>"']/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] ?? c));
}

function trimValue(value?: string): string {
  return String(value ?? "").trim();
}

function hasValue(value?: string): boolean {
  return trimValue(value) !== "";
}

function displayProviderName(providerId: string): string {
  return ({ web: "Claude", chatgpt: "ChatGPT", "zai-image": "Z.ai Image", "zai-tts": "Z.ai TTS", "zai-ocr": "Z.ai OCR" } as Record<string, string>)[providerId] || providerId;
}

function maskSecret(secret?: string): string {
  if (!secret) return "未设置";
  if (secret.length <= 16) return `${secret.slice(0, 4)}****`;
  return `${secret.slice(0, 8)}…${secret.slice(-8)}`;
}

function must<T extends HTMLElement>(selector: string): T {
  const el = document.querySelector<T>(selector);
  if (!el) throw new Error(`missing element: ${selector}`);
  return el;
}

// ── Theme System ──

function initTheme(): string {
  const saved = localStorage.getItem(THEME_STORAGE_KEY);
  const theme = saved === "light" ? "light" : "dark";
  document.documentElement.setAttribute("data-theme", theme);
  updateThemeIcons(theme);
  return theme;
}

function toggleTheme(): string {
  const current = document.documentElement.getAttribute("data-theme") || "dark";
  const next = current === "dark" ? "light" : "dark";
  document.documentElement.setAttribute("data-theme", next);
  localStorage.setItem(THEME_STORAGE_KEY, next);
  updateThemeIcons(next);
  return next;
}

function updateThemeIcons(theme: string) {
  const darkIcon = document.getElementById("theme-icon-dark");
  const lightIcon = document.getElementById("theme-icon-light");
  if (darkIcon) darkIcon.classList.toggle("hidden", theme === "light");
  if (lightIcon) lightIcon.classList.toggle("hidden", theme === "dark");
}

// ── Sidebar ──

function initSidebar(): boolean {
  const collapsed = localStorage.getItem(SIDEBAR_STORAGE_KEY) === "true";
  const sidebar = document.getElementById("sidebar");
  if (sidebar && collapsed) sidebar.classList.add("collapsed");
  return collapsed;
}

function toggleSidebar() {
  const sidebar = document.getElementById("sidebar");
  if (!sidebar) return;
  const collapsed = sidebar.classList.toggle("collapsed");
  localStorage.setItem(SIDEBAR_STORAGE_KEY, String(collapsed));
}

function toggleMobileSidebar(open?: boolean) {
  const sidebar = document.getElementById("sidebar");
  const overlay = document.getElementById("sidebar-overlay");
  if (!sidebar || !overlay) return;
  const isOpen = open ?? !sidebar.classList.contains("mobile-open");
  sidebar.classList.toggle("mobile-open", isOpen);
  overlay.classList.toggle("visible", isOpen);
}

// ── Main App ──

window.addEventListener("DOMContentLoaded", () => {
  initTheme();
  initSidebar();

  // DOM refs
  const backendInput = must<HTMLInputElement>("#backend-url");
  const adminUrl = must<HTMLElement>("#admin-url");
  const statusText = must<HTMLElement>("#status-text");
  const loginCard = must<HTMLElement>("#login-card");
  const appShell = must<HTMLElement>("#app-shell");
  const sessionChip = must<HTMLElement>("#session-chip");
  const sessionDot = must<HTMLElement>("#session-dot");
  const authChip = must<HTMLElement>("#auth-chip");
  const loginSubtitle = must<HTMLElement>("#login-subtitle");
  const sidebarSubtitle = must<HTMLElement>("#sidebar-subtitle");
  const statusGrid = must<HTMLElement>("#status-grid");
  const loginPasswordInput = must<HTMLInputElement>("#login-password");
  const apiKeyInput = must<HTMLInputElement>("#api-key");
  const defaultProviderSelect = must<HTMLSelectElement>("#default-provider");
  const adminPasswordInput = must<HTMLInputElement>("#admin-password");
  const cursorAPIURLInput = must<HTMLInputElement>("#cursor-api-url");
  const cursorScriptURLInput = must<HTMLInputElement>("#cursor-script-url");
  const cursorXIsHumanInput = must<HTMLInputElement>("#cursor-x-is-human");
  const cursorUserAgentInput = must<HTMLInputElement>("#cursor-user-agent");
  const cursorRefererInput = must<HTMLInputElement>("#cursor-referer");
  const cursorWebGLVendorInput = must<HTMLInputElement>("#cursor-webgl-vendor");
  const cursorWebGLRendererInput = must<HTMLInputElement>("#cursor-webgl-renderer");
  const cursorCookieInput = must<HTMLTextAreaElement>("#cursor-cookie");
  const kiroList = must<HTMLElement>("#kiro-list");
  const grokList = must<HTMLElement>("#grok-list");
  const grokAPIURLInput = must<HTMLInputElement>("#grok-api-url");
  const grokProxyURLInput = must<HTMLInputElement>("#grok-proxy-url");
  const grokUserAgentInput = must<HTMLInputElement>("#grok-user-agent");
  const grokOriginInput = must<HTMLInputElement>("#grok-origin");
  const grokRefererInput = must<HTMLInputElement>("#grok-referer");
  const grokCFClearanceInput = must<HTMLInputElement>("#grok-cf-clearance");
  const grokCFCookiesInput = must<HTMLTextAreaElement>("#grok-cf-cookies");
  const kiroImport = must<HTMLTextAreaElement>("#kiro-import");
  const grokImport = must<HTMLTextAreaElement>("#grok-import");
  const orchidsAPIURLInput = must<HTMLInputElement>("#orchids-api-url");
  const orchidsClerkURLInput = must<HTMLInputElement>("#orchids-clerk-url");
  const orchidsAgentModeInput = must<HTMLInputElement>("#orchids-agent-mode");
  const orchidsClientUATInput = must<HTMLInputElement>("#orchids-client-uat");
  const orchidsSessionIDInput = must<HTMLInputElement>("#orchids-session-id");
  const orchidsProjectIDInput = must<HTMLInputElement>("#orchids-project-id");
  const orchidsUserIDInput = must<HTMLInputElement>("#orchids-user-id");
  const orchidsEmailInput = must<HTMLInputElement>("#orchids-email");
  const orchidsClientCookieInput = must<HTMLTextAreaElement>("#orchids-client-cookie");
  const claudeBaseURLInput = must<HTMLInputElement>("#claude-base-url");
  const claudeTypeInput = must<HTMLInputElement>("#claude-type");
  const claudeAPIKeyInput = must<HTMLTextAreaElement>("#claude-api-key");
  const chatGPTBaseURLInput = must<HTMLInputElement>("#chatgpt-base-url");
  const chatGPTTokenInput = must<HTMLTextAreaElement>("#chatgpt-token");
  const zaiImageAPIURLInput = must<HTMLInputElement>("#zai-image-api-url");
  const zaiImageSessionTokenInput = must<HTMLTextAreaElement>("#zai-image-session-token");
  const zaiTTSAPIURLInput = must<HTMLInputElement>("#zai-tts-api-url");
  const zaiTTSUserIDInput = must<HTMLInputElement>("#zai-tts-user-id");
  const zaiTTSTokenInput = must<HTMLTextAreaElement>("#zai-tts-token");
  const zaiOCRAPIURLInput = must<HTMLInputElement>("#zai-ocr-api-url");
  const zaiOCRTokenInput = must<HTMLTextAreaElement>("#zai-ocr-token");
  const flash = must<HTMLElement>("#flash");
  const providerGrid = must<HTMLElement>("#provider-grid");
  const cursorSummary = must<HTMLElement>("#cursor-summary");
  const orchidsSummary = must<HTMLElement>("#orchids-summary");
  const claudeSummary = must<HTMLElement>("#claude-summary");
  const chatGPTSummary = must<HTMLElement>("#chatgpt-summary");
  const kiroImportCard = must<HTMLElement>("#kiro-import-card");
  const grokImportCard = must<HTMLElement>("#grok-import-card");
  const configModalOverlay = must<HTMLElement>("#config-modal-overlay");
  const configModal = must<HTMLElement>("#config-modal");
  const configModalTitle = must<HTMLElement>("#config-modal-title");
  const configModalDescription = must<HTMLElement>("#config-modal-description");
  const configModalSaveButton = must<HTMLButtonElement>("#config-modal-save-btn");
  const configModalCloseButton = must<HTMLButtonElement>("#config-modal-close");
  const configModalCancelButton = must<HTMLButtonElement>("#config-modal-cancel");
  const configModalSections = Array.from(document.querySelectorAll<HTMLElement>("[data-config-section]"));
  const entryModalOverlay = must<HTMLElement>("#entry-modal-overlay");
  const entryModal = must<HTMLElement>("#entry-modal");
  const entryModalTitle = must<HTMLElement>("#entry-modal-title");
  const entryModalDescription = must<HTMLElement>("#entry-modal-description");
  const entryModalSaveButton = must<HTMLButtonElement>("#entry-modal-save-btn");
  const entryModalCloseButton = must<HTMLButtonElement>("#entry-modal-close");
  const entryModalCancelButton = must<HTMLButtonElement>("#entry-modal-cancel");
  const entryModalSections = Array.from(document.querySelectorAll<HTMLElement>("[data-entry-section]"));
  const kiroEntryNameInput = must<HTMLInputElement>("#kiro-entry-name");
  const kiroEntryMachineIDInput = must<HTMLInputElement>("#kiro-entry-machine-id");
  const kiroEntryEndpointInput = must<HTMLSelectElement>("#kiro-entry-endpoint");
  const kiroEntryActiveInput = must<HTMLInputElement>("#kiro-entry-active");
  const kiroEntryAccessTokenInput = must<HTMLTextAreaElement>("#kiro-entry-access-token");
  const grokEntryNameInput = must<HTMLInputElement>("#grok-entry-name");
  const grokEntryActiveInput = must<HTMLInputElement>("#grok-entry-active");
  const grokEntryCookieTokenInput = must<HTMLTextAreaElement>("#grok-entry-cookie-token");

  // State
  const state = {
    baseUrl: normalizeBaseUrl(localStorage.getItem(BACKEND_STORAGE_KEY) || DEFAULT_BACKEND_URL),
    currentTab: "overview" as AdminTab,
    meta: null as AdminMeta | null,
    session: null as AdminSession | null,
    sessionToken: "",
    status: null as AdminStatus | null,
    settings: null as AdminSettings | null,
    cursorConfig: {} as CursorConfig,
    kiroAccounts: [] as KiroAccount[],
    grokConfig: {} as GrokConfig,
    grokTokens: [] as GrokToken[],
    orchidsConfig: {} as OrchidsConfig,
    webConfig: {} as WebConfig,
    chatgptConfig: {} as ChatGPTConfig,
    zaiImageConfig: {} as ZaiImageConfig,
    zaiTTSConfig: {} as ZaiTTSConfig,
    zaiOCRConfig: {} as ZaiOCRConfig,
    configModalProvider: null as ModalProvider | null,
    entryModalProvider: null as EntryModalProvider | null,
  };

  // ── Toast ──
  let flashTimer: ReturnType<typeof setTimeout> | null = null;
  const toast = (text: string, type: "info" | "error" | "success" = "info") => {
    if (flashTimer) clearTimeout(flashTimer);
    if (!text) { flash.className = "flash hidden"; flash.textContent = ""; return; }
    flash.className = `flash ${type === "info" ? "" : type}`.trim();
    flash.textContent = text;
    flashTimer = setTimeout(() => { flash.className = "flash hidden"; flash.textContent = ""; }, 3000);
  };

  const setStatus = (text: string) => {
    adminUrl.textContent = state.baseUrl;
    statusText.textContent = text;
  };

  const getFeatures = (): AdminFeatures => ({ ...defaultFeatures, ...(state.meta?.features || {}) });
  const authModeLabel = (): string => state.meta?.auth?.mode || "unknown";
  const backendLabel = (): string => {
    const lang = state.meta?.backend?.language || "unknown";
    const ver = state.meta?.backend?.version || "dev";
    return `${String(lang).toUpperCase()} / ${ver}`;
  };
  const supportsProviderOverview = (): boolean => {
    const f = getFeatures();
    return !!(f.providers || f.providerState);
  };
  const supportsProviderCredentials = (): boolean => {
    const f = getFeatures();
    return !!(f.providers && f.credentials);
  };
  const supportsSettings = (): boolean => {
    const f = getFeatures();
    return !!(f.providers || f.credentials || f.providerState);
  };
  const enabledFeatureCount = (): number => Object.values(getFeatures()).filter(Boolean).length;
  const availableTabs = (): AdminTab[] => ADMIN_TABS.filter((tab) => isTabEnabled(tab));
  const defaultProviderLabel = (): string => displayProviderName(state.settings?.defaultProvider || "cursor");

  type SummaryItem = { label: string; value: string; hint?: string };
  type ProviderDescriptor = {
    key: string;
    label: string;
    initial: string;
    tab: AdminTab;
    group: "对话类" | "多媒体";
    mode: "列表" | "配置" | "预留";
    includeInSummary?: boolean;
  };

  const providerDescriptors = (): ProviderDescriptor[] => ([
    { key: "cursor", label: "Cursor", initial: "C", tab: "cursor", group: "对话类", mode: "列表" },
    { key: "kiro", label: "Kiro", initial: "K", tab: "kiro", group: "对话类", mode: "列表" },
    { key: "grok", label: "Grok", initial: "G", tab: "grok", group: "对话类", mode: "列表" },
    { key: "orchids", label: "Orchids", initial: "O", tab: "orchids", group: "对话类", mode: "列表" },
    { key: "web", label: "Claude", initial: "Cl", tab: "claude", group: "对话类", mode: "列表" },
    { key: "chatgpt", label: "ChatGPT", initial: "CG", tab: "chatgpt", group: "对话类", mode: "列表" },
    { key: "zai", label: "Z.ai", initial: "Z", tab: "zai", group: "对话类", mode: "预留", includeInSummary: false },
    { key: "zaiImage", label: "Z.ai Image", initial: "ZI", tab: "zai-image", group: "多媒体", mode: "配置" },
    { key: "zaiTTS", label: "Z.ai TTS", initial: "ZT", tab: "zai-tts", group: "多媒体", mode: "配置" },
    { key: "zaiOCR", label: "Z.ai OCR", initial: "ZO", tab: "zai-ocr", group: "多媒体", mode: "配置" },
  ]);

  const persistToken = (token: string) => {
    state.sessionToken = token;
    if (token) localStorage.setItem(tokenStorageKey(state.baseUrl), token);
    else localStorage.removeItem(tokenStorageKey(state.baseUrl));
  };

  const setLoggedOutState = () => {
    state.session = null;
    loginCard.classList.remove("hidden");
    appShell.classList.add("hidden");
    sessionChip.textContent = "未登录";
    sessionDot.classList.remove("online");
    authChip.textContent = authModeLabel();
    setStatus(`已连接 ${state.baseUrl}，等待登录`);
  };

  const setAuthenticatedState = (session: AdminSession | null) => {
    state.session = session;
    loginCard.classList.add("hidden");
    appShell.classList.remove("hidden");
    sessionChip.textContent = session?.user?.name || "已登录";
    sessionDot.classList.add("online");
    authChip.textContent = session?.user?.role || authModeLabel();
    setStatus(`${backendLabel()} · 已连接`);
  };

  // ── Tab / Feature Gating ──

  const isTabEnabled = (tab: string): boolean => {
    const f = getFeatures();
    if (tab === "overview") return true;
    if (tab === "providers") return supportsProviderOverview();
    if (tab === "logs") return !!f.logs;
    if (tab === "users") return !!f.users;
    if (["cursor", "kiro", "grok", "orchids", "claude", "chatgpt", "zai", "zai-image", "zai-tts", "zai-ocr"].includes(tab)) return supportsProviderCredentials();
    if (tab === "settings") return supportsSettings();
    return false;
  };

  const updateNavSections = () => {
    (Object.entries(NAV_SECTIONS) as Array<[NavSection, AdminTab[]]>).forEach(([section, tabs]) => {
      const el = document.querySelector<HTMLElement>(`[data-nav-section="${section}"]`);
      if (!el) return;
      el.classList.toggle("hidden", !tabs.some((tab) => isTabEnabled(tab)));
    });
  };

  const switchTab = (tab: string) => {
    const enabledTabs = availableTabs();
    if (!ADMIN_TABS.includes(tab as AdminTab) || !enabledTabs.includes(tab as AdminTab)) {
      tab = enabledTabs[0] || "overview";
    }
    state.currentTab = tab as AdminTab;
    state.configModalProvider = null;
    state.entryModalProvider = null;
    configModal.classList.add("hidden");
    configModalOverlay.classList.add("hidden");
    configModal.setAttribute("aria-hidden", "true");
    entryModal.classList.add("hidden");
    entryModalOverlay.classList.add("hidden");
    entryModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
    updateNavSections();

    document.querySelectorAll<HTMLElement>("[data-tab]").forEach((el) => {
      const targetTab = (el.dataset.tab || "") as AdminTab;
      const enabled = isTabEnabled(targetTab);
      el.classList.toggle("hidden", !enabled);
      el.classList.toggle("disabled", !enabled);
      el.classList.toggle("active", enabled && targetTab === tab);
    });

    document.querySelectorAll<HTMLElement>("[data-panel]").forEach((el) => {
      const panel = el.dataset.panel || "";
      el.classList.toggle("hidden", panel !== tab || !isTabEnabled(panel));
    });

    // Close mobile sidebar on tab switch
    toggleMobileSidebar(false);
  };

  const applyMeta = (meta: AdminMeta) => {
    state.meta = meta;
    loginSubtitle.textContent = `当前连接 ${backendLabel()}，认证模式 ${authModeLabel()}`;
    sidebarSubtitle.textContent = `${String(meta.backend.language || "unknown").toUpperCase()} · ${meta.backend.version || "dev"}`;
    authChip.textContent = authModeLabel();

    // Update feature-gated content
    const f = getFeatures();
    const chartCard = document.getElementById("overview-chart-card");
    if (chartCard) chartCard.classList.toggle("hidden", !f.stats);

    switchTab(state.currentTab || "overview");
  };

  // ── API Client ──

  async function api<T>(path: string, opts: APIOptions = {}): Promise<T> {
    const headers: Record<string, string> = { "Content-Type": "application/json", ...(opts.headers || {}) };
    if (state.sessionToken) headers.Authorization = `Bearer ${state.sessionToken}`;
    let res: Response;
    try {
      res = await fetch(`${state.baseUrl}${path}`, {
        method: opts.method || "GET", headers, body: opts.body, mode: "cors",
      });
    } catch (e) {
      throw new Error(`无法连接后台：${e instanceof Error ? e.message : "network error"}`);
    }
    const body = await res.json().catch(() => ({}));
    if (res.status === 401) throw new Error("UNAUTHORIZED");
    if (!res.ok) throw new Error(typeof body?.error === "string" ? body.error : "请求失败");
    return body as T;
  }

  async function optionalApi<T>(enabled: boolean, path: string, opts: APIOptions = {}): Promise<T | null> {
    if (!enabled) return null;
    try {
      return await api<T>(path, opts);
    } catch (e) {
      const message = e instanceof Error ? e.message : "请求失败";
      if (message === "not found" || message === "method not allowed") return null;
      throw e;
    }
  }

  // ── Helpers ──

  const downloadJSON = (filename: string, payload: unknown) => {
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url; a.download = filename;
    document.body.appendChild(a); a.click(); a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 0);
  };

  const ensureSingleActive = <T extends { active?: boolean }>(items: T[]) => {
    let found = false;
    items.forEach((item) => { if (item.active && !found) found = true; else item.active = false; });
    if (!found && items.length > 0) items[0].active = true;
  };

  const resetAdminData = () => {
    state.status = null;
    state.settings = null;
    state.cursorConfig = {};
    state.kiroAccounts = [];
    state.grokConfig = {};
    state.grokTokens = [];
    state.orchidsConfig = {};
    state.webConfig = {};
    state.chatgptConfig = {};
    state.zaiImageConfig = {};
    state.zaiTTSConfig = {};
    state.zaiOCRConfig = {};
    state.configModalProvider = null;
    state.entryModalProvider = null;
    apiKeyInput.value = "";
    defaultProviderSelect.value = "cursor";
    adminPasswordInput.value = "";
    kiroImport.value = "";
    grokImport.value = "";
    kiroEntryNameInput.value = "";
    kiroEntryMachineIDInput.value = "";
    kiroEntryEndpointInput.value = "";
    kiroEntryActiveInput.checked = false;
    kiroEntryAccessTokenInput.value = "";
    grokEntryNameInput.value = "";
    grokEntryActiveInput.checked = false;
    grokEntryCookieTokenInput.value = "";
    kiroImportCard.classList.add("hidden");
    grokImportCard.classList.add("hidden");
    must<HTMLButtonElement>("#kiro-import-toggle-btn").textContent = "导入 JSON";
    must<HTMLButtonElement>("#grok-import-toggle-btn").textContent = "导入 JSON";
    configModal.classList.add("hidden");
    configModalOverlay.classList.add("hidden");
    configModal.setAttribute("aria-hidden", "true");
    entryModal.classList.add("hidden");
    entryModalOverlay.classList.add("hidden");
    entryModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
    renderCursorConfig();
    renderGrokConfig();
    renderKiroList();
    renderGrokList();
    renderOrchidsConfig();
    renderClaudeConfig();
    renderChatGPTConfig();
    renderZaiImageConfig();
    renderZaiTTSConfig();
    renderZaiOCRConfig();
    renderStatus(null, null);
  };

  // ── Render Functions ──

  const renderSummaryGrid = (container: HTMLElement, items: SummaryItem[]) => {
    if (!items.length) {
      container.innerHTML = '<div class="empty-state compact">当前暂无可展示的配置摘要。</div>';
      return;
    }
    container.innerHTML = items.map((item) => `
      <div class="summary-item">
        <div class="summary-label">${escapeHtml(item.label)}</div>
        <div class="summary-value">${escapeHtml(item.value)}</div>
        ${item.hint ? `<div class="summary-hint">${escapeHtml(item.hint)}</div>` : ""}
      </div>`).join("");
  };

  const renderCursorConfig = () => {
    cursorAPIURLInput.value = state.cursorConfig.apiUrl || "";
    cursorScriptURLInput.value = state.cursorConfig.scriptUrl || "";
    cursorXIsHumanInput.value = state.cursorConfig.xIsHuman || "";
    cursorUserAgentInput.value = state.cursorConfig.userAgent || "";
    cursorRefererInput.value = state.cursorConfig.referer || "";
    cursorWebGLVendorInput.value = state.cursorConfig.webglVendor || "";
    cursorWebGLRendererInput.value = state.cursorConfig.webglRenderer || "";
    cursorCookieInput.value = state.cursorConfig.cookie || "";
    renderSummaryGrid(cursorSummary, [
      { label: "API URL", value: trimValue(state.cursorConfig.apiUrl) || "未设置" },
      { label: "Script URL", value: trimValue(state.cursorConfig.scriptUrl) || "未设置" },
      { label: "Cookie", value: hasValue(state.cursorConfig.cookie) ? "已设置" : "未设置", hint: hasValue(state.cursorConfig.cookie) ? maskSecret(state.cursorConfig.cookie) : "用于 Cursor 会话鉴权" },
      { label: "User Agent", value: trimValue(state.cursorConfig.userAgent) || "未设置" },
    ]);
  };

  const renderGrokConfig = () => {
    grokAPIURLInput.value = state.grokConfig.apiUrl || "";
    grokProxyURLInput.value = state.grokConfig.proxyUrl || "";
    grokUserAgentInput.value = state.grokConfig.userAgent || "";
    grokOriginInput.value = state.grokConfig.origin || "";
    grokRefererInput.value = state.grokConfig.referer || "";
    grokCFClearanceInput.value = state.grokConfig.cfClearance || "";
    grokCFCookiesInput.value = state.grokConfig.cfCookies || "";
  };

  const renderOrchidsConfig = () => {
    orchidsAPIURLInput.value = state.orchidsConfig.apiUrl || "";
    orchidsClerkURLInput.value = state.orchidsConfig.clerkUrl || "";
    orchidsAgentModeInput.value = state.orchidsConfig.agentMode || "";
    orchidsClientUATInput.value = state.orchidsConfig.clientUat || "";
    orchidsSessionIDInput.value = state.orchidsConfig.sessionId || "";
    orchidsProjectIDInput.value = state.orchidsConfig.projectId || "";
    orchidsUserIDInput.value = state.orchidsConfig.userId || "";
    orchidsEmailInput.value = state.orchidsConfig.email || "";
    orchidsClientCookieInput.value = state.orchidsConfig.clientCookie || "";
    renderSummaryGrid(orchidsSummary, [
      { label: "API URL", value: trimValue(state.orchidsConfig.apiUrl) || "未设置" },
      { label: "Clerk URL", value: trimValue(state.orchidsConfig.clerkUrl) || "未设置" },
      { label: "会话", value: hasValue(state.orchidsConfig.sessionId) ? "已设置" : "未设置", hint: hasValue(state.orchidsConfig.sessionId) ? trimValue(state.orchidsConfig.sessionId) : "等待 Session ID" },
      { label: "Client Cookie", value: hasValue(state.orchidsConfig.clientCookie) ? "已设置" : "未设置", hint: hasValue(state.orchidsConfig.clientCookie) ? maskSecret(state.orchidsConfig.clientCookie) : "用于鉴权" },
    ]);
  };

  const renderClaudeConfig = () => {
    claudeBaseURLInput.value = state.webConfig.baseUrl || "";
    claudeTypeInput.value = state.webConfig.type || "";
    claudeAPIKeyInput.value = state.webConfig.apiKey || "";
    renderSummaryGrid(claudeSummary, [
      { label: "Base URL", value: trimValue(state.webConfig.baseUrl) || "未设置" },
      { label: "类型", value: trimValue(state.webConfig.type) || "未设置", hint: "兼容现有 web provider" },
      { label: "API Key", value: hasValue(state.webConfig.apiKey) ? "已设置" : "未设置", hint: hasValue(state.webConfig.apiKey) ? maskSecret(state.webConfig.apiKey) : "可按上游要求留空" },
    ]);
  };

  const renderChatGPTConfig = () => {
    chatGPTBaseURLInput.value = state.chatgptConfig.baseUrl || "";
    chatGPTTokenInput.value = state.chatgptConfig.token || "";
    renderSummaryGrid(chatGPTSummary, [
      { label: "Base URL", value: trimValue(state.chatgptConfig.baseUrl) || "未设置" },
      { label: "Token", value: hasValue(state.chatgptConfig.token) ? "已设置" : "未设置", hint: hasValue(state.chatgptConfig.token) ? maskSecret(state.chatgptConfig.token) : "当前无访问 token" },
    ]);
  };

  const renderZaiImageConfig = () => {
    zaiImageAPIURLInput.value = state.zaiImageConfig.apiUrl || "";
    zaiImageSessionTokenInput.value = state.zaiImageConfig.sessionToken || "";
  };

  const renderZaiTTSConfig = () => {
    zaiTTSAPIURLInput.value = state.zaiTTSConfig.apiUrl || "";
    zaiTTSUserIDInput.value = state.zaiTTSConfig.userId || "";
    zaiTTSTokenInput.value = state.zaiTTSConfig.token || "";
  };

  const renderZaiOCRConfig = () => {
    zaiOCRAPIURLInput.value = state.zaiOCRConfig.apiUrl || "";
    zaiOCRTokenInput.value = state.zaiOCRConfig.token || "";
  };

  const readCursorConfig = (): CursorConfig => ({
    apiUrl: cursorAPIURLInput.value, scriptUrl: cursorScriptURLInput.value,
    xIsHuman: cursorXIsHumanInput.value, userAgent: cursorUserAgentInput.value,
    referer: cursorRefererInput.value, webglVendor: cursorWebGLVendorInput.value,
    webglRenderer: cursorWebGLRendererInput.value, cookie: cursorCookieInput.value,
  });

  const readGrokConfig = (): GrokConfig => ({
    apiUrl: grokAPIURLInput.value, proxyUrl: grokProxyURLInput.value,
    cfCookies: grokCFCookiesInput.value, cfClearance: grokCFClearanceInput.value,
    userAgent: grokUserAgentInput.value, origin: grokOriginInput.value,
    referer: grokRefererInput.value,
  });

  const readOrchidsConfig = (): OrchidsConfig => ({
    apiUrl: orchidsAPIURLInput.value, clerkUrl: orchidsClerkURLInput.value,
    agentMode: orchidsAgentModeInput.value, clientUat: orchidsClientUATInput.value,
    sessionId: orchidsSessionIDInput.value, projectId: orchidsProjectIDInput.value,
    userId: orchidsUserIDInput.value, email: orchidsEmailInput.value,
    clientCookie: orchidsClientCookieInput.value,
  });

  const readClaudeConfig = (): WebConfig => ({
    baseUrl: claudeBaseURLInput.value,
    type: claudeTypeInput.value,
    apiKey: claudeAPIKeyInput.value,
  });

  const readChatGPTConfig = (): ChatGPTConfig => ({
    baseUrl: chatGPTBaseURLInput.value,
    token: chatGPTTokenInput.value,
  });

  const readZaiImageConfig = (): ZaiImageConfig => ({
    apiUrl: zaiImageAPIURLInput.value,
    sessionToken: zaiImageSessionTokenInput.value,
  });

  const readZaiTTSConfig = (): ZaiTTSConfig => ({
    apiUrl: zaiTTSAPIURLInput.value,
    userId: zaiTTSUserIDInput.value,
    token: zaiTTSTokenInput.value,
  });

  const readZaiOCRConfig = (): ZaiOCRConfig => ({
    apiUrl: zaiOCRAPIURLInput.value,
    token: zaiOCRTokenInput.value,
  });

  const inferProviderStatus = (providerKey: string): ProviderStatus => {
    switch (providerKey) {
      case "cursor": {
        const configured = hasValue(state.cursorConfig.cookie);
        return { count: configured ? 1 : 0, configured, active: configured ? "已接入" : "" };
      }
      case "kiro": {
        const active = state.kiroAccounts.find((item) => item.active)?.name || "激活账号";
        return { count: state.kiroAccounts.length, configured: state.kiroAccounts.some((item) => hasValue(item.accessToken)), active };
      }
      case "grok": {
        const active = state.grokTokens.find((item) => item.active)?.name || "激活 Token";
        return { count: state.grokTokens.length, configured: state.grokTokens.some((item) => hasValue(item.cookieToken)), active };
      }
      case "orchids": {
        const configured = hasValue(state.orchidsConfig.clientCookie);
        return { count: configured ? 1 : 0, configured, active: configured ? "已接入" : "" };
      }
      case "web": {
        const configured = hasValue(state.webConfig.baseUrl) && hasValue(state.webConfig.type);
        return { count: configured ? 1 : 0, configured, active: trimValue(state.webConfig.type) || "已接入" };
      }
      case "chatgpt": {
        const configured = hasValue(state.chatgptConfig.baseUrl) && hasValue(state.chatgptConfig.token);
        return { count: configured ? 1 : 0, configured, active: configured ? "已接入" : "" };
      }
      case "zaiImage": {
        const configured = hasValue(state.zaiImageConfig.sessionToken);
        return { count: configured ? 1 : 0, configured, active: configured ? "已配置" : "" };
      }
      case "zaiTTS": {
        const configured = hasValue(state.zaiTTSConfig.token);
        return { count: configured ? 1 : 0, configured, active: trimValue(state.zaiTTSConfig.userId) || (configured ? "已配置" : "") };
      }
      case "zaiOCR": {
        const configured = hasValue(state.zaiOCRConfig.token);
        return { count: configured ? 1 : 0, configured, active: configured ? "已配置" : "" };
      }
      default:
        return { count: 0, configured: false, active: "预留入口" };
    }
  };

  const providerStatusFor = (status: AdminStatus | null, providerKey: string): ProviderStatus => {
    const fromBackend = status?.providers?.[providerKey as keyof AdminStatus["providers"]];
    return fromBackend || inferProviderStatus(providerKey);
  };

  const renderStatus = (status: AdminStatus | null, settings: AdminSettings | null) => {
    const descriptors = providerDescriptors();
    const summaryDescriptors = descriptors.filter((item) => item.includeInSummary !== false);
    const visibleSummaryDescriptors = supportsProviderOverview() ? summaryDescriptors : [];
    const configuredCount = visibleSummaryDescriptors.filter((item) => providerStatusFor(status, item.key).configured).length;
    const currentUser = state.session?.user?.name || "Admin";
    const providerSummary = supportsProviderOverview()
      ? `${configuredCount}/${summaryDescriptors.length || 1}`
      : "--";
    const providerMeta = settings?.defaultProvider
      ? `默认 ${escapeHtml(defaultProviderLabel())}`
      : supportsProviderOverview()
        ? "等待 Provider 数据"
        : "当前后端未启用 Provider 管理";

    statusGrid.innerHTML = `
      <div class="stat-card">
        <div class="stat-label">后端</div>
        <div class="stat-value" style="font-size:18px">${escapeHtml(backendLabel())}</div>
        <div class="stat-meta">认证 ${escapeHtml(authModeLabel())}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">会话</div>
        <div class="stat-value" style="font-size:18px">${escapeHtml(currentUser)}</div>
        <div class="stat-meta">${escapeHtml(state.session?.user?.role || authModeLabel())}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">已启用能力</div>
        <div class="stat-value">${enabledFeatureCount()}</div>
        <div class="stat-meta">shared admin contract</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Provider 管理</div>
        <div class="stat-value">${providerSummary}</div>
        <div class="stat-meta">${providerMeta}</div>
      </div>
      ${settings ? `
      <div class="stat-card">
        <div class="stat-label">API Key</div>
        <div class="stat-value" style="font-size:18px">${settings.apiKey ? "已启用" : "关闭"}</div>
        <div class="stat-meta">${settings.apiKey ? "Bearer 校验中" : "允许无鉴权"}</div>
      </div>` : `
      <div class="stat-card">
        <div class="stat-label">系统设置</div>
        <div class="stat-value" style="font-size:18px">未暴露</div>
        <div class="stat-meta">当前后端仅支持基础后台协议</div>
      </div>`}
      ${visibleSummaryDescriptors.map((item) => {
        const provider = providerStatusFor(status, item.key);
        return `
        <div class="stat-card">
          <div class="stat-label">${escapeHtml(item.label)}</div>
          <div class="stat-value">${provider.count}</div>
          <div class="stat-meta">
            <span class="status-dot ${provider.configured ? "healthy" : "unknown"}"></span>
            ${provider.configured ? `已配置${provider.active ? ` · ${escapeHtml(provider.active)}` : ""}` : "未配置"}
          </div>
        </div>`;
      }).join("")}`;

    if (!supportsProviderOverview()) {
      providerGrid.innerHTML = '<div class="empty-state">当前后端未启用 Provider 管理能力，桌面端已自动跳过 Go 专属管理接口。</div>';
      return;
    }

    providerGrid.innerHTML = descriptors.map((item) => {
      const provider = providerStatusFor(status, item.key);
      const available = isTabEnabled(item.tab);
      return `
      <button class="provider-card provider-card-button ${available ? "" : "disabled"}" data-provider-tab="${escapeHtml(item.tab)}" type="button" ${available ? "" : "disabled"}>
        <div class="provider-card-header">
          <div class="provider-card-info">
            <div class="provider-icon">${escapeHtml(item.initial)}</div>
            <div>
              <div class="provider-name">${escapeHtml(item.label)}</div>
              <div class="provider-meta">${escapeHtml(item.group)} · ${escapeHtml(item.mode)}</div>
            </div>
          </div>
          <span class="status-dot ${provider.configured ? "healthy" : item.mode === "预留" ? "degraded" : "unknown"}"></span>
        </div>
        <div class="provider-card-count">${provider.count}${item.mode === "列表" ? " 条" : " 项"}</div>
        <div class="provider-card-tags">
          <span class="tag ${provider.configured ? "success" : item.mode === "预留" ? "warning" : ""}">${provider.configured ? "已配置" : item.mode === "预留" ? "预留" : "未配置"}</span>
          ${provider.active ? `<span class="tag active">${escapeHtml(provider.active)}</span>` : ""}
        </div>
      </button>`;
    }).join("");

    providerGrid.querySelectorAll<HTMLButtonElement>("[data-provider-tab]").forEach((el) => {
      el.onclick = () => switchTab(el.dataset.providerTab || "providers");
    });
  };

  // ── Kiro / Grok List Rendering ──

  const renderKiroList = () => {
    if (!state.kiroAccounts.length) {
      kiroList.innerHTML = '<div class="empty-state">还没有 Kiro 账号，先新增一条或从 JSON 导入。</div>';
      return;
    }
    kiroList.innerHTML = state.kiroAccounts.map((item, i) => `
      <div class="item-card" id="kiro-item-${i}">
        <div class="item-card-header">
          <div>
            <div class="item-card-title">${escapeHtml(item.name || `Kiro 账号 ${i + 1}`)}</div>
            <div class="item-card-subtitle">Machine ID: ${escapeHtml(item.machineId || "未设置")}</div>
            <div class="item-card-tags">
              <span class="tag ${item.active ? "active" : ""}">${item.active ? "当前激活" : "未启用"}</span>
              <span class="tag">${escapeHtml(item.preferredEndpoint || "auto")}</span>
            </div>
          </div>
          <div class="item-card-actions">
            <button class="btn btn-sm btn-ghost" data-kind="kiro-toggle" data-index="${i}" type="button">详情</button>
            <button class="btn btn-sm btn-primary" data-kind="kiro-save" data-index="${i}" type="button">保存</button>
            <button class="btn btn-sm" data-kind="kiro-active" data-index="${i}" type="button">${item.active ? "已激活" : "设为激活"}</button>
            <button class="btn btn-sm btn-danger" data-kind="kiro-remove" data-index="${i}" type="button">删除</button>
          </div>
        </div>
        <div class="item-card-body collapsed">
          <div class="form-grid">
            <div><label>名称</label><input data-kind="kiro" data-index="${i}" data-field="name" value="${escapeHtml(item.name || "")}" /></div>
            <div><label>Machine ID</label><input data-kind="kiro" data-index="${i}" data-field="machineId" value="${escapeHtml(item.machineId || "")}" /></div>
            <div><label>Endpoint</label><select data-kind="kiro" data-index="${i}" data-field="preferredEndpoint"><option value="">auto</option><option value="amazonq" ${item.preferredEndpoint === "amazonq" ? "selected" : ""}>amazonq</option><option value="codewhisperer" ${item.preferredEndpoint === "codewhisperer" ? "selected" : ""}>codewhisperer</option></select></div>
            <div class="full-width"><label>Access Token</label><textarea data-kind="kiro" data-index="${i}" data-field="accessToken">${escapeHtml(item.accessToken || "")}</textarea></div>
          </div>
        </div>
      </div>`).join("");
  };

  const renderGrokList = () => {
    if (!state.grokTokens.length) {
      grokList.innerHTML = '<div class="empty-state">还没有 Grok Token，先新增一条。</div>';
      return;
    }
    grokList.innerHTML = state.grokTokens.map((item, i) => `
      <div class="item-card" id="grok-item-${i}">
        <div class="item-card-header">
          <div>
            <div class="item-card-title">${escapeHtml(item.name || `Grok Token ${i + 1}`)}</div>
            <div class="item-card-subtitle">Token: ${escapeHtml(maskSecret(item.cookieToken))}</div>
            <div class="item-card-tags"><span class="tag ${item.active ? "active" : ""}">${item.active ? "当前激活" : "未启用"}</span></div>
          </div>
          <div class="item-card-actions">
            <button class="btn btn-sm btn-ghost" data-kind="grok-toggle" data-index="${i}" type="button">详情</button>
            <button class="btn btn-sm btn-primary" data-kind="grok-save" data-index="${i}" type="button">保存</button>
            <button class="btn btn-sm" data-kind="grok-active" data-index="${i}" type="button">${item.active ? "已激活" : "设为激活"}</button>
            <button class="btn btn-sm btn-danger" data-kind="grok-remove" data-index="${i}" type="button">删除</button>
          </div>
        </div>
        <div class="item-card-body collapsed">
          <div class="form-grid">
            <div><label>名称</label><input data-kind="grok" data-index="${i}" data-field="name" value="${escapeHtml(item.name || "")}" /></div>
            <div class="full-width"><label>Cookie Token</label><textarea data-kind="grok" data-index="${i}" data-field="cookieToken">${escapeHtml(item.cookieToken || "")}</textarea></div>
          </div>
        </div>
      </div>`).join("");
  };

  const wireLists = () => {
    document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>("[data-kind='kiro'], [data-kind='grok']").forEach((el) => {
      el.oninput = (e) => {
        const t = e.target as HTMLInputElement;
        const { kind, index, field } = t.dataset;
        const list = kind === "kiro" ? state.kiroAccounts : state.grokTokens;
        if (index && field) (list[Number(index)] as Record<string, unknown>)[field] = t.value;
      };
    });

    document.querySelectorAll<HTMLElement>("[data-kind='kiro-toggle']").forEach((el) => {
      el.onclick = () => {
        const card = document.getElementById(`kiro-item-${el.dataset.index}`);
        if (!card) return;
        const body = card.querySelector(".item-card-body");
        if (!body) return;
        body.classList.toggle("collapsed");
        el.textContent = body.classList.contains("collapsed") ? "详情" : "收起";
      };
    });

    document.querySelectorAll<HTMLElement>("[data-kind='grok-toggle']").forEach((el) => {
      el.onclick = () => {
        const card = document.getElementById(`grok-item-${el.dataset.index}`);
        if (!card) return;
        const body = card.querySelector(".item-card-body");
        if (!body) return;
        body.classList.toggle("collapsed");
        el.textContent = body.classList.contains("collapsed") ? "详情" : "收起";
      };
    });

    document.querySelectorAll<HTMLElement>("[data-kind='kiro-save']").forEach((el) => {
      el.onclick = async () => {
        try {
          await updateKiroAccount(Number(el.dataset.index));
          await loadAdmin(false);
          toast("Kiro 账号已保存", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });

    document.querySelectorAll<HTMLElement>("[data-kind='grok-save']").forEach((el) => {
      el.onclick = async () => {
        try {
          await updateGrokToken(Number(el.dataset.index));
          await loadAdmin(false);
          toast("Grok Token 已保存", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });

    document.querySelectorAll<HTMLElement>("[data-kind='kiro-active']").forEach((el) => {
      el.onclick = async () => {
        try {
          await updateKiroAccount(Number(el.dataset.index), { active: true });
          await loadAdmin(false);
          toast("Kiro 激活账号已切换", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });
    document.querySelectorAll<HTMLElement>("[data-kind='grok-active']").forEach((el) => {
      el.onclick = async () => {
        try {
          await updateGrokToken(Number(el.dataset.index), { active: true });
          await loadAdmin(false);
          toast("Grok 激活 Token 已切换", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });
    document.querySelectorAll<HTMLElement>("[data-kind='kiro-remove']").forEach((el) => {
      el.onclick = async () => {
        if (!window.confirm("确认删除这条 Kiro 账号吗？")) return;
        try {
          await deleteKiroAccount(Number(el.dataset.index));
          await loadAdmin(false);
          toast("Kiro 账号已删除", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });
    document.querySelectorAll<HTMLElement>("[data-kind='grok-remove']").forEach((el) => {
      el.onclick = async () => {
        if (!window.confirm("确认删除这条 Grok Token 吗？")) return;
        try {
          await deleteGrokToken(Number(el.dataset.index));
          await loadAdmin(false);
          toast("Grok Token 已删除", "success");
        } catch (e) {
          toast((e as Error).message, "error");
        }
      };
    });
  };

  const clearEntryModalFields = () => {
    kiroEntryNameInput.value = "";
    kiroEntryMachineIDInput.value = "";
    kiroEntryEndpointInput.value = "";
    kiroEntryActiveInput.checked = false;
    kiroEntryAccessTokenInput.value = "";
    grokEntryNameInput.value = "";
    grokEntryActiveInput.checked = false;
    grokEntryCookieTokenInput.value = "";
  };

  const buildKiroAccountFromEntryModal = (): KiroAccount => {
    const accessToken = trimValue(kiroEntryAccessTokenInput.value);
    if (!accessToken) throw new Error("请先填写 Kiro Access Token");
    return {
      name: trimValue(kiroEntryNameInput.value) || `Kiro 账号 ${state.kiroAccounts.length + 1}`,
      accessToken,
      machineId: trimValue(kiroEntryMachineIDInput.value),
      preferredEndpoint: trimValue(kiroEntryEndpointInput.value).toLowerCase(),
      active: kiroEntryActiveInput.checked,
    };
  };

  const buildGrokTokenFromEntryModal = (): GrokToken => {
    const cookieToken = trimValue(grokEntryCookieTokenInput.value);
    if (!cookieToken) throw new Error("请先填写 Grok Cookie Token");
    return {
      name: trimValue(grokEntryNameInput.value) || `Grok Token ${state.grokTokens.length + 1}`,
      cookieToken,
      active: grokEntryActiveInput.checked,
    };
  };

  const requireItemID = (id: string | undefined, label: string): string => {
    const value = trimValue(id);
    if (!value) throw new Error(`${label} 缺少 ID，请刷新列表后重试`);
    return encodeURIComponent(value);
  };

  const normalizeKiroAccountPayload = (item: KiroAccount): KiroAccount => {
    const accessToken = trimValue(item.accessToken);
    if (!accessToken) throw new Error("请先填写 Kiro Access Token");
    return {
      id: trimValue(item.id),
      name: trimValue(item.name),
      accessToken,
      machineId: trimValue(item.machineId),
      preferredEndpoint: trimValue(item.preferredEndpoint).toLowerCase(),
      active: !!item.active,
    };
  };

  const normalizeGrokTokenPayload = (item: GrokToken): GrokToken => {
    const cookieToken = trimValue(item.cookieToken);
    if (!cookieToken) throw new Error("请先填写 Grok Cookie Token");
    return {
      id: trimValue(item.id),
      name: trimValue(item.name),
      cookieToken,
      active: !!item.active,
    };
  };

  const createKiroAccount = async (item: KiroAccount) => {
    await api<{ account: KiroAccount }>("/admin/api/providers/kiro/accounts/create", {
      method: "POST",
      body: JSON.stringify(normalizeKiroAccountPayload(item)),
    });
  };

  const updateKiroAccount = async (index: number, overrides: Partial<KiroAccount> = {}) => {
    const current = state.kiroAccounts[index];
    if (!current) throw new Error("Kiro 账号不存在");
    const payload = normalizeKiroAccountPayload({ ...current, ...overrides });
    const accountID = requireItemID(payload.id, "Kiro 账号");
    await api<{ account: KiroAccount }>(`/admin/api/providers/kiro/accounts/update/${accountID}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  };

  const deleteKiroAccount = async (index: number) => {
    const current = state.kiroAccounts[index];
    if (!current) throw new Error("Kiro 账号不存在");
    const accountID = requireItemID(current.id, "Kiro 账号");
    await api<{ ok: boolean }>(`/admin/api/providers/kiro/accounts/delete/${accountID}`, {
      method: "DELETE",
    });
  };

  const createGrokToken = async (item: GrokToken) => {
    await api<{ token: GrokToken }>("/admin/api/providers/grok/tokens/create", {
      method: "POST",
      body: JSON.stringify(normalizeGrokTokenPayload(item)),
    });
  };

  const updateGrokToken = async (index: number, overrides: Partial<GrokToken> = {}) => {
    const current = state.grokTokens[index];
    if (!current) throw new Error("Grok Token 不存在");
    const payload = normalizeGrokTokenPayload({ ...current, ...overrides });
    const tokenID = requireItemID(payload.id, "Grok Token");
    await api<{ token: GrokToken }>(`/admin/api/providers/grok/tokens/update/${tokenID}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  };

  const deleteGrokToken = async (index: number) => {
    const current = state.grokTokens[index];
    if (!current) throw new Error("Grok Token 不存在");
    const tokenID = requireItemID(current.id, "Grok Token");
    await api<{ ok: boolean }>(`/admin/api/providers/grok/tokens/delete/${tokenID}`, {
      method: "DELETE",
    });
  };

  // ── Data Loading ──

  const loadAdmin = async (showReadyToast = true) => {
    const [status, settings, cursor, kiro, grokConfig, grok, orchids, webConfig, chatgptConfig, zaiImage, zaiTTS, zaiOCR] = await Promise.all([
      optionalApi<AdminStatus>(supportsProviderOverview(), "/admin/api/status"),
      optionalApi<AdminSettings>(supportsSettings(), "/admin/api/settings"),
      optionalApi<{ config: CursorConfig }>(supportsProviderCredentials(), "/admin/api/providers/cursor/config"),
      optionalApi<{ accounts: KiroAccount[] }>(supportsProviderCredentials(), "/admin/api/providers/kiro/accounts/list"),
      optionalApi<{ config: GrokConfig }>(supportsProviderCredentials(), "/admin/api/providers/grok/config"),
      optionalApi<{ tokens: GrokToken[] }>(supportsProviderCredentials(), "/admin/api/providers/grok/tokens/list"),
      optionalApi<{ config: OrchidsConfig }>(supportsProviderCredentials(), "/admin/api/providers/orchids/config"),
      optionalApi<{ config: WebConfig }>(supportsProviderCredentials(), "/admin/api/providers/web/config"),
      optionalApi<{ config: ChatGPTConfig }>(supportsProviderCredentials(), "/admin/api/providers/chatgpt/config"),
      optionalApi<{ config: ZaiImageConfig }>(supportsProviderCredentials(), "/admin/api/providers/zai/image/config"),
      optionalApi<{ config: ZaiTTSConfig }>(supportsProviderCredentials(), "/admin/api/providers/zai/tts/config"),
      optionalApi<{ config: ZaiOCRConfig }>(supportsProviderCredentials(), "/admin/api/providers/zai/ocr/config"),
    ]);

    state.status = status;
    state.settings = settings;
    state.cursorConfig = cursor?.config || {};
    state.kiroAccounts = kiro?.accounts || [];
    state.grokConfig = grokConfig?.config || {};
    state.grokTokens = grok?.tokens || [];
    state.orchidsConfig = orchids?.config || {};
    state.webConfig = webConfig?.config || {};
    state.chatgptConfig = chatgptConfig?.config || {};
    state.zaiImageConfig = zaiImage?.config || {};
    state.zaiTTSConfig = zaiTTS?.config || {};
    state.zaiOCRConfig = zaiOCR?.config || {};
    apiKeyInput.value = settings?.apiKey || "";
    defaultProviderSelect.value = settings?.defaultProvider || "cursor";
    ensureSingleActive(state.kiroAccounts);
    ensureSingleActive(state.grokTokens);
    renderCursorConfig();
    renderGrokConfig();
    renderKiroList();
    renderGrokList();
    renderOrchidsConfig();
    renderClaudeConfig();
    renderChatGPTConfig();
    renderZaiImageConfig();
    renderZaiTTSConfig();
    renderZaiOCRConfig();
    renderStatus(status, settings);
    wireLists();
    switchTab(state.currentTab || "overview");
    setAuthenticatedState(state.session);
    if (showReadyToast) {
      toast(supportsSettings() || supportsProviderOverview() ? "管理台已就绪" : "已连接基础后台协议", "success");
    }
  };

  const bootstrapAdmin = async () => {
    toast("正在连接后台...");
    backendInput.value = state.baseUrl;
    setStatus(`正在连接 ${state.baseUrl} ...`);
    resetAdminData();
    try {
      const meta = await api<AdminMeta>("/api/admin/meta");
      applyMeta(meta);
      if (!state.sessionToken) persistToken(localStorage.getItem(tokenStorageKey(state.baseUrl)) || "");
      if (state.sessionToken) {
        try {
          state.session = await api<AdminSession>("/api/admin/auth/session");
          await loadAdmin();
          return;
        } catch (e) {
          if ((e as Error).message !== "UNAUTHORIZED") throw e;
          persistToken("");
        }
      }
      setLoggedOutState();
      toast("");
    } catch (e) {
      persistToken("");
      resetAdminData();
      setLoggedOutState();
      toast((e as Error).message, "error");
      setStatus(`连接失败：${(e as Error).message}`);
    }
  };

  const connectToBackend = async (baseUrl: string) => {
    state.baseUrl = normalizeBaseUrl(baseUrl);
    localStorage.setItem(BACKEND_STORAGE_KEY, state.baseUrl);
    persistToken(localStorage.getItem(tokenStorageKey(state.baseUrl)) || "");
    await bootstrapAdmin();
  };

  const saveCursorConfig = async () => {
    await api<{ config: CursorConfig }>("/admin/api/providers/cursor/config", {
      method: "PUT", body: JSON.stringify({ config: readCursorConfig() }),
    });
    await loadAdmin(false);
    toast("Cursor 配置已保存", "success");
  };

  const saveGrokConfig = async () => {
    await api<{ config: GrokConfig }>("/admin/api/providers/grok/config", {
      method: "PUT", body: JSON.stringify({ config: readGrokConfig() }),
    });
    await loadAdmin(false);
    toast("Grok 配置已保存", "success");
  };

  const saveOrchidsConfig = async () => {
    await api<{ config: OrchidsConfig }>("/admin/api/providers/orchids/config", {
      method: "PUT", body: JSON.stringify({ config: readOrchidsConfig() }),
    });
    await loadAdmin(false);
    toast("Orchids 配置已保存", "success");
  };

  const saveClaudeConfig = async () => {
    await api<{ config: WebConfig }>("/admin/api/providers/web/config", {
      method: "PUT", body: JSON.stringify({ config: readClaudeConfig() }),
    });
    await loadAdmin(false);
    toast("Claude 配置已保存", "success");
  };

  const saveChatGPTConfig = async () => {
    await api<{ config: ChatGPTConfig }>("/admin/api/providers/chatgpt/config", {
      method: "PUT", body: JSON.stringify({ config: readChatGPTConfig() }),
    });
    await loadAdmin(false);
    toast("ChatGPT 配置已保存", "success");
  };

  const closeConfigModal = () => {
    state.configModalProvider = null;
    configModal.classList.add("hidden");
    configModalOverlay.classList.add("hidden");
    configModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
    configModalSections.forEach((section) => section.classList.add("hidden"));
  };

  const closeEntryModal = () => {
    state.entryModalProvider = null;
    entryModal.classList.add("hidden");
    entryModalOverlay.classList.add("hidden");
    entryModal.setAttribute("aria-hidden", "true");
    document.body.classList.remove("modal-open");
    entryModalSections.forEach((section) => section.classList.add("hidden"));
    entryModalSaveButton.textContent = "创建";
    clearEntryModalFields();
  };

  const openConfigModal = (provider: ModalProvider) => {
    closeEntryModal();
    const configMap: Record<ModalProvider, { title: string; description: string }> = {
      cursor: { title: "Cursor 配置", description: "维护 Cursor 的全局接入参数与指纹字段。" },
      grok: { title: "Grok 配置", description: "维护 Grok 的 API URL、代理与请求头字段。" },
      orchids: { title: "Orchids 配置", description: "维护 Orchids 的 Clerk、会话与项目字段。" },
      claude: { title: "Claude 配置", description: "UI 显示为 Claude，底层仍保存到 web provider 配置。" },
      chatgpt: { title: "ChatGPT 配置", description: "维护 ChatGPT 的目标地址与访问 token。" },
    };
    state.configModalProvider = provider;
    configModalTitle.textContent = configMap[provider].title;
    configModalDescription.textContent = configMap[provider].description;
    configModalSections.forEach((section) => {
      section.classList.toggle("hidden", section.dataset.configSection !== provider);
    });
    configModal.classList.remove("hidden");
    configModalOverlay.classList.remove("hidden");
    configModal.setAttribute("aria-hidden", "false");
    document.body.classList.add("modal-open");
  };

  const openEntryModal = (provider: EntryModalProvider) => {
    closeConfigModal();
    clearEntryModalFields();
    const entryMap: Record<EntryModalProvider, { title: string; description: string }> = {
      kiro: { title: "新增 Kiro 账号", description: "先填写账号信息，再调用 create 接口创建当前条目。" },
      grok: { title: "新增 Grok Token", description: "先填写 Token 信息，再调用 create 接口创建当前条目。" },
    };
    state.entryModalProvider = provider;
    entryModalTitle.textContent = entryMap[provider].title;
    entryModalDescription.textContent = entryMap[provider].description;
    entryModalSaveButton.textContent = provider === "kiro" ? "创建账号" : "创建 Token";
    if (provider === "kiro") kiroEntryActiveInput.checked = state.kiroAccounts.length === 0;
    if (provider === "grok") grokEntryActiveInput.checked = state.grokTokens.length === 0;
    entryModalSections.forEach((section) => {
      section.classList.toggle("hidden", section.dataset.entrySection !== provider);
    });
    entryModal.classList.remove("hidden");
    entryModalOverlay.classList.remove("hidden");
    entryModal.setAttribute("aria-hidden", "false");
    document.body.classList.add("modal-open");
  };

  const saveConfigModal = async () => {
    if (!state.configModalProvider) return;
    const saveMap: Record<ModalProvider, () => Promise<void>> = {
      cursor: saveCursorConfig,
      grok: saveGrokConfig,
      orchids: saveOrchidsConfig,
      claude: saveClaudeConfig,
      chatgpt: saveChatGPTConfig,
    };
    await saveMap[state.configModalProvider]();
    closeConfigModal();
  };

  const saveEntryModal = async () => {
    if (!state.entryModalProvider) return;
    if (state.entryModalProvider === "kiro") {
      await createKiroAccount(buildKiroAccountFromEntryModal());
      await loadAdmin(false);
      closeEntryModal();
      toast("Kiro 账号已新增", "success");
      return;
    }
    await createGrokToken(buildGrokTokenFromEntryModal());
    await loadAdmin(false);
    closeEntryModal();
    toast("Grok Token 已新增", "success");
  };

  const toggleImportCard = (card: HTMLElement, button: HTMLButtonElement) => {
    const opened = card.classList.toggle("hidden") === false;
    button.textContent = opened ? "收起导入" : "导入 JSON";
  };

  // ── Event Bindings ──

  // Theme toggle
  must<HTMLButtonElement>("#theme-toggle").onclick = () => toggleTheme();

  // Hamburger menu
  must<HTMLButtonElement>("#hamburger-btn").onclick = () => toggleMobileSidebar();
  must<HTMLElement>("#sidebar-overlay").onclick = () => toggleMobileSidebar(false);

  // Sidebar collapse
  must<HTMLButtonElement>("#sidebar-collapse").onclick = () => toggleSidebar();

  // Tab navigation
  document.querySelectorAll<HTMLElement>("[data-tab]").forEach((el) => {
    el.onclick = () => switchTab(el.dataset.tab || "overview");
  });

  // Backend connection
  must<HTMLButtonElement>("#connect-btn").onclick = async () => { await connectToBackend(backendInput.value); };
  must<HTMLButtonElement>("#reset-btn").onclick = async () => { await connectToBackend(DEFAULT_BACKEND_URL); toast("已恢复默认后台地址", "success"); };
  backendInput.addEventListener("keydown", async (e) => { if (e.key === "Enter") await connectToBackend(backendInput.value); });

  // Login
  must<HTMLButtonElement>("#login-btn").onclick = async () => {
    try {
      const res = await api<AdminLoginResponse>("/api/admin/auth/login", {
        method: "POST", body: JSON.stringify({ password: loginPasswordInput.value }),
      });
      if (!res.token) throw new Error("登录成功但未收到会话令牌");
      persistToken(res.token);
      state.session = await api<AdminSession>("/api/admin/auth/session");
      await loadAdmin();
    } catch (e) {
      toast((e as Error).message === "UNAUTHORIZED" ? "登录失败，请检查管理密码" : (e as Error).message, "error");
    }
  };
  loginPasswordInput.addEventListener("keydown", async (e) => {
    if (e.key === "Enter") must<HTMLButtonElement>("#login-btn").click();
  });

  // Logout
  must<HTMLButtonElement>("#logout-btn").onclick = async () => {
    try { await api<{ ok: boolean }>("/api/admin/auth/logout", { method: "POST" }); }
    catch (e) { if ((e as Error).message !== "UNAUTHORIZED") { toast((e as Error).message, "error"); return; } }
    persistToken(""); setLoggedOutState(); toast("已退出登录", "success");
  };

  // Settings save
  must<HTMLButtonElement>("#save-settings-btn").onclick = async () => {
    try {
      await api<AdminSettings>("/admin/api/settings", {
        method: "PUT",
        body: JSON.stringify({ apiKey: apiKeyInput.value, defaultProvider: defaultProviderSelect.value, adminPassword: adminPasswordInput.value }),
      });
      adminPasswordInput.value = "";
      await loadAdmin(false);
      toast("系统设置已保存", "success");
    } catch (e) { toast((e as Error).message, "error"); }
  };

  // Config modal entrypoints
  must<HTMLButtonElement>("#cursor-config-btn").onclick = () => openConfigModal("cursor");
  must<HTMLButtonElement>("#grok-config-btn").onclick = () => openConfigModal("grok");
  must<HTMLButtonElement>("#orchids-config-btn").onclick = () => openConfigModal("orchids");
  must<HTMLButtonElement>("#claude-config-btn").onclick = () => openConfigModal("claude");
  must<HTMLButtonElement>("#chatgpt-config-btn").onclick = () => openConfigModal("chatgpt");
  configModalOverlay.onclick = () => closeConfigModal();
  configModalCloseButton.onclick = () => closeConfigModal();
  configModalCancelButton.onclick = () => closeConfigModal();
  configModalSaveButton.onclick = async () => {
    try {
      await saveConfigModal();
    } catch (e) {
      toast((e as Error).message, "error");
    }
  };
  entryModalOverlay.onclick = () => closeEntryModal();
  entryModalCloseButton.onclick = () => closeEntryModal();
  entryModalCancelButton.onclick = () => closeEntryModal();
  entryModalSaveButton.onclick = async () => {
    try {
      await saveEntryModal();
    } catch (e) {
      toast((e as Error).message, "error");
    }
  };
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !configModal.classList.contains("hidden")) closeConfigModal();
    if (e.key === "Escape" && !entryModal.classList.contains("hidden")) closeEntryModal();
  });

  // Kiro
  must<HTMLButtonElement>("#kiro-add-btn").onclick = () => openEntryModal("kiro");
  must<HTMLButtonElement>("#kiro-import-toggle-btn").onclick = (e) => toggleImportCard(kiroImportCard, e.currentTarget as HTMLButtonElement);
  must<HTMLButtonElement>("#kiro-export-btn").onclick = () => { downloadJSON("kiro-accounts.json", { accounts: state.kiroAccounts }); toast("Kiro 账号已导出", "success"); };
  must<HTMLButtonElement>("#kiro-import-btn").onclick = async () => {
    try {
      const parsed = JSON.parse(kiroImport.value || "{}") as ImportedKiroAccount & { accounts?: ImportedKiroAccount[] };
      const accounts: ImportedKiroAccount[] = Array.isArray(parsed.accounts) ? parsed.accounts : [parsed];
      for (const account of accounts) {
        await createKiroAccount({
          name: account.email || account.name || account.id || "Imported Kiro",
          accessToken: account.accessToken || account.credentials?.accessToken || "",
          machineId: account.machineId || account.credentials?.machineId || account.machineID || "",
          preferredEndpoint: account.preferredEndpoint || "",
          active: !!account.active,
        });
      }
      await loadAdmin(false);
      kiroImport.value = "";
      kiroImportCard.classList.add("hidden");
      must<HTMLButtonElement>("#kiro-import-toggle-btn").textContent = "导入 JSON";
      toast(`Kiro JSON 已导入 ${accounts.length} 条`, "success");
    } catch (e) { toast(`Kiro JSON 解析失败: ${(e as Error).message}`, "error"); }
  };

  // Grok
  must<HTMLButtonElement>("#grok-add-btn").onclick = () => openEntryModal("grok");
  must<HTMLButtonElement>("#grok-import-toggle-btn").onclick = (e) => toggleImportCard(grokImportCard, e.currentTarget as HTMLButtonElement);
  must<HTMLButtonElement>("#grok-export-btn").onclick = () => { downloadJSON("grok-tokens.json", { tokens: state.grokTokens }); toast("Grok Token 已导出", "success"); };
  must<HTMLButtonElement>("#grok-import-btn").onclick = async () => {
    try {
      const parsed = JSON.parse(grokImport.value || "{}") as ImportedGrokToken & { tokens?: ImportedGrokToken[] };
      const tokens: ImportedGrokToken[] = Array.isArray(parsed.tokens) ? parsed.tokens : [parsed];
      for (const token of tokens) {
        await createGrokToken({
          name: token.name || token.id || "Imported Grok",
          cookieToken: token.cookieToken || token.token || token.value || "",
          active: !!token.active,
        });
      }
      await loadAdmin(false);
      grokImport.value = "";
      grokImportCard.classList.add("hidden");
      must<HTMLButtonElement>("#grok-import-toggle-btn").textContent = "导入 JSON";
      toast(`Grok JSON 已导入 ${tokens.length} 条`, "success");
    } catch (e) { toast(`Grok JSON 解析失败: ${(e as Error).message}`, "error"); }
  };

  // Z.ai saves
  must<HTMLButtonElement>("#zai-image-save-btn").onclick = async () => {
    try {
      await api<{ config: ZaiImageConfig }>("/admin/api/providers/zai/image/config", {
        method: "PUT", body: JSON.stringify({ config: readZaiImageConfig() }),
      });
      await loadAdmin(false);
      toast("Z.ai Image 配置已保存", "success");
    } catch (e) { toast((e as Error).message, "error"); }
  };

  must<HTMLButtonElement>("#zai-tts-save-btn").onclick = async () => {
    try {
      await api<{ config: ZaiTTSConfig }>("/admin/api/providers/zai/tts/config", {
        method: "PUT", body: JSON.stringify({ config: readZaiTTSConfig() }),
      });
      await loadAdmin(false);
      toast("Z.ai TTS 配置已保存", "success");
    } catch (e) { toast((e as Error).message, "error"); }
  };

  must<HTMLButtonElement>("#zai-ocr-save-btn").onclick = async () => {
    try {
      await api<{ config: ZaiOCRConfig }>("/admin/api/providers/zai/ocr/config", {
        method: "PUT", body: JSON.stringify({ config: readZaiOCRConfig() }),
      });
      await loadAdmin(false);
      toast("Z.ai OCR 配置已保存", "success");
    } catch (e) { toast((e as Error).message, "error"); }
  };

  // ── Bootstrap ──
  backendInput.value = state.baseUrl;
  setStatus(`准备连接 ${state.baseUrl}`);
  void bootstrapAdmin();
});
