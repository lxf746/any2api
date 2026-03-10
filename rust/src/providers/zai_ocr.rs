//! Z.ai OCR API client.
//!
//! Endpoint: POST https://ocr.z.ai/api/v1/z-ocr/tasks/process
//! Auth: Bearer JWT token, no signature required.
//! Request: multipart/form-data with "file" field.

use std::fs;
use std::path::Path;
use std::time::{SystemTime, UNIX_EPOCH};

use reqwest::blocking::Client;
use reqwest::blocking::multipart;
use serde_json::{self, json, Value};

const DEFAULT_ENDPOINT: &str = "https://ocr.z.ai/api/v1/z-ocr/tasks/process";
const DEFAULT_AUTH_ENDPOINT: &str = "https://ocr.z.ai/api/v1/z-ocr/auth/";

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

/// Parsed OCR API response.
pub struct OCRResponse {
    pub code: i64,
    pub message: String,
    pub task_id: String,
    pub status: String,
    pub file_name: String,
    pub file_size: i64,
    pub file_type: String,
    pub file_url: String,
    pub created_at: String,
    pub markdown_content: String,
    /// Parsed from the stringified json_content field.
    pub json_content: Option<Value>,
    pub layout: Vec<Value>,
    pub data_info: Value,
    pub timestamp: i64,
    /// The full raw JSON response.
    pub raw: Value,
}

/// Client for the Z.ai OCR API.
pub struct OCRClient {
    pub endpoint: String,
    pub auth_endpoint: String,
    pub token: String,
    client: Client,
}

impl OCRClient {
    /// Create a new client with the given Bearer token.
    pub fn new(token: &str) -> Self {
        Self {
            endpoint: DEFAULT_ENDPOINT.to_string(),
            auth_endpoint: DEFAULT_AUTH_ENDPOINT.to_string(),
            token: token.to_string(),
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

    /// Exchange an OAuth code for an auth token. Auto-sets self.token on success.
    pub fn authenticate(&mut self, code: &str) -> Result<AuthResponse, String> {
        let payload = json!({"code": code}).to_string();
        let resp = self
            .client
            .post(&self.auth_endpoint)
            .header("Content-Type", "application/json")
            .header("Accept", "application/json, text/plain, */*")
            .header("X-Request-ID", generate_uuid())
            .header("Origin", "https://ocr.z.ai")
            .header("Referer", "https://ocr.z.ai/")
            .body(payload)
            .send()
            .map_err(|e| format!("auth http request: {e}"))?;

        let status = resp.status();
        let body = resp.text().map_err(|e| format!("read auth response: {e}"))?;
        if !status.is_success() {
            return Err(format!("auth HTTP {}: {}", status.as_u16(), body));
        }

        let raw: Value = serde_json::from_str(&body).map_err(|e| format!("parse auth json: {e}"))?;
        let result = parse_auth_response(raw);
        if !result.auth_token.is_empty() {
            self.token = result.auth_token.clone();
        }
        Ok(result)
    }

    /// Upload a local file and return the OCR result.
    pub fn process_file(&self, file_path: &Path) -> Result<OCRResponse, String> {
        let data = fs::read(file_path).map_err(|e| format!("read file: {e}"))?;
        let filename = file_path
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("file")
            .to_string();
        self.process_bytes(data, &filename)
    }

    /// Upload raw bytes and return the OCR result.
    pub fn process_bytes(&self, data: Vec<u8>, filename: &str) -> Result<OCRResponse, String> {
        let part = multipart::Part::bytes(data)
            .file_name(filename.to_string())
            .mime_str("application/octet-stream")
            .map_err(|e| format!("build part: {e}"))?;

        let form = multipart::Form::new().part("file", part);

        let resp = self
            .client
            .post(&self.endpoint)
            .bearer_auth(&self.token)
            .header("X-Request-ID", generate_uuid())
            .header("Accept", "application/json, text/plain, */*")
            .header("Origin", "https://ocr.z.ai")
            .header("Referer", "https://ocr.z.ai/")
            .multipart(form)
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
}

fn parse_response(raw: Value) -> OCRResponse {
    let data = raw.get("data").cloned().unwrap_or(Value::Null);
    let s = |key: &str| data.get(key).and_then(Value::as_str).unwrap_or("").to_string();
    let i = |key: &str| data.get(key).and_then(Value::as_i64).unwrap_or(0);

    // json_content is a stringified JSON — double parse
    let json_content = data
        .get("json_content")
        .and_then(Value::as_str)
        .filter(|s| !s.is_empty())
        .and_then(|s| serde_json::from_str::<Value>(s).ok());

    let layout = data
        .get("layout")
        .and_then(Value::as_array)
        .cloned()
        .unwrap_or_default();

    OCRResponse {
        code: raw.get("code").and_then(Value::as_i64).unwrap_or(0),
        message: raw.get("message").and_then(Value::as_str).unwrap_or("").to_string(),
        task_id: s("task_id"),
        status: s("status"),
        file_name: s("file_name"),
        file_size: i("file_size"),
        file_type: s("file_type"),
        file_url: s("file_url"),
        created_at: s("created_at"),
        markdown_content: s("markdown_content"),
        json_content,
        layout,
        data_info: data.get("data_info").cloned().unwrap_or(Value::Null),
        timestamp: i("timestamp"),
        raw,
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

fn generate_uuid() -> String {
    let nanos = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_nanos();
    let mut b = [0u8; 16];
    let mut state = nanos;
    for byte in b.iter_mut() {
        state = state.wrapping_mul(6364136223846793005).wrapping_add(1);
        *byte = (state >> 33) as u8;
    }
    b[6] = (b[6] & 0x0f) | 0x40;
    b[8] = (b[8] & 0x3f) | 0x80;
    format!(
        "{:02x}{:02x}{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}{:02x}{:02x}{:02x}{:02x}",
        b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
        b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]
    )
}
