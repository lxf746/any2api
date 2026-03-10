use std::collections::HashSet;
use std::env;
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::Mutex;
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Clone, Debug, Default)]
pub struct AdminSettings {
    pub admin_password: String,
    pub api_key: String,
    pub default_provider: String,
}

#[derive(Clone, Debug, Default)]
pub struct CursorRuntimeConfig {
    pub api_url: String,
    pub script_url: String,
    pub cookie: String,
    pub x_is_human: String,
    pub user_agent: String,
    pub referer: String,
    pub webgl_vendor: String,
    pub webgl_renderer: String,
}

#[derive(Clone, Debug, Default)]
pub struct KiroAccount {
    pub id: String,
    pub name: String,
    pub access_token: String,
    pub machine_id: String,
    pub preferred_endpoint: String,
    pub active: bool,
}

#[derive(Clone, Debug, Default)]
pub struct GrokToken {
    pub id: String,
    pub name: String,
    pub cookie_token: String,
    pub active: bool,
}

#[derive(Clone, Debug, Default)]
pub struct GrokRuntimeConfig {
    pub api_url: String,
    pub proxy_url: String,
    pub cf_cookies: String,
    pub cf_clearance: String,
    pub user_agent: String,
    pub origin: String,
    pub referer: String,
}

#[derive(Clone, Debug, Default)]
pub struct OrchidsRuntimeConfig {
    pub api_url: String,
    pub clerk_url: String,
    pub client_cookie: String,
    pub client_uat: String,
    pub session_id: String,
    pub project_id: String,
    pub user_id: String,
    pub email: String,
    pub agent_mode: String,
}

#[derive(Clone, Debug, Default)]
pub struct WebRuntimeConfig {
    pub base_url: String,
    pub type_name: String,
    pub api_key: String,
}

#[derive(Clone, Debug, Default)]
pub struct ChatGPTRuntimeConfig {
    pub base_url: String,
    pub token: String,
}

#[derive(Clone, Debug, Default)]
pub struct ZAIImageRuntimeConfig {
    pub session_token: String,
    pub api_url: String,
}

#[derive(Clone, Debug, Default)]
pub struct ZAITTSRuntimeConfig {
    pub token: String,
    pub user_id: String,
    pub api_url: String,
}

#[derive(Clone, Debug, Default)]
pub struct ZAIOCRRuntimeConfig {
    pub token: String,
    pub api_url: String,
}

#[derive(Clone, Debug, Default)]
pub struct ProviderStore {
    pub cursor_config: CursorRuntimeConfig,
    pub kiro_accounts: Vec<KiroAccount>,
    pub grok_config: GrokRuntimeConfig,
    pub grok_tokens: Vec<GrokToken>,
    pub orchids_config: OrchidsRuntimeConfig,
    pub web_config: WebRuntimeConfig,
    pub chatgpt_config: ChatGPTRuntimeConfig,
    pub zai_image_config: ZAIImageRuntimeConfig,
    pub zai_tts_config: ZAITTSRuntimeConfig,
    pub zai_ocr_config: ZAIOCRRuntimeConfig,
}

#[derive(Clone, Debug, Default)]
pub struct AdminConfig {
    pub settings: AdminSettings,
    pub providers: ProviderStore,
}

struct AdminStoreState {
    path: PathBuf,
    data: AdminConfig,
}

pub struct AdminStore {
    state: Mutex<AdminStoreState>,
}

pub(crate) fn json_escape(input: &str) -> String {
    input.replace('\\', "\\\\").replace('"', "\\\"")
}

pub(crate) fn json_string_field(body: &str, field: &str) -> Option<String> {
    let key = format!("\"{field}\"");
    let start = body.find(&key)?;
    let rest = &body[start + key.len()..];
    let colon = rest.find(':')?;
    let value = rest[colon + 1..].trim_start();
    let mut chars = value.chars();
    if chars.next()? != '"' {
        return None;
    }
    let mut result = String::new();
    let mut escaped = false;
    for ch in chars {
        if escaped {
            result.push(ch);
            escaped = false;
            continue;
        }
        match ch {
            '\\' => escaped = true,
            '"' => return Some(result),
            _ => result.push(ch),
        }
    }
    None
}

pub(crate) fn json_bool_field(body: &str, field: &str) -> Option<bool> {
    let key = format!("\"{field}\"");
    let start = body.find(&key)?;
    let rest = &body[start + key.len()..];
    let colon = rest.find(':')?;
    let value = rest[colon + 1..].trim_start();
    if value.starts_with("true") {
        return Some(true);
    }
    if value.starts_with("false") {
        return Some(false);
    }
    None
}

fn slice_balanced(value: &str, open: char, close: char) -> Option<String> {
    let mut depth = 0_u32;
    let mut in_string = false;
    let mut escaped = false;
    let mut started = false;
    let mut result = String::new();
    for ch in value.chars() {
        if !started {
            if ch == open {
                started = true;
                depth = 1;
                result.push(ch);
            } else if ch.is_whitespace() {
                continue;
            } else {
                return None;
            }
            continue;
        }
        result.push(ch);
        if in_string {
            if escaped {
                escaped = false;
            } else if ch == '\\' {
                escaped = true;
            } else if ch == '"' {
                in_string = false;
            }
            continue;
        }
        if ch == '"' {
            in_string = true;
            continue;
        }
        if ch == open {
            depth += 1;
        } else if ch == close {
            depth -= 1;
            if depth == 0 {
                return Some(result);
            }
        }
    }
    None
}

pub(crate) fn json_object_field(body: &str, field: &str) -> Option<String> {
    let key = format!("\"{field}\"");
    let start = body.find(&key)?;
    let rest = &body[start + key.len()..];
    let colon = rest.find(':')?;
    slice_balanced(&rest[colon + 1..], '{', '}')
}

pub(crate) fn json_array_field(body: &str, field: &str) -> Option<String> {
    let key = format!("\"{field}\"");
    let start = body.find(&key)?;
    let rest = &body[start + key.len()..];
    let colon = rest.find(':')?;
    slice_balanced(&rest[colon + 1..], '[', ']')
}

pub(crate) fn json_array_objects(body: &str) -> Vec<String> {
    let mut items = Vec::new();
    let trimmed = body.trim();
    if !trimmed.starts_with('[') || !trimmed.ends_with(']') {
        return items;
    }
    let inner = &trimmed[1..trimmed.len() - 1];
    let mut cursor = inner;
    while let Some(index) = cursor.find('{') {
        let candidate = &cursor[index..];
        let Some(item) = slice_balanced(candidate, '{', '}') else {
            break;
        };
        cursor = &candidate[item.len()..];
        items.push(item);
    }
    items
}

fn current_default_provider() -> String {
    env::var("NEWPLATFORM2API_DEFAULT_PROVIDER")
        .unwrap_or_else(|_| "cursor".to_string())
        .trim()
        .to_string()
}

fn current_admin_password() -> String {
    env::var("NEWPLATFORM2API_ADMIN_PASSWORD")
        .unwrap_or_else(|_| "changeme".to_string())
        .trim()
        .to_string()
}

fn current_api_key() -> String {
    env::var("NEWPLATFORM2API_API_KEY")
        .unwrap_or_default()
        .trim()
        .to_string()
}

fn current_store_path() -> PathBuf {
    env::var("NEWPLATFORM2API_ADMIN_STORE_PATH")
        .map(PathBuf::from)
        .unwrap_or_else(|_| Path::new(env!("CARGO_MANIFEST_DIR")).join("data/admin.json"))
}

fn default_grok_config() -> GrokRuntimeConfig {
    GrokRuntimeConfig {
        api_url: env::var("NEWPLATFORM2API_GROK_API_URL")
            .unwrap_or_else(|_| "https://grok.com/rest/app-chat/conversations/new".to_string())
            .trim()
            .to_string(),
        proxy_url: env::var("NEWPLATFORM2API_GROK_PROXY_URL")
            .unwrap_or_default()
            .trim()
            .to_string(),
        cf_cookies: env::var("NEWPLATFORM2API_GROK_CF_COOKIES")
            .unwrap_or_default()
            .trim()
            .to_string(),
        cf_clearance: env::var("NEWPLATFORM2API_GROK_CF_CLEARANCE")
            .unwrap_or_default()
            .trim()
            .to_string(),
        user_agent: env::var("NEWPLATFORM2API_GROK_USER_AGENT")
            .unwrap_or_else(|_| "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36".to_string())
            .trim()
            .to_string(),
        origin: env::var("NEWPLATFORM2API_GROK_ORIGIN")
            .unwrap_or_else(|_| "https://grok.com".to_string())
            .trim()
            .to_string(),
        referer: env::var("NEWPLATFORM2API_GROK_REFERER")
            .unwrap_or_else(|_| "https://grok.com/".to_string())
            .trim()
            .to_string(),
    }
}

fn default_web_config() -> WebRuntimeConfig {
    normalize_web_config(WebRuntimeConfig {
        base_url: env::var("NEWPLATFORM2API_WEB_BASE_URL")
            .unwrap_or_else(|_| "http://127.0.0.1:9000".to_string())
            .trim()
            .to_string(),
        type_name: env::var("NEWPLATFORM2API_WEB_TYPE")
            .unwrap_or_else(|_| "claude".to_string())
            .trim()
            .to_string(),
        api_key: env::var("NEWPLATFORM2API_WEB_API_KEY")
            .unwrap_or_default()
            .trim()
            .to_string(),
    })
}

fn default_chatgpt_config() -> ChatGPTRuntimeConfig {
    normalize_chatgpt_config(ChatGPTRuntimeConfig {
        base_url: env::var("NEWPLATFORM2API_CHATGPT_BASE_URL")
            .unwrap_or_else(|_| "http://127.0.0.1:5005".to_string())
            .trim()
            .to_string(),
        token: env::var("NEWPLATFORM2API_CHATGPT_TOKEN")
            .unwrap_or_default()
            .trim()
            .to_string(),
    })
}

fn default_zai_image_config() -> ZAIImageRuntimeConfig {
    ZAIImageRuntimeConfig {
        session_token: env::var("NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN")
            .or_else(|_| env::var("ZAI_IMAGE_SESSION_TOKEN"))
            .unwrap_or_default()
            .trim()
            .to_string(),
        api_url: env::var("NEWPLATFORM2API_ZAI_IMAGE_API_URL")
            .unwrap_or_else(|_| "https://image.z.ai/api/proxy/images/generate".to_string())
            .trim()
            .to_string(),
    }
}

fn default_zai_tts_config() -> ZAITTSRuntimeConfig {
    ZAITTSRuntimeConfig {
        token: env::var("NEWPLATFORM2API_ZAI_TTS_TOKEN")
            .or_else(|_| env::var("ZAI_TTS_TOKEN"))
            .unwrap_or_default()
            .trim()
            .to_string(),
        user_id: env::var("NEWPLATFORM2API_ZAI_TTS_USER_ID")
            .or_else(|_| env::var("ZAI_TTS_USER_ID"))
            .unwrap_or_default()
            .trim()
            .to_string(),
        api_url: env::var("NEWPLATFORM2API_ZAI_TTS_API_URL")
            .unwrap_or_else(|_| "https://audio.z.ai/api/v1/z-audio/tts/create".to_string())
            .trim()
            .to_string(),
    }
}

fn default_zai_ocr_config() -> ZAIOCRRuntimeConfig {
    ZAIOCRRuntimeConfig {
        token: env::var("NEWPLATFORM2API_ZAI_OCR_TOKEN")
            .or_else(|_| env::var("ZAI_OCR_TOKEN"))
            .unwrap_or_default()
            .trim()
            .to_string(),
        api_url: env::var("NEWPLATFORM2API_ZAI_OCR_API_URL")
            .unwrap_or_else(|_| "https://ocr.z.ai/api/v1/z-ocr/tasks/process".to_string())
            .trim()
            .to_string(),
    }
}

fn default_admin_config() -> AdminConfig {
    AdminConfig {
        settings: AdminSettings {
            admin_password: current_admin_password(),
            api_key: current_api_key(),
            default_provider: current_default_provider(),
        },
        providers: ProviderStore {
            grok_config: default_grok_config(),
            web_config: default_web_config(),
            chatgpt_config: default_chatgpt_config(),
            zai_image_config: default_zai_image_config(),
            zai_tts_config: default_zai_tts_config(),
            zai_ocr_config: default_zai_ocr_config(),
            ..ProviderStore::default()
        },
    }
}

impl CursorRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
            script_url: json_string_field(input, "scriptUrl").unwrap_or_default(),
            cookie: json_string_field(input, "cookie").unwrap_or_default(),
            x_is_human: json_string_field(input, "xIsHuman").unwrap_or_default(),
            user_agent: json_string_field(input, "userAgent").unwrap_or_default(),
            referer: json_string_field(input, "referer").unwrap_or_default(),
            webgl_vendor: json_string_field(input, "webglVendor").unwrap_or_default(),
            webgl_renderer: json_string_field(input, "webglRenderer").unwrap_or_default(),
        }
    }
}

impl KiroAccount {
    fn from_json(input: &str) -> Self {
        Self {
            id: json_string_field(input, "id").unwrap_or_default(),
            name: json_string_field(input, "name").unwrap_or_default(),
            access_token: json_string_field(input, "accessToken").unwrap_or_default(),
            machine_id: json_string_field(input, "machineId").unwrap_or_default(),
            preferred_endpoint: json_string_field(input, "preferredEndpoint").unwrap_or_default().to_lowercase(),
            active: json_bool_field(input, "active").unwrap_or(false),
        }
    }
}

impl GrokToken {
    fn from_json(input: &str) -> Self {
        Self {
            id: json_string_field(input, "id").unwrap_or_default(),
            name: json_string_field(input, "name").unwrap_or_default(),
            cookie_token: json_string_field(input, "cookieToken").unwrap_or_default(),
            active: json_bool_field(input, "active").unwrap_or(false),
        }
    }
}

impl GrokRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
            proxy_url: json_string_field(input, "proxyUrl").unwrap_or_default(),
            cf_cookies: json_string_field(input, "cfCookies").unwrap_or_default(),
            cf_clearance: json_string_field(input, "cfClearance").unwrap_or_default(),
            user_agent: json_string_field(input, "userAgent").unwrap_or_default(),
            origin: json_string_field(input, "origin").unwrap_or_default(),
            referer: json_string_field(input, "referer").unwrap_or_default(),
        }
    }
}

impl OrchidsRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
            clerk_url: json_string_field(input, "clerkUrl").unwrap_or_default(),
            client_cookie: json_string_field(input, "clientCookie").unwrap_or_default(),
            client_uat: json_string_field(input, "clientUat").unwrap_or_default(),
            session_id: json_string_field(input, "sessionId").unwrap_or_default(),
            project_id: json_string_field(input, "projectId").unwrap_or_default(),
            user_id: json_string_field(input, "userId").unwrap_or_default(),
            email: json_string_field(input, "email").unwrap_or_default(),
            agent_mode: json_string_field(input, "agentMode").unwrap_or_default(),
        }
    }
}

impl WebRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            base_url: json_string_field(input, "baseUrl").unwrap_or_default(),
            type_name: json_string_field(input, "type").unwrap_or_default(),
            api_key: json_string_field(input, "apiKey").unwrap_or_default(),
        }
    }
}

impl ChatGPTRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            base_url: json_string_field(input, "baseUrl").unwrap_or_default(),
            token: json_string_field(input, "token").unwrap_or_default(),
        }
    }
}

impl ZAIImageRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            session_token: json_string_field(input, "sessionToken").unwrap_or_default(),
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
        }
    }
}

impl ZAITTSRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            token: json_string_field(input, "token").unwrap_or_default(),
            user_id: json_string_field(input, "userId").unwrap_or_default(),
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
        }
    }
}

impl ZAIOCRRuntimeConfig {
    pub fn from_json(input: &str) -> Self {
        Self {
            token: json_string_field(input, "token").unwrap_or_default(),
            api_url: json_string_field(input, "apiUrl").unwrap_or_default(),
        }
    }
}

fn next_item_id(prefix: &str, existing_ids: &HashSet<String>) -> String {
    let base = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|duration| duration.as_millis())
        .unwrap_or(0);
    let mut seq = 0_u64;
    loop {
        let candidate = if seq == 0 {
            format!("{prefix}-{base}")
        } else {
            format!("{prefix}-{base}-{seq}")
        };
        if !existing_ids.contains(&candidate) {
            return candidate;
        }
        seq += 1;
    }
}

fn prepare_kiro_account(
    mut account: KiroAccount,
    existing_ids: &HashSet<String>,
    fallback_name_index: usize,
) -> Option<KiroAccount> {
    account.access_token = account.access_token.trim().to_string();
    account.machine_id = account.machine_id.trim().to_string();
    account.preferred_endpoint = account.preferred_endpoint.trim().to_lowercase();
    if account.access_token.is_empty() && account.machine_id.is_empty() {
        return None;
    }
    let account_id = account.id.trim().to_string();
    if account_id.is_empty() || existing_ids.contains(&account_id) {
        account.id = next_item_id("kiro", existing_ids);
    } else {
        account.id = account_id;
    }
    if account.name.trim().is_empty() {
        account.name = format!("Kiro Account {}", fallback_name_index);
    }
    Some(account)
}

fn prepare_grok_token(
    mut token: GrokToken,
    existing_ids: &HashSet<String>,
    fallback_name_index: usize,
) -> Option<GrokToken> {
    token.cookie_token = token.cookie_token.trim().to_string();
    if token.cookie_token.is_empty() {
        return None;
    }
    let token_id = token.id.trim().to_string();
    if token_id.is_empty() || existing_ids.contains(&token_id) {
        token.id = next_item_id("grok", existing_ids);
    } else {
        token.id = token_id;
    }
    if token.name.trim().is_empty() {
        token.name = format!("Grok Token {}", fallback_name_index);
    }
    Some(token)
}

fn find_kiro_account_index(accounts: &[KiroAccount], account_id: &str) -> Option<usize> {
    let target = account_id.trim();
    if target.is_empty() {
        return None;
    }
    accounts.iter().position(|item| item.id.trim() == target)
}

fn find_grok_token_index(tokens: &[GrokToken], token_id: &str) -> Option<usize> {
    let target = token_id.trim();
    if target.is_empty() {
        return None;
    }
    tokens.iter().position(|item| item.id.trim() == target)
}

fn normalize_kiro_accounts(accounts: Vec<KiroAccount>) -> Vec<KiroAccount> {
    let mut normalized = Vec::new();
    let mut existing_ids = HashSet::new();
    let mut active_set = false;
    for account in accounts {
        let Some(mut account) = prepare_kiro_account(account, &existing_ids, normalized.len() + 1) else {
            continue;
        };
        existing_ids.insert(account.id.clone());
        if active_set {
            account.active = false;
        } else if account.active {
            active_set = true;
        }
        normalized.push(account);
    }
    if !active_set && !normalized.is_empty() {
        normalized[0].active = true;
    }
    normalized
}

fn normalize_grok_tokens(tokens: Vec<GrokToken>) -> Vec<GrokToken> {
    let mut normalized = Vec::new();
    let mut existing_ids = HashSet::new();
    let mut active_set = false;
    for token in tokens {
        let Some(mut token) = prepare_grok_token(token, &existing_ids, normalized.len() + 1) else {
            continue;
        };
        existing_ids.insert(token.id.clone());
        if active_set {
            token.active = false;
        } else if token.active {
            active_set = true;
        }
        normalized.push(token);
    }
    if !active_set && !normalized.is_empty() {
        normalized[0].active = true;
    }
    normalized
}

fn normalize_grok_config(mut config: GrokRuntimeConfig) -> GrokRuntimeConfig {
    config.api_url = config.api_url.trim().to_string();
    config.proxy_url = config.proxy_url.trim().to_string();
    config.cf_cookies = config.cf_cookies.trim().to_string();
    config.cf_clearance = config.cf_clearance.trim().to_string();
    config.user_agent = config.user_agent.trim().to_string();
    config.origin = config.origin.trim().to_string();
    config.referer = config.referer.trim().to_string();
    config
}

fn normalize_web_config(mut config: WebRuntimeConfig) -> WebRuntimeConfig {
    config.base_url = config.base_url.trim().trim_end_matches('/').to_string();
    config.type_name = config.type_name.trim().trim_matches('/').to_string();
    config.api_key = config.api_key.trim().to_string();
    config
}

fn normalize_chatgpt_config(mut config: ChatGPTRuntimeConfig) -> ChatGPTRuntimeConfig {
    config.base_url = config.base_url.trim().trim_end_matches('/').to_string();
    config.token = config.token.trim().to_string();
    config
}

fn normalize_zai_image_config(mut config: ZAIImageRuntimeConfig) -> ZAIImageRuntimeConfig {
    let defaults = default_zai_image_config();
    config.session_token = if config.session_token.trim().is_empty() {
        defaults.session_token
    } else {
        config.session_token.trim().to_string()
    };
    config.api_url = if config.api_url.trim().is_empty() {
        defaults.api_url
    } else {
        config.api_url.trim().to_string()
    };
    config
}

fn normalize_zai_tts_config(mut config: ZAITTSRuntimeConfig) -> ZAITTSRuntimeConfig {
    let defaults = default_zai_tts_config();
    config.token = if config.token.trim().is_empty() {
        defaults.token
    } else {
        config.token.trim().to_string()
    };
    config.user_id = if config.user_id.trim().is_empty() {
        defaults.user_id
    } else {
        config.user_id.trim().to_string()
    };
    config.api_url = if config.api_url.trim().is_empty() {
        defaults.api_url
    } else {
        config.api_url.trim().to_string()
    };
    config
}

fn normalize_zai_ocr_config(mut config: ZAIOCRRuntimeConfig) -> ZAIOCRRuntimeConfig {
    let defaults = default_zai_ocr_config();
    config.token = if config.token.trim().is_empty() {
        defaults.token
    } else {
        config.token.trim().to_string()
    };
    config.api_url = if config.api_url.trim().is_empty() {
        defaults.api_url
    } else {
        config.api_url.trim().to_string()
    };
    config
}

fn parse_admin_config(input: &str) -> AdminConfig {
    let defaults = default_admin_config();
    let settings_json = json_object_field(input, "settings").unwrap_or_default();
    let providers_json = json_object_field(input, "providers").unwrap_or_default();
    let cursor_json = json_object_field(&providers_json, "cursorConfig").unwrap_or_default();
    let grok_config_json = json_object_field(&providers_json, "grokConfig").unwrap_or_default();
    let orchids_json = json_object_field(&providers_json, "orchidsConfig").unwrap_or_default();
    let web_json = json_object_field(&providers_json, "webConfig").unwrap_or_default();
    let chatgpt_json = json_object_field(&providers_json, "chatgptConfig").unwrap_or_default();
    let zai_image_json = json_object_field(&providers_json, "zaiImageConfig").unwrap_or_default();
    let zai_tts_json = json_object_field(&providers_json, "zaiTTSConfig").unwrap_or_default();
    let zai_ocr_json = json_object_field(&providers_json, "zaiOCRConfig").unwrap_or_default();
    let kiro_json = json_array_field(&providers_json, "kiroAccounts").unwrap_or_default();
    let grok_json = json_array_field(&providers_json, "grokTokens").unwrap_or_default();
    AdminConfig {
        settings: AdminSettings {
            admin_password: json_string_field(&settings_json, "adminPassword")
                .unwrap_or(defaults.settings.admin_password),
            api_key: json_string_field(&settings_json, "apiKey")
                .unwrap_or(defaults.settings.api_key),
            default_provider: json_string_field(&settings_json, "defaultProvider")
                .unwrap_or(defaults.settings.default_provider),
        },
        providers: ProviderStore {
            cursor_config: CursorRuntimeConfig::from_json(&cursor_json),
            kiro_accounts: normalize_kiro_accounts(
                json_array_objects(&kiro_json)
                    .into_iter()
                    .map(|item| KiroAccount::from_json(&item))
                    .collect(),
            ),
            grok_config: if grok_config_json.is_empty() {
                defaults.providers.grok_config
            } else {
                normalize_grok_config(GrokRuntimeConfig::from_json(&grok_config_json))
            },
            grok_tokens: normalize_grok_tokens(
                json_array_objects(&grok_json)
                    .into_iter()
                    .map(|item| GrokToken::from_json(&item))
                    .collect(),
            ),
            orchids_config: OrchidsRuntimeConfig::from_json(&orchids_json),
            web_config: if web_json.is_empty() {
                defaults.providers.web_config
            } else {
                normalize_web_config(WebRuntimeConfig::from_json(&web_json))
            },
            chatgpt_config: if chatgpt_json.is_empty() {
                defaults.providers.chatgpt_config.clone()
            } else {
                normalize_chatgpt_config(ChatGPTRuntimeConfig::from_json(&chatgpt_json))
            },
            zai_image_config: if zai_image_json.is_empty() {
                defaults.providers.zai_image_config.clone()
            } else {
                normalize_zai_image_config(ZAIImageRuntimeConfig::from_json(&zai_image_json))
            },
            zai_tts_config: if zai_tts_json.is_empty() {
                defaults.providers.zai_tts_config.clone()
            } else {
                normalize_zai_tts_config(ZAITTSRuntimeConfig::from_json(&zai_tts_json))
            },
            zai_ocr_config: if zai_ocr_json.is_empty() {
                defaults.providers.zai_ocr_config.clone()
            } else {
                normalize_zai_ocr_config(ZAIOCRRuntimeConfig::from_json(&zai_ocr_json))
            },
        },
    }
}

fn render_admin_config(data: &AdminConfig) -> String {
    let kiro = data.providers.kiro_accounts.iter().map(|item| format!(
        "{{\"id\":\"{}\",\"name\":\"{}\",\"accessToken\":\"{}\",\"machineId\":\"{}\",\"preferredEndpoint\":\"{}\",\"active\":{}}}",
        json_escape(&item.id), json_escape(&item.name), json_escape(&item.access_token), json_escape(&item.machine_id), json_escape(&item.preferred_endpoint), item.active
    )).collect::<Vec<_>>().join(",");
    let grok = data.providers.grok_tokens.iter().map(|item| format!(
        "{{\"id\":\"{}\",\"name\":\"{}\",\"cookieToken\":\"{}\",\"active\":{}}}",
        json_escape(&item.id), json_escape(&item.name), json_escape(&item.cookie_token), item.active
    )).collect::<Vec<_>>().join(",");
    let mut out = String::new();
    out.push_str("{\n");
    out.push_str("  \"settings\": {\n");
    out.push_str(&format!(
        "    \"adminPassword\": \"{}\",\n",
        json_escape(&data.settings.admin_password)
    ));
    out.push_str(&format!(
        "    \"apiKey\": \"{}\",\n",
        json_escape(&data.settings.api_key)
    ));
    out.push_str(&format!(
        "    \"defaultProvider\": \"{}\"\n",
        json_escape(&data.settings.default_provider)
    ));
    out.push_str("  },\n");
    out.push_str("  \"providers\": {\n");
    out.push_str(&format!(
        "    \"cursorConfig\": {{\"apiUrl\":\"{}\",\"scriptUrl\":\"{}\",\"cookie\":\"{}\",\"xIsHuman\":\"{}\",\"userAgent\":\"{}\",\"referer\":\"{}\",\"webglVendor\":\"{}\",\"webglRenderer\":\"{}\"}},\n",
        json_escape(&data.providers.cursor_config.api_url),
        json_escape(&data.providers.cursor_config.script_url),
        json_escape(&data.providers.cursor_config.cookie),
        json_escape(&data.providers.cursor_config.x_is_human),
        json_escape(&data.providers.cursor_config.user_agent),
        json_escape(&data.providers.cursor_config.referer),
        json_escape(&data.providers.cursor_config.webgl_vendor),
        json_escape(&data.providers.cursor_config.webgl_renderer),
    ));
    out.push_str(&format!("    \"kiroAccounts\": [{}],\n", kiro));
    out.push_str(&format!(
        "    \"grokConfig\": {{\"apiUrl\":\"{}\",\"proxyUrl\":\"{}\",\"cfCookies\":\"{}\",\"cfClearance\":\"{}\",\"userAgent\":\"{}\",\"origin\":\"{}\",\"referer\":\"{}\"}},\n",
        json_escape(&data.providers.grok_config.api_url),
        json_escape(&data.providers.grok_config.proxy_url),
        json_escape(&data.providers.grok_config.cf_cookies),
        json_escape(&data.providers.grok_config.cf_clearance),
        json_escape(&data.providers.grok_config.user_agent),
        json_escape(&data.providers.grok_config.origin),
        json_escape(&data.providers.grok_config.referer),
    ));
    out.push_str(&format!("    \"grokTokens\": [{}],\n", grok));
    out.push_str(&format!(
        "    \"orchidsConfig\": {{\"apiUrl\":\"{}\",\"clerkUrl\":\"{}\",\"clientCookie\":\"{}\",\"clientUat\":\"{}\",\"sessionId\":\"{}\",\"projectId\":\"{}\",\"userId\":\"{}\",\"email\":\"{}\",\"agentMode\":\"{}\"}}\n",
        json_escape(&data.providers.orchids_config.api_url),
        json_escape(&data.providers.orchids_config.clerk_url),
        json_escape(&data.providers.orchids_config.client_cookie),
        json_escape(&data.providers.orchids_config.client_uat),
        json_escape(&data.providers.orchids_config.session_id),
        json_escape(&data.providers.orchids_config.project_id),
        json_escape(&data.providers.orchids_config.user_id),
        json_escape(&data.providers.orchids_config.email),
        json_escape(&data.providers.orchids_config.agent_mode),
    ));
    out.pop();
    out.push_str(&format!(
        ",\n    \"webConfig\": {{\"baseUrl\":\"{}\",\"type\":\"{}\",\"apiKey\":\"{}\"}},\n",
        json_escape(&data.providers.web_config.base_url),
        json_escape(&data.providers.web_config.type_name),
        json_escape(&data.providers.web_config.api_key),
    ));
    out.push_str(&format!(
        "    \"chatgptConfig\": {{\"baseUrl\":\"{}\",\"token\":\"{}\"}}\n",
        json_escape(&data.providers.chatgpt_config.base_url),
        json_escape(&data.providers.chatgpt_config.token),
    ));
    out.pop();
    out.push_str(&format!(
        ",\n    \"zaiImageConfig\": {{\"sessionToken\":\"{}\",\"apiUrl\":\"{}\"}},\n",
        json_escape(&data.providers.zai_image_config.session_token),
        json_escape(&data.providers.zai_image_config.api_url),
    ));
    out.push_str(&format!(
        "    \"zaiTTSConfig\": {{\"token\":\"{}\",\"userId\":\"{}\",\"apiUrl\":\"{}\"}},\n",
        json_escape(&data.providers.zai_tts_config.token),
        json_escape(&data.providers.zai_tts_config.user_id),
        json_escape(&data.providers.zai_tts_config.api_url),
    ));
    out.push_str(&format!(
        "    \"zaiOCRConfig\": {{\"token\":\"{}\",\"apiUrl\":\"{}\"}}\n",
        json_escape(&data.providers.zai_ocr_config.token),
        json_escape(&data.providers.zai_ocr_config.api_url),
    ));
    out.push_str("  }\n");
    out.push_str("}\n");
    out
}

fn load_or_init(path: &Path) -> AdminConfig {
    match fs::read_to_string(path) {
        Ok(content) if !content.trim().is_empty() => parse_admin_config(&content),
        _ => default_admin_config(),
    }
}

impl AdminStore {
    pub fn from_env() -> Self {
        let path = current_store_path();
        let data = load_or_init(&path);
        let store = Self {
            state: Mutex::new(AdminStoreState { path, data }),
        };
        let _ = store.persist();
        store
    }

    pub fn sync_from_env(&self) {
        let path = current_store_path();
        let data = load_or_init(&path);
        if let Ok(mut state) = self.state.lock() {
            state.path = path;
            state.data = data;
        }
        let _ = self.persist();
    }

    pub fn snapshot(&self) -> AdminConfig {
        self.state.lock().expect("admin store lock poisoned").data.clone()
    }

    pub fn admin_password(&self) -> String {
        self.snapshot().settings.admin_password
    }

    pub fn update_settings(&self, api_key: &str, default_provider: &str, admin_password: &str) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.settings.api_key = api_key.trim().to_string();
        state.data.settings.default_provider = if default_provider.trim().is_empty() { current_default_provider() } else { default_provider.trim().to_string() };
        if !admin_password.trim().is_empty() {
            state.data.settings.admin_password = admin_password.trim().to_string();
        }
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_cursor_config(&self, config: CursorRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.cursor_config = config;
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn kiro_account(&self, account_id: &str) -> Option<KiroAccount> {
        let state = self.state.lock().expect("admin store lock poisoned");
        let index = find_kiro_account_index(&state.data.providers.kiro_accounts, account_id)?;
        Some(state.data.providers.kiro_accounts[index].clone())
    }

    pub fn create_kiro_account(&self, account: KiroAccount) -> Result<KiroAccount, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let existing_ids = state
            .data
            .providers
            .kiro_accounts
            .iter()
            .map(|item| item.id.clone())
            .collect::<HashSet<_>>();
        let prepared = prepare_kiro_account(
            account,
            &existing_ids,
            state.data.providers.kiro_accounts.len() + 1,
        )
        .ok_or_else(|| "invalid kiro account".to_string())?;
        let mut next = state.data.providers.kiro_accounts.clone();
        if prepared.active {
            for item in &mut next {
                item.active = false;
            }
        }
        next.push(prepared.clone());
        next = normalize_kiro_accounts(next);
        let index = find_kiro_account_index(&next, &prepared.id)
            .ok_or_else(|| "invalid kiro account".to_string())?;
        let created = next[index].clone();
        state.data.providers.kiro_accounts = next;
        persist_state(&state.path, &state.data)?;
        Ok(created)
    }

    pub fn update_kiro_account(
        &self,
        account_id: &str,
        account: KiroAccount,
    ) -> Result<Option<KiroAccount>, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let Some(index) = find_kiro_account_index(&state.data.providers.kiro_accounts, account_id) else {
            return Ok(None);
        };
        let mut candidate = account;
        candidate.id = account_id.trim().to_string();
        let existing_ids = state
            .data
            .providers
            .kiro_accounts
            .iter()
            .enumerate()
            .filter(|(item_index, _)| *item_index != index)
            .map(|(_, item)| item.id.clone())
            .collect::<HashSet<_>>();
        let prepared = prepare_kiro_account(candidate, &existing_ids, index + 1)
            .ok_or_else(|| "invalid kiro account".to_string())?;
        let mut next = state.data.providers.kiro_accounts.clone();
        if prepared.active {
            for item in &mut next {
                item.active = false;
            }
        }
        next[index] = prepared;
        next = normalize_kiro_accounts(next);
        let updated_index = find_kiro_account_index(&next, account_id)
            .ok_or_else(|| "invalid kiro account".to_string())?;
        let updated = next[updated_index].clone();
        state.data.providers.kiro_accounts = next;
        persist_state(&state.path, &state.data)?;
        Ok(Some(updated))
    }

    pub fn delete_kiro_account(&self, account_id: &str) -> Result<bool, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let Some(index) = find_kiro_account_index(&state.data.providers.kiro_accounts, account_id) else {
            return Ok(false);
        };
        let mut next = state.data.providers.kiro_accounts.clone();
        next.remove(index);
        state.data.providers.kiro_accounts = normalize_kiro_accounts(next);
        persist_state(&state.path, &state.data)?;
        Ok(true)
    }

    pub fn replace_kiro_accounts(&self, accounts: Vec<KiroAccount>) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.kiro_accounts = normalize_kiro_accounts(accounts);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn grok_token(&self, token_id: &str) -> Option<GrokToken> {
        let state = self.state.lock().expect("admin store lock poisoned");
        let index = find_grok_token_index(&state.data.providers.grok_tokens, token_id)?;
        Some(state.data.providers.grok_tokens[index].clone())
    }

    pub fn create_grok_token(&self, token: GrokToken) -> Result<GrokToken, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let existing_ids = state
            .data
            .providers
            .grok_tokens
            .iter()
            .map(|item| item.id.clone())
            .collect::<HashSet<_>>();
        let prepared = prepare_grok_token(token, &existing_ids, state.data.providers.grok_tokens.len() + 1)
            .ok_or_else(|| "invalid grok token".to_string())?;
        let mut next = state.data.providers.grok_tokens.clone();
        if prepared.active {
            for item in &mut next {
                item.active = false;
            }
        }
        next.push(prepared.clone());
        next = normalize_grok_tokens(next);
        let index = find_grok_token_index(&next, &prepared.id)
            .ok_or_else(|| "invalid grok token".to_string())?;
        let created = next[index].clone();
        state.data.providers.grok_tokens = next;
        persist_state(&state.path, &state.data)?;
        Ok(created)
    }

    pub fn update_grok_token(
        &self,
        token_id: &str,
        token: GrokToken,
    ) -> Result<Option<GrokToken>, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let Some(index) = find_grok_token_index(&state.data.providers.grok_tokens, token_id) else {
            return Ok(None);
        };
        let mut candidate = token;
        candidate.id = token_id.trim().to_string();
        let existing_ids = state
            .data
            .providers
            .grok_tokens
            .iter()
            .enumerate()
            .filter(|(item_index, _)| *item_index != index)
            .map(|(_, item)| item.id.clone())
            .collect::<HashSet<_>>();
        let prepared = prepare_grok_token(candidate, &existing_ids, index + 1)
            .ok_or_else(|| "invalid grok token".to_string())?;
        let mut next = state.data.providers.grok_tokens.clone();
        if prepared.active {
            for item in &mut next {
                item.active = false;
            }
        }
        next[index] = prepared;
        next = normalize_grok_tokens(next);
        let updated_index = find_grok_token_index(&next, token_id)
            .ok_or_else(|| "invalid grok token".to_string())?;
        let updated = next[updated_index].clone();
        state.data.providers.grok_tokens = next;
        persist_state(&state.path, &state.data)?;
        Ok(Some(updated))
    }

    pub fn delete_grok_token(&self, token_id: &str) -> Result<bool, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        let Some(index) = find_grok_token_index(&state.data.providers.grok_tokens, token_id) else {
            return Ok(false);
        };
        let mut next = state.data.providers.grok_tokens.clone();
        next.remove(index);
        state.data.providers.grok_tokens = normalize_grok_tokens(next);
        persist_state(&state.path, &state.data)?;
        Ok(true)
    }

    pub fn replace_grok_tokens(&self, tokens: Vec<GrokToken>) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.grok_tokens = normalize_grok_tokens(tokens);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_grok_config(&self, config: GrokRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.grok_config = normalize_grok_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_orchids_config(&self, config: OrchidsRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.orchids_config = config;
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_web_config(&self, config: WebRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.web_config = normalize_web_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_chatgpt_config(&self, config: ChatGPTRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.chatgpt_config = normalize_chatgpt_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_zai_image_config(&self, config: ZAIImageRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.zai_image_config = normalize_zai_image_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_zai_tts_config(&self, config: ZAITTSRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.zai_tts_config = normalize_zai_tts_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    pub fn replace_zai_ocr_config(&self, config: ZAIOCRRuntimeConfig) -> Result<AdminConfig, String> {
        let mut state = self.state.lock().expect("admin store lock poisoned");
        state.data.providers.zai_ocr_config = normalize_zai_ocr_config(config);
        persist_state(&state.path, &state.data)?;
        Ok(state.data.clone())
    }

    fn persist(&self) -> Result<(), String> {
        let state = self.state.lock().expect("admin store lock poisoned");
        persist_state(&state.path, &state.data)
    }
}

fn persist_state(path: &Path, data: &AdminConfig) -> Result<(), String> {
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).map_err(|err| format!("create admin config dir: {err}"))?;
    }
    fs::write(path, render_admin_config(data)).map_err(|err| format!("write admin config: {err}"))
}