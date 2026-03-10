use std::collections::BTreeMap;
use std::env;
use std::sync::Mutex;
use std::time::Duration;

use reqwest::blocking::{Client, Response};
use reqwest::{header::HeaderMap, Proxy};
use serde_json::{json, Value};

use crate::admin_store::{GrokRuntimeConfig, GrokToken};
use crate::providers::common::{
    env_int, normalize_incremental_chunk, normalize_messages, pick_active_grok_token,
    random_hex, trim_cookie_value, DEFAULT_MAX_INPUT_LENGTH,
};
use crate::registry::Provider;
use crate::types::{ModelInfo, ProviderCapabilities, UnifiedRequest};

const DEFAULT_GROK_API_URL: &str = "https://grok.com/rest/app-chat/conversations/new";
const DEFAULT_GROK_USER_AGENT: &str = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36";
const DEFAULT_GROK_ORIGIN: &str = "https://grok.com";
const DEFAULT_GROK_REFERER: &str = "https://grok.com/";
const DEFAULT_TIMEOUT_SECONDS: u64 = 60;
const GROK_MAX_RETRIES: usize = 3;
const GROK_RETRY_BUDGET_SECONDS: u64 = 12;
const GROK_RETRY_BACKOFF_BASE_MILLIS: u64 = 400;
const GROK_RETRY_BACKOFF_MAX_MILLIS: u64 = 4_000;

struct GrokStreamFilter {
    tool_card_open: bool,
    buffer: String,
}

impl GrokStreamFilter {
    fn new() -> Self {
        Self {
            tool_card_open: false,
            buffer: String::new(),
        }
    }

    fn filter(&mut self, token: &str) -> String {
        if token.is_empty() {
            return String::new();
        }
        let start_tag = "<xai:tool_usage_card";
        let end_tag = "</xai:tool_usage_card>";
        let mut output = String::new();
        let mut remaining = token;
        while !remaining.is_empty() {
            if self.tool_card_open {
                let Some(end_index) = remaining.find(end_tag) else {
                    self.buffer.push_str(remaining);
                    return output;
                };
                let end_pos = end_index + end_tag.len();
                self.buffer.push_str(&remaining[..end_pos]);
                let summary = summarize_grok_tool_card(&self.buffer);
                if !summary.is_empty() {
                    output.push_str(&summary);
                    if !summary.ends_with('\n') {
                        output.push('\n');
                    }
                }
                self.buffer.clear();
                self.tool_card_open = false;
                remaining = &remaining[end_pos..];
                continue;
            }
            let Some(start_index) = remaining.find(start_tag) else {
                output.push_str(remaining);
                break;
            };
            if start_index > 0 {
                output.push_str(&remaining[..start_index]);
            }
            let Some(end_index) = remaining[start_index..].find(end_tag) else {
                self.tool_card_open = true;
                self.buffer.push_str(&remaining[start_index..]);
                break;
            };
            let end_pos = start_index + end_index + end_tag.len();
            let summary = summarize_grok_tool_card(&remaining[start_index..end_pos]);
            if !summary.is_empty() {
                output.push_str(&summary);
                if !summary.ends_with('\n') {
                    output.push('\n');
                }
            }
            remaining = &remaining[end_pos..];
        }
        output
    }
}

pub struct GrokProvider {
    tokens: Vec<GrokToken>,
    config: GrokRuntimeConfig,
    client: Mutex<Client>,
    client_error: Option<String>,
    sleep: fn(Duration),
}

impl GrokProvider {
    pub fn new(tokens: Vec<GrokToken>, config: GrokRuntimeConfig) -> Self {
        let config = normalize_grok_runtime_config(config);
        let (client, client_error) = match Self::build_client(&config) {
            Ok(client) => (client, None),
            Err(err) => (Self::fallback_client(), Some(err)),
        };
        Self {
            tokens,
            config,
            client: Mutex::new(client),
            client_error,
            sleep: std::thread::sleep,
        }
    }

    fn fallback_client() -> Client {
        Client::builder()
            .timeout(Duration::from_secs(env_int(
                "NEWPLATFORM2API_GROK_TIMEOUT",
                DEFAULT_TIMEOUT_SECONDS,
            )))
            .build()
            .expect("build fallback grok reqwest client")
    }

    fn build_client(config: &GrokRuntimeConfig) -> Result<Client, String> {
        let mut builder = Client::builder().timeout(Duration::from_secs(env_int(
            "NEWPLATFORM2API_GROK_TIMEOUT",
            DEFAULT_TIMEOUT_SECONDS,
        )));
        let proxy_url = config.proxy_url.trim();
        if !proxy_url.is_empty() {
            let scheme = proxy_url
                .split(':')
                .next()
                .unwrap_or_default()
                .trim()
                .to_ascii_lowercase();
            if scheme != "http" && scheme != "https" {
                return Err(format!(
                    "grok proxy scheme is not supported by rust backend: {proxy_url}"
                ));
            }
            let proxy = Proxy::all(proxy_url)
                .map_err(|err| format!("invalid grok proxy url: {err}"))?;
            builder = builder.proxy(proxy);
        }
        builder
            .build()
            .map_err(|err| format!("build grok reqwest client: {err}"))
    }

    fn token(&self) -> Result<GrokToken, String> {
        pick_active_grok_token(&self.tokens)
            .or_else(Self::env_token)
            .ok_or_else(|| "grok cookie token is not configured".to_string())
    }

    fn env_token() -> Option<GrokToken> {
        let cookie_token = env::var("NEWPLATFORM2API_GROK_COOKIE_TOKEN").ok()?;
        let cookie_token = cookie_token.trim().to_string();
        if cookie_token.is_empty() {
            return None;
        }
        Some(GrokToken {
            id: "grok-env".to_string(),
            name: "Env Grok Token".to_string(),
            cookie_token,
            active: true,
        })
    }

    fn api_url(&self) -> String {
        self.config.api_url.clone()
    }

    fn headers(&self, cookie_token: &str) -> Vec<(&'static str, String)> {
        vec![
            ("Accept", "*/*".to_string()),
            ("Accept-Encoding", "gzip, deflate, br".to_string()),
            ("Accept-Language", "en-US,en;q=0.9".to_string()),
            ("Content-Type", "application/json".to_string()),
            (
                "Cookie",
                build_grok_cookie_header(
                    cookie_token,
                    &self.config.cf_cookies,
                    &self.config.cf_clearance,
                ),
            ),
            ("Origin", self.config.origin.clone()),
            ("Priority", "u=1, i".to_string()),
            ("Referer", self.config.referer.clone()),
            ("Sec-Fetch-Dest", "empty".to_string()),
            ("Sec-Fetch-Mode", "cors".to_string()),
            (
                "Sec-Fetch-Site",
                grok_sec_fetch_site(&self.config.origin, &self.config.referer),
            ),
            ("User-Agent", self.config.user_agent.clone()),
            ("X-Statsig-Id", random_hex(16)),
            ("X-XAI-Request-Id", random_hex(32)),
            ("X-Requested-With", "XMLHttpRequest".to_string()),
        ]
    }

    fn current_client(&self) -> Result<Client, String> {
        self.client
            .lock()
            .map_err(|_| "grok client lock poisoned".to_string())
            .map(|client| client.clone())
    }

    fn reset_client(&self) -> Result<(), String> {
        if let Some(err) = &self.client_error {
            return Err(err.clone());
        }
        let client = Self::build_client(&self.config)?;
        let mut guard = self
            .client
            .lock()
            .map_err(|_| "grok client lock poisoned".to_string())?;
        *guard = client;
        Ok(())
    }

    fn send_request(&self, body: &str, cookie_token: &str) -> Result<Response, String> {
        if let Some(err) = &self.client_error {
            return Err(err.clone());
        }
        let client = self.current_client()?;
        let mut request = client
            .post(self.api_url())
            .header("Content-Type", "application/json")
            .body(body.to_string());
        for (key, value) in self.headers(cookie_token) {
            request = request.header(key, value);
        }
        request
            .send()
            .map_err(|err| format!("grok upstream request failed: {err}"))
    }

    fn build_payload(&self, req: &UnifiedRequest) -> Value {
        let model = map_grok_model(&req.model);
        json!({
            "deviceEnvInfo": {
                "darkModeEnabled": false,
                "devicePixelRatio": 2,
                "screenWidth": 2056,
                "screenHeight": 1329,
                "viewportWidth": 2056,
                "viewportHeight": 1083,
            },
            "disableMemory": false,
            "disableSearch": false,
            "disableSelfHarmShortCircuit": false,
            "disableTextFollowUps": false,
            "enableImageGeneration": true,
            "enableImageStreaming": true,
            "enableSideBySide": true,
            "fileAttachments": [],
            "forceConcise": false,
            "forceSideBySide": false,
            "imageAttachments": [],
            "imageGenerationCount": 2,
            "isAsyncChat": false,
            "isReasoning": false,
            "message": self.flatten_messages(req),
            "modelName": model,
            "responseMetadata": {
                "requestModelDetails": {"modelId": map_grok_model(&req.model)}
            },
            "returnImageBytes": false,
            "returnRawGrokInXaiRequest": false,
            "sendFinalMetadata": true,
            "temporary": false,
            "toolOverrides": {},
        })
    }

    fn flatten_messages(&self, req: &UnifiedRequest) -> String {
        let normalized = normalize_messages(req, DEFAULT_MAX_INPUT_LENGTH);
        let mut parts = Vec::new();
        for message in normalized {
            let text = super::super::common::content_text(&message.content);
            if text.is_empty() {
                continue;
            }
            let role = message.role.trim().to_lowercase();
            parts.push((if role.is_empty() { "user".to_string() } else { role }, text));
        }
        if parts.is_empty() {
            return ".".to_string();
        }
        let mut last_user_index = None;
        for (index, (role, _)) in parts.iter().enumerate().rev() {
            if role == "user" {
                last_user_index = Some(index);
                break;
            }
        }
        let last_user_index = last_user_index.unwrap_or(parts.len().saturating_sub(1));
        parts
            .iter()
            .enumerate()
            .map(|(index, (role, text))| {
                if index == last_user_index {
                    text.clone()
                } else {
                    format!("{role}: {text}")
                }
            })
            .collect::<Vec<_>>()
            .join("\n\n")
    }

    fn collect_text(&self, raw: &str) -> String {
        let mut filter = GrokStreamFilter::new();
        let mut last_message = String::new();
        let mut token_seen = false;
        let mut parts = Vec::new();
        for line in raw.lines() {
            let line = line.trim();
            if line.is_empty() {
                continue;
            }
            let Ok(payload) = serde_json::from_str::<Value>(line) else {
                continue;
            };
            let Some(response) = nested_value(&payload, &["result", "response"]) else {
                continue;
            };
            if let Some(token) = response.get("token").and_then(Value::as_str) {
                if !token.is_empty() {
                    token_seen = true;
                    let filtered = strip_grok_artifacts(&filter.filter(token));
                    if !filtered.is_empty() {
                        parts.push(filtered);
                    }
                    continue;
                }
            }
            if token_seen {
                continue;
            }
            let Some(message) = response
                .get("modelResponse")
                .and_then(Value::as_object)
                .and_then(|item| item.get("message"))
                .and_then(Value::as_str)
            else {
                continue;
            };
            if message.is_empty() {
                continue;
            }
            let filtered = strip_grok_artifacts(message);
            let delta = normalize_incremental_chunk(&filtered, &last_message);
            if !delta.is_empty() {
                last_message = filtered;
                parts.push(delta);
            }
        }
        parts.join("").trim().to_string()
    }
}

impl Provider for GrokProvider {
    fn id(&self) -> &'static str {
        "grok"
    }

    fn capabilities(&self) -> ProviderCapabilities {
        ProviderCapabilities {
            openai_compatible: true,
            anthropic_compatible: false,
            tools: true,
            images: true,
            multi_account: true,
        }
    }

    fn models(&self) -> Vec<ModelInfo> {
        vec![ModelInfo {
            provider: "grok",
            public_model: "grok-4",
            upstream_model: "grok-4",
            owned_by: "xai",
        }]
    }

    fn build_upstream_preview(&self, req: &UnifiedRequest) -> String {
        format!(
            "url={} auth=sso-cookie protocol={} model={} proxy_configured={} cf_configured={}",
            self.api_url(),
            req.protocol,
            map_grok_model(&req.model),
            !self.config.proxy_url.trim().is_empty(),
            !self.config.cf_cookies.trim().is_empty() || !self.config.cf_clearance.trim().is_empty(),
        )
    }

    fn generate_reply(&self, req: &UnifiedRequest) -> Result<String, String> {
        let token = self.token()?;
        let cookie_token = token.cookie_token.trim();
        if cookie_token.is_empty() {
            return Err("grok cookie token is not configured".to_string());
        }
        let payload = self.build_payload(req).to_string();
        let mut total_delay = Duration::from_secs(0);
        let mut last_delay = retry_backoff_base();
        for attempt in 0..=GROK_MAX_RETRIES {
            match self.send_request(&payload, cookie_token) {
                Ok(response) => {
                    let status = response.status();
                    if status.is_success() {
                        let body = response
                            .text()
                            .map_err(|err| format!("read grok upstream body: {err}"))?;
                        return Ok(self.collect_text(&body));
                    }
                    let headers = response.headers().clone();
                    let body = response.text().unwrap_or_default();
                    let err = format!(
                        "grok upstream error: status={} body={}",
                        status.as_u16(),
                        body.trim()
                    );
                    if attempt == GROK_MAX_RETRIES || !should_retry_grok_status(status.as_u16()) {
                        return Err(err);
                    }
                    if should_reset_grok_session(status.as_u16()) {
                        self.reset_client()?;
                    }
                    let delay = grok_retry_delay(
                        last_delay,
                        parse_retry_after_seconds(&headers),
                        status.as_u16(),
                        attempt,
                    );
                    if total_delay + delay > retry_budget() {
                        return Err(err);
                    }
                    total_delay += delay;
                    last_delay = if delay > retry_backoff_base() {
                        delay
                    } else {
                        retry_backoff_base()
                    };
                    if delay > Duration::from_secs(0) {
                        (self.sleep)(delay);
                    }
                }
                Err(err) => {
                    if attempt == GROK_MAX_RETRIES || !should_retry_grok_transport_error(&err) {
                        return Err(err);
                    }
                    self.reset_client()?;
                    let delay = grok_retry_delay(last_delay, Duration::from_secs(0), 0, attempt);
                    if total_delay + delay > retry_budget() {
                        return Err(err);
                    }
                    total_delay += delay;
                    last_delay = if delay > retry_backoff_base() {
                        delay
                    } else {
                        retry_backoff_base()
                    };
                    if delay > Duration::from_secs(0) {
                        (self.sleep)(delay);
                    }
                }
            }
        }
        Err("grok upstream request failed after retries".to_string())
    }
}

fn map_grok_model(model: &str) -> String {
    let trimmed = model.trim();
    if trimmed.is_empty() {
        "grok-4".to_string()
    } else {
        trimmed.to_string()
    }
}

fn retry_backoff_base() -> Duration {
    Duration::from_millis(GROK_RETRY_BACKOFF_BASE_MILLIS)
}

fn retry_backoff_max() -> Duration {
    Duration::from_millis(GROK_RETRY_BACKOFF_MAX_MILLIS)
}

fn retry_budget() -> Duration {
    Duration::from_secs(GROK_RETRY_BUDGET_SECONDS)
}

fn duration_from_millis(value: u128) -> Duration {
    Duration::from_millis(value.min(u64::MAX as u128) as u64)
}

fn normalize_grok_runtime_config(mut config: GrokRuntimeConfig) -> GrokRuntimeConfig {
    config.api_url = config.api_url.trim().to_string();
    if config.api_url.is_empty() {
        config.api_url = DEFAULT_GROK_API_URL.to_string();
    }
    config.proxy_url = config.proxy_url.trim().to_string();
    config.cf_cookies = config.cf_cookies.trim().to_string();
    config.cf_clearance = config.cf_clearance.trim().to_string();
    config.user_agent = config.user_agent.trim().to_string();
    if config.user_agent.is_empty() {
        config.user_agent = DEFAULT_GROK_USER_AGENT.to_string();
    }
    config.origin = config.origin.trim().to_string();
    if config.origin.is_empty() {
        config.origin = DEFAULT_GROK_ORIGIN.to_string();
    }
    config.referer = config.referer.trim().to_string();
    if config.referer.is_empty() {
        config.referer = DEFAULT_GROK_REFERER.to_string();
    }
    config
}

fn build_grok_cookie_header(token: &str, cf_cookies: &str, cf_clearance: &str) -> String {
    let trimmed = token.trim();
    if trimmed.is_empty() {
        return String::new();
    }
    let base = if trimmed.contains(';') {
        trimmed.to_string()
    } else {
        let trimmed = trim_cookie_value(trimmed, "sso=");
        format!("sso={trimmed}; sso-rw={trimmed}")
    };
    let mut cookies = parse_cookie_map(&base);
    for (name, value) in parse_cookie_map(cf_cookies) {
        cookies.insert(name, value);
    }
    let clearance = cf_clearance.trim();
    if !clearance.is_empty() {
        cookies.insert("cf_clearance".to_string(), clearance.to_string());
    }
    cookies
        .into_iter()
        .map(|(name, value)| format!("{name}={value}"))
        .collect::<Vec<_>>()
        .join("; ")
}

fn parse_cookie_map(raw: &str) -> BTreeMap<String, String> {
    let mut cookies = BTreeMap::new();
    for chunk in raw.split(';') {
        let (name, value) = chunk.trim().split_once('=').unwrap_or(("", ""));
        if name.trim().is_empty() {
            continue;
        }
        cookies.insert(name.trim().to_string(), value.trim().to_string());
    }
    cookies
}

fn parse_retry_after_seconds(headers: &HeaderMap) -> Duration {
    let Some(value) = headers.get("Retry-After") else {
        return Duration::from_secs(0);
    };
    let Ok(raw) = value.to_str() else {
        return Duration::from_secs(0);
    };
    let Ok(seconds) = raw.trim().parse::<u64>() else {
        return Duration::from_secs(0);
    };
    Duration::from_secs(seconds)
}

fn grok_retry_delay(
    last_delay: Duration,
    retry_after: Duration,
    status_code: u16,
    attempt: usize,
) -> Duration {
    if retry_after > Duration::from_secs(0) {
        return retry_after;
    }
    if status_code == 429 {
        let next = duration_from_millis(
            (last_delay.as_millis().max(retry_backoff_base().as_millis()) * 2)
                .min(retry_backoff_max().as_millis()),
        );
        return next;
    }
    duration_from_millis(
        (retry_backoff_base().as_millis() * (1_u128 << attempt.min(4)))
            .min(retry_backoff_max().as_millis()),
    )
}

fn should_retry_grok_status(status_code: u16) -> bool {
    matches!(status_code, 403 | 408 | 429 | 502 | 503 | 504) || status_code >= 500
}

fn should_reset_grok_session(status_code: u16) -> bool {
    status_code == 403
}

fn should_retry_grok_transport_error(err: &str) -> bool {
    let message = err.trim().to_ascii_lowercase();
    if message.is_empty() {
        return true;
    }
    !message.contains("unsupported protocol scheme")
        && !message.contains("proxy scheme is not supported")
        && !message.contains("invalid grok proxy url")
}

fn grok_sec_fetch_site(origin: &str, referer: &str) -> String {
    let origin_host = origin
        .split("//")
        .nth(1)
        .and_then(|value| value.split('/').next())
        .unwrap_or_default()
        .to_ascii_lowercase();
    let referer_host = referer
        .split("//")
        .nth(1)
        .and_then(|value| value.split('/').next())
        .unwrap_or_default()
        .to_ascii_lowercase();
    if origin_host.is_empty() || referer_host.is_empty() {
        return "same-origin".to_string();
    }
    if origin_host == referer_host {
        "same-origin".to_string()
    } else {
        "cross-site".to_string()
    }
}

fn nested_value<'a>(root: &'a Value, keys: &[&str]) -> Option<&'a Value> {
    let mut current = root;
    for key in keys {
        current = current.get(*key)?;
    }
    Some(current)
}

fn summarize_grok_tool_card(raw: &str) -> String {
    let name = extract_tag_text(raw, "<xai:tool_name>", "</xai:tool_name>");
    let args = extract_tag_text(raw, "<xai:tool_args>", "</xai:tool_args>");
    if name.is_empty() && args.is_empty() {
        return String::new();
    }
    if args.is_empty() {
        return format!("[{name}]");
    }
    format!("[{name}] {args}")
}

fn extract_tag_text(raw: &str, start_tag: &str, end_tag: &str) -> String {
    let Some(start) = raw.find(start_tag) else {
        return String::new();
    };
    let content_start = start + start_tag.len();
    let Some(relative_end) = raw[content_start..].find(end_tag) else {
        return String::new();
    };
    let text = &raw[content_start..content_start + relative_end];
    strip_cdata(text).trim().to_string()
}

fn strip_cdata(text: &str) -> String {
    let trimmed = text.trim();
    if let Some(rest) = trimmed.strip_prefix("<![CDATA[") {
        if let Some(end) = rest.strip_suffix("]]>") {
            return end.to_string();
        }
    }
    trimmed.to_string()
}

fn strip_wrapped_sections(text: &str, start_tag: &str, end_tag: &str) -> String {
    let mut remaining = text;
    let mut out = String::new();
    while let Some(start) = remaining.find(start_tag) {
        out.push_str(&remaining[..start]);
        let tail = &remaining[start + start_tag.len()..];
        if let Some(end) = tail.find(end_tag) {
            remaining = &tail[end + end_tag.len()..];
        } else {
            remaining = "";
            break;
        }
    }
    out.push_str(remaining);
    out
}

fn remove_xai_tags(text: &str) -> String {
    let mut out = String::new();
    let mut remaining = text;
    while let Some(start) = remaining.find('<') {
        out.push_str(&remaining[..start]);
        let tail = &remaining[start..];
        let Some(end) = tail.find('>') else {
            out.push_str(tail);
            return out;
        };
        let tag = &tail[1..end].trim();
        if !tag.starts_with("xai:") && !tag.starts_with("/xai:") {
            out.push_str(&tail[..=end]);
        }
        remaining = &tail[end + 1..];
    }
    out.push_str(remaining);
    out
}

fn replace_tool_cards(text: &str) -> String {
    let start_tag = "<xai:tool_usage_card";
    let end_tag = "</xai:tool_usage_card>";
    let mut remaining = text;
    let mut out = String::new();
    while let Some(start) = remaining.find(start_tag) {
        out.push_str(&remaining[..start]);
        let tail = &remaining[start..];
        let Some(end) = tail.find(end_tag) else {
            break;
        };
        let end_pos = end + end_tag.len();
        out.push_str(&summarize_grok_tool_card(&tail[..end_pos]));
        remaining = &tail[end_pos..];
    }
    out.push_str(remaining);
    out
}

fn strip_grok_artifacts(text: &str) -> String {
    if text.is_empty() {
        return String::new();
    }
    let cleaned = replace_tool_cards(text);
    let cleaned = strip_wrapped_sections(&cleaned, "<rolloutId>", "</rolloutId>");
    remove_xai_tags(&cleaned)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;
    use std::io::{Read, Write};
    use std::net::TcpListener;
    use std::sync::atomic::{AtomicUsize, Ordering};
    use std::sync::Arc;
    use std::thread;

    use crate::types::Message;

    #[derive(Debug)]
    struct TestRequest {
        method: String,
        path: String,
        headers: HashMap<String, String>,
        body: String,
    }

    impl TestRequest {
        fn header(&self, name: &str) -> Option<&str> {
            self.headers
                .get(&name.to_ascii_lowercase())
                .map(String::as_str)
        }
    }

    struct MockResponse {
        status: &'static str,
        content_type: &'static str,
        headers: Vec<(String, String)>,
        body: String,
    }

    fn parse_request(raw: &str) -> TestRequest {
        let (head, body) = raw.split_once("\r\n\r\n").unwrap_or((raw, ""));
        let mut lines = head.lines();
        let first = lines.next().unwrap_or_default();
        let mut parts = first.split_whitespace();
        let method = parts.next().unwrap_or_default().to_string();
        let path = parts.next().unwrap_or_default().to_string();
        let mut headers = HashMap::new();
        for line in lines {
            if let Some((name, value)) = line.split_once(':') {
                headers.insert(name.trim().to_ascii_lowercase(), value.trim().to_string());
            }
        }
        TestRequest {
            method,
            path,
            headers,
            body: body.to_string(),
        }
    }

    fn spawn_mock_server<F>(expected_requests: usize, handler: F) -> String
    where
        F: Fn(&TestRequest) -> MockResponse + Send + 'static,
    {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind mock server");
        let addr = listener.local_addr().expect("mock server local addr");
        thread::spawn(move || {
            for _ in 0..expected_requests {
                let (mut stream, _) = listener.accept().expect("accept mock request");
                let mut buffer = [0_u8; 16 * 1024];
                let size = stream.read(&mut buffer).expect("read mock request");
                let request = parse_request(&String::from_utf8_lossy(&buffer[..size]));
                let response = handler(&request);
                let mut head = vec![
                    format!("HTTP/1.1 {}", response.status),
                    format!("Content-Type: {}", response.content_type),
                    format!("Content-Length: {}", response.body.len()),
                    "Connection: close".to_string(),
                ];
                for (key, value) in response.headers {
                    head.push(format!("{key}: {value}"));
                }
                let payload = format!("{}\r\n\r\n{}", head.join("\r\n"), response.body);
                stream
                    .write_all(payload.as_bytes())
                    .expect("write mock response");
            }
        });
        format!("http://{}", addr)
    }

    fn test_request(message: &str) -> UnifiedRequest {
        UnifiedRequest {
            provider_hint: "grok".to_string(),
            protocol: "openai",
            model: "grok-4".to_string(),
            messages: vec![Message {
                role: "user".to_string(),
                content: Value::String(message.to_string()),
            }],
            system: None,
            stream: false,
        }
    }

    fn test_token() -> Vec<GrokToken> {
        vec![GrokToken {
            id: "grok-1".to_string(),
            name: "Main".to_string(),
            cookie_token: "test-token".to_string(),
            active: true,
        }]
    }

    #[test]
    fn build_grok_cookie_header_merges_cloudflare_values() {
        let header = build_grok_cookie_header("test-token", "theme=dark; cf_clearance=old", "new");
        assert!(header.contains("sso=test-token"));
        assert!(header.contains("sso-rw=test-token"));
        assert!(header.contains("theme=dark"));
        assert!(header.contains("cf_clearance=new"));
        assert!(!header.contains("cf_clearance=old"));
    }

    #[test]
    fn generate_reply_retries_retry_after_and_uses_runtime_config() {
        let attempts = Arc::new(AtomicUsize::new(0));
        let attempts_clone = attempts.clone();
        let mock_url = spawn_mock_server(2, move |request| {
            let attempt = attempts_clone.fetch_add(1, Ordering::SeqCst) + 1;
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/grok");
            assert_eq!(request.header("origin"), Some("https://grok.test"));
            assert_eq!(request.header("referer"), Some("https://grok.test/"));
            assert_eq!(request.header("user-agent"), Some("Mozilla/Test"));
            let cookie = request.header("cookie").unwrap_or_default();
            assert!(cookie.contains("sso=test-token"));
            assert!(cookie.contains("theme=dark"));
            assert!(cookie.contains("cf_clearance=cf-1"));
            let body: Value = serde_json::from_str(&request.body).expect("parse request body");
            assert_eq!(body["modelName"], "grok-4");
            if attempt == 1 {
                return MockResponse {
                    status: "429 Too Many Requests",
                    content_type: "application/json",
                    headers: vec![("Retry-After".to_string(), "1".to_string())],
                    body: "busy".to_string(),
                };
            }
            MockResponse {
                status: "200 OK",
                content_type: "application/json",
                headers: vec![],
                body: "{\"result\":{\"response\":{\"token\":\"OK\"}}}\n".to_string(),
            }
        });

        let mut provider = GrokProvider::new(
            test_token(),
            GrokRuntimeConfig {
                api_url: format!("{mock_url}/grok"),
                proxy_url: String::new(),
                cf_cookies: "theme=dark".to_string(),
                cf_clearance: "cf-1".to_string(),
                user_agent: "Mozilla/Test".to_string(),
                origin: "https://grok.test".to_string(),
                referer: "https://grok.test/".to_string(),
            },
        );
        provider.sleep = |_| {};

        let reply = provider
            .generate_reply(&test_request("reply exactly OK"))
            .expect("grok reply");
        assert_eq!(reply, "OK");
        assert_eq!(attempts.load(Ordering::SeqCst), 2);
    }

    #[test]
    fn generate_reply_retries_after_forbidden() {
        let attempts = Arc::new(AtomicUsize::new(0));
        let attempts_clone = attempts.clone();
        let mock_url = spawn_mock_server(2, move |_| {
            let attempt = attempts_clone.fetch_add(1, Ordering::SeqCst) + 1;
            if attempt == 1 {
                return MockResponse {
                    status: "403 Forbidden",
                    content_type: "application/json",
                    headers: vec![],
                    body: "forbidden".to_string(),
                };
            }
            MockResponse {
                status: "200 OK",
                content_type: "application/json",
                headers: vec![],
                body: "{\"result\":{\"response\":{\"token\":\"retried\"}}}\n".to_string(),
            }
        });

        let mut provider = GrokProvider::new(
            test_token(),
            GrokRuntimeConfig {
                api_url: format!("{mock_url}/grok"),
                proxy_url: String::new(),
                cf_cookies: String::new(),
                cf_clearance: String::new(),
                user_agent: String::new(),
                origin: String::new(),
                referer: String::new(),
            },
        );
        provider.sleep = |_| {};

        let reply = provider
            .generate_reply(&test_request("retry forbidden"))
            .expect("grok reply after forbidden");
        assert_eq!(reply, "retried");
        assert_eq!(attempts.load(Ordering::SeqCst), 2);
    }
}