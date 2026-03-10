//! Z.ai Image Generation API client.
//!
//! Endpoint: POST https://image.z.ai/api/proxy/images/generate
//! Auth: Cookie-based session JWT (not Bearer header).

use std::time::{SystemTime, UNIX_EPOCH};

use reqwest::blocking::Client;
use reqwest::header;
use serde_json::{self, json, Value};

const DEFAULT_ENDPOINT: &str = "https://image.z.ai/api/proxy/images/generate";
const DEFAULT_AUTH_ENDPOINT: &str = "https://image.z.ai/api/v1/z-image/auth/";
const DEFAULT_CALLBACK_ENDPOINT: &str = "https://image.z.ai/api/auth/callback";

/// Parsed auth API response.
pub struct AuthResponse {
    pub code: i64,
    pub message: String,
    pub user_id: String,
    pub auth_token: String,
    pub name: String,
    pub profile_image_url: String,
    pub timestamp: i64,
    pub raw: Value,
}

/// Generated image details.
pub struct ImageInfo {
    pub image_id: String,
    pub prompt: String,
    pub size: String,
    pub ratio: String,
    pub resolution: String,
    pub image_url: String,
    pub status: String,
    pub created_at: String,
    pub updated_at: String,
    pub width: i64,
    pub height: i64,
}

/// Parsed image generation API response.
pub struct ImageResponse {
    pub code: i64,
    pub message: String,
    pub image: ImageInfo,
    pub timestamp: i64,
    pub raw: Value,
}

/// Client for the Z.ai Image Generation API.
pub struct ImageClient {
    pub endpoint: String,
    pub auth_endpoint: String,
    pub callback_endpoint: String,
    pub session_token: String,
    client: Client,
}

impl ImageClient {
    /// Create a new client with the given session JWT token.
    pub fn new(session_token: &str) -> Self {
        Self {
            endpoint: DEFAULT_ENDPOINT.to_string(),
            auth_endpoint: DEFAULT_AUTH_ENDPOINT.to_string(),
            callback_endpoint: DEFAULT_CALLBACK_ENDPOINT.to_string(),
            session_token: session_token.to_string(),
            client: Client::builder()
                .timeout(std::time::Duration::from_secs(120))
                .build()
                .unwrap_or_else(|_| Client::new()),
        }
    }

    /// Create a client by authenticating with an OAuth code.
    pub fn from_code(code: &str) -> Result<(Self, AuthResponse), String> {
        let mut client = Self::new("");
        let auth = client.authenticate(code)?;
        Ok((client, auth))
    }

    /// Full auth flow: code → token → session cookie. Auto-sets session_token.
    pub fn authenticate(&mut self, code: &str) -> Result<AuthResponse, String> {
        // Step 1: exchange code for token
        let payload = json!({"code": code}).to_string();
        let resp = self
            .client
            .post(&self.auth_endpoint)
            .header(header::CONTENT_TYPE, "application/json")
            .header(header::ACCEPT, "*/*")
            .header("X-Request-ID", random_id(22))
            .header(header::ORIGIN, "https://image.z.ai")
            .header(header::REFERER, "https://image.z.ai/")
            .body(payload)
            .send()
            .map_err(|e| format!("auth request: {e}"))?;

        let status = resp.status();
        let body = resp.text().map_err(|e| format!("read auth response: {e}"))?;
        if !status.is_success() {
            return Err(format!("auth HTTP {}: {}", status.as_u16(), body));
        }

        let raw: Value = serde_json::from_str(&body).map_err(|e| format!("parse auth json: {e}"))?;
        let result = parse_auth_response(raw);
        if result.auth_token.is_empty() {
            return Err("auth returned empty token".to_string());
        }

        // Step 2: register token as session cookie
        self.register_callback(&result.auth_token)?;
        self.session_token = result.auth_token.clone();
        Ok(result)
    }

    fn register_callback(&self, token: &str) -> Result<(), String> {
        let payload = json!({"token": token}).to_string();
        let resp = self
            .client
            .post(&self.callback_endpoint)
            .header(header::CONTENT_TYPE, "application/json")
            .header(header::ACCEPT, "*/*")
            .header(header::ORIGIN, "https://image.z.ai")
            .header(header::REFERER, "https://image.z.ai/")
            .body(payload)
            .send()
            .map_err(|e| format!("callback request: {e}"))?;

        if !resp.status().is_success() {
            return Err(format!("callback HTTP {}", resp.status().as_u16()));
        }
        Ok(())
    }

    /// Generate an image from a text prompt with full options.
    pub fn generate(&self, prompt: &str, ratio: &str, resolution: &str,
                    rm_label_watermark: bool) -> Result<ImageResponse, String> {
        let payload = json!({
            "prompt": prompt,
            "ratio": ratio,
            "resolution": resolution,
            "rm_label_watermark": rm_label_watermark,
        });

        let resp = self
            .client
            .post(&self.endpoint)
            .header(header::CONTENT_TYPE, "application/json")
            .header(header::ACCEPT, "*/*")
            .header("X-Request-ID", random_id(22))
            .header(header::ORIGIN, "https://image.z.ai")
            .header(header::REFERER, "https://image.z.ai/create")
            .header(header::COOKIE, format!("session={}", self.session_token))
            .body(payload.to_string())
            .send()
            .map_err(|e| format!("http request: {e}"))?;

        let status = resp.status();
        let body = resp.text().map_err(|e| format!("read response: {e}"))?;
        if !status.is_success() {
            return Err(format!("HTTP {}: {}", status.as_u16(), body));
        }

        let raw: Value = serde_json::from_str(&body).map_err(|e| format!("parse json: {e}"))?;
        Ok(parse_response(raw))
    }

    /// Convenience method with common defaults (1:1, 1K, no watermark).
    pub fn generate_simple(&self, prompt: &str) -> Result<ImageResponse, String> {
        self.generate(prompt, "1:1", "1K", true)
    }
}

fn parse_auth_response(raw: Value) -> AuthResponse {
    let data = raw.get("data").cloned().unwrap_or(Value::Null);
    let s = |key: &str| data.get(key).and_then(Value::as_str).unwrap_or("").to_string();
    AuthResponse {
        code: raw.get("code").and_then(Value::as_i64).unwrap_or(0),
        message: raw.get("message").and_then(Value::as_str).unwrap_or("").to_string(),
        user_id: s("user_id"),
        auth_token: s("auth_token"),
        name: s("name"),
        profile_image_url: s("profile_image_url"),
        timestamp: raw.get("timestamp").and_then(Value::as_i64).unwrap_or(0),
        raw,
    }
}

fn parse_response(raw: Value) -> ImageResponse {
    let data = raw.get("data").cloned().unwrap_or(Value::Null);
    let img = data.get("image").cloned().unwrap_or(Value::Null);
    let s = |v: &Value, key: &str| v.get(key).and_then(Value::as_str).unwrap_or("").to_string();
    let i = |v: &Value, key: &str| v.get(key).and_then(Value::as_i64).unwrap_or(0);

    ImageResponse {
        code: raw.get("code").and_then(Value::as_i64).unwrap_or(0),
        message: raw.get("message").and_then(Value::as_str).unwrap_or("").to_string(),
        image: ImageInfo {
            image_id: s(&img, "image_id"),
            prompt: s(&img, "prompt"),
            size: s(&img, "size"),
            ratio: s(&img, "ratio"),
            resolution: s(&img, "resolution"),
            image_url: s(&img, "image_url"),
            status: s(&img, "status"),
            created_at: s(&img, "created_at"),
            updated_at: s(&img, "updated_at"),
            width: i(&img, "width"),
            height: i(&img, "height"),
        },
        timestamp: raw.get("timestamp").and_then(Value::as_i64).unwrap_or(0),
        raw,
    }
}

fn random_id(length: usize) -> String {
    const CHARS: &[u8] = b"abcdefghijklmnopqrstuvwxyz0123456789";
    let mut state = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_nanos();
    let mut result = Vec::with_capacity(length);
    for _ in 0..length {
        state = state.wrapping_mul(6364136223846793005).wrapping_add(1442695040888963407);
        result.push(CHARS[(state >> 33) as usize % CHARS.len()]);
    }
    String::from_utf8(result).unwrap_or_default()
}
