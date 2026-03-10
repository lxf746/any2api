//! Z.ai TTS (Text-to-Speech) API client.
//!
//! Endpoint: POST https://audio.z.ai/api/v1/z-audio/tts/create
//! Auth: Bearer JWT token.
//! Response: SSE stream with {"audio":"<base64 WAV>"} chunks, ending with [DONE].

use reqwest::blocking::Client;
use reqwest::header;
use serde_json::{self, json, Value};

const DEFAULT_ENDPOINT: &str = "https://audio.z.ai/api/v1/z-audio/tts/create";
const DEFAULT_AUTH_ENDPOINT: &str = "https://audio.z.ai/api/v1/z-audio/auth/";

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

/// Client for the Z.ai TTS API.
pub struct TTSClient {
    pub endpoint: String,
    pub auth_endpoint: String,
    pub token: String,
    pub user_id: String,
    client: Client,
}

impl TTSClient {
    /// Create a new client with the given Bearer token and user ID.
    pub fn new(token: &str, user_id: &str) -> Self {
        Self {
            endpoint: DEFAULT_ENDPOINT.to_string(),
            auth_endpoint: DEFAULT_AUTH_ENDPOINT.to_string(),
            token: token.to_string(),
            user_id: user_id.to_string(),
            client: Client::builder()
                .timeout(std::time::Duration::from_secs(120))
                .build()
                .unwrap_or_else(|_| Client::new()),
        }
    }

    /// Create a client by authenticating with an OAuth code.
    pub fn from_code(code: &str) -> Result<(Self, AuthResponse), String> {
        let mut client = Self::new("", "");
        let auth = client.authenticate(code)?;
        Ok((client, auth))
    }

    /// Exchange an OAuth code for a token. Auto-sets token and user_id.
    pub fn authenticate(&mut self, code: &str) -> Result<AuthResponse, String> {
        let payload = json!({"code": code}).to_string();
        let resp = self.client
            .post(&self.auth_endpoint)
            .header(header::CONTENT_TYPE, "application/json")
            .header(header::ACCEPT, "*/*")
            .header(header::ORIGIN, "https://audio.z.ai")
            .header(header::REFERER, "https://audio.z.ai/")
            .body(payload)
            .send()
            .map_err(|e| format!("auth request: {e}"))?;

        let status = resp.status();
        let body = resp.text().map_err(|e| format!("read auth response: {e}"))?;
        if !status.is_success() {
            return Err(format!("auth HTTP {}: {}", status.as_u16(), body));
        }

        let raw: Value = serde_json::from_str(&body).map_err(|e| format!("parse auth: {e}"))?;
        let data = raw.get("data").cloned().unwrap_or(Value::Null);
        let s = |key: &str| data.get(key).and_then(Value::as_str).unwrap_or("").to_string();

        let result = AuthResponse {
            code: raw.get("code").and_then(Value::as_i64).unwrap_or(0),
            message: raw.get("message").and_then(Value::as_str).unwrap_or("").to_string(),
            user_id: s("user_id"),
            auth_token: s("auth_token"),
            name: s("name"),
            profile_image_url: s("profile_image_url"),
            timestamp: raw.get("timestamp").and_then(Value::as_i64).unwrap_or(0),
            raw,
        };
        if !result.auth_token.is_empty() {
            self.token = result.auth_token.clone();
        }
        if !result.user_id.is_empty() {
            self.user_id = result.user_id.clone();
        }
        Ok(result)
    }

    /// Convert text to speech. Returns WAV audio bytes.
    pub fn synthesize(&self, text: &str, voice_id: &str, voice_name: &str,
                      speed: f64, volume: f64) -> Result<Vec<u8>, String> {
        let payload = json!({
            "voice_name": voice_name,
            "voice_id": voice_id,
            "user_id": self.user_id,
            "input_text": text,
            "speed": speed,
            "volume": volume,
        });

        let resp = self.client
            .post(&self.endpoint)
            .header(header::CONTENT_TYPE, "application/json")
            .header(header::ACCEPT, "text/event-stream")
            .bearer_auth(&self.token)
            .header(header::ORIGIN, "https://audio.z.ai")
            .header(header::REFERER, "https://audio.z.ai/")
            .body(payload.to_string())
            .send()
            .map_err(|e| format!("http request: {e}"))?;

        let status = resp.status();
        if !status.is_success() {
            let body = resp.text().unwrap_or_default();
            return Err(format!("HTTP {}: {}", status.as_u16(), body));
        }

        let body = resp.text().map_err(|e| format!("read response: {e}"))?;
        read_sse_audio(&body)
    }

    /// Convenience method with default voice settings.
    pub fn synthesize_simple(&self, text: &str) -> Result<Vec<u8>, String> {
        self.synthesize(text, "system_003", "通用男声", 1.0, 1.0)
    }
}

/// Parse SSE text and collect base64 audio chunks into raw bytes.
fn read_sse_audio(body: &str) -> Result<Vec<u8>, String> {
    let mut audio_data = Vec::new();
    for line in body.lines() {
        let Some(data) = line.strip_prefix("data: ") else { continue };
        if data == "[DONE]" {
            break;
        }
        let Ok(parsed) = serde_json::from_str::<Value>(data) else { continue };
        let Some(audio_b64) = parsed.get("audio").and_then(Value::as_str) else { continue };
        if audio_b64.is_empty() {
            continue;
        }
        // Use a simple base64 decoder — reqwest doesn't include one,
        // so we do a minimal inline decode.
        let decoded = base64_decode(audio_b64)?;
        audio_data.extend_from_slice(&decoded);
    }
    Ok(audio_data)
}

/// Minimal base64 decoder (standard alphabet, with padding).
fn base64_decode(input: &str) -> Result<Vec<u8>, String> {
    const TABLE: [u8; 128] = {
        let mut t = [255u8; 128];
        let mut i = 0u8;
        while i < 26 { t[(b'A' + i) as usize] = i; i += 1; }
        i = 0;
        while i < 26 { t[(b'a' + i) as usize] = 26 + i; i += 1; }
        i = 0;
        while i < 10 { t[(b'0' + i) as usize] = 52 + i; i += 1; }
        t[b'+' as usize] = 62;
        t[b'/' as usize] = 63;
        t
    };

    let bytes: Vec<u8> = input.bytes().filter(|&b| b != b'=' && b != b'\n' && b != b'\r').collect();
    let mut out = Vec::with_capacity(bytes.len() * 3 / 4);
    let chunks = bytes.chunks(4);
    for chunk in chunks {
        let mut buf = [0u8; 4];
        for (i, &b) in chunk.iter().enumerate() {
            if b > 127 || TABLE[b as usize] == 255 {
                return Err(format!("invalid base64 byte: {b}"));
            }
            buf[i] = TABLE[b as usize];
        }
        let n = chunk.len();
        if n >= 2 { out.push((buf[0] << 2) | (buf[1] >> 4)); }
        if n >= 3 { out.push((buf[1] << 4) | (buf[2] >> 2)); }
        if n >= 4 { out.push((buf[2] << 6) | buf[3]); }
    }
    Ok(out)
}
