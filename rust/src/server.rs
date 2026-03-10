use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::{Mutex, OnceLock};
use std::time::{SystemTime, UNIX_EPOCH};

use serde_json::{json, Value};

use crate::admin_store::{
    json_array_field, json_array_objects, json_bool_field, json_escape, json_object_field,
    json_string_field, AdminConfig, AdminSettings, AdminStore, ChatGPTRuntimeConfig,
    CursorRuntimeConfig, GrokRuntimeConfig, GrokToken, KiroAccount, OrchidsRuntimeConfig,
    WebRuntimeConfig, ZAIOCRRuntimeConfig, ZAIImageRuntimeConfig, ZAITTSRuntimeConfig,
};
use crate::providers::{default_registry, zai_image, zai_ocr, zai_tts};
use crate::registry::Registry;
use crate::types::{Message, UnifiedRequest};

const ADMIN_SESSION_COOKIE: &str = "newplatform2api_admin_session";
const ADMIN_AUTH_MODE: &str = "session_cookie";
const ADMIN_BACKEND_VERSION: &str = "0.1.0";
const DEFAULT_ZAI_IMAGE_API_URL: &str = "https://image.z.ai/api/proxy/images/generate";
const DEFAULT_ZAI_TTS_API_URL: &str = "https://audio.z.ai/api/v1/z-audio/tts/create";
const DEFAULT_ZAI_OCR_API_URL: &str = "https://ocr.z.ai/api/v1/z-ocr/tasks/process";

#[derive(Debug, Clone)]
struct ParsedRequest {
    method: String,
    path: String,
    headers: HashMap<String, String>,
    body: String,
    body_bytes: Vec<u8>,
}

impl ParsedRequest {
    fn header(&self, name: &str) -> Option<&str> {
        self.headers
            .get(&name.to_ascii_lowercase())
            .map(String::as_str)
    }
}

struct AdminSessionStore {
    sessions: Mutex<HashMap<String, u64>>,
    counter: AtomicU64,
}

impl AdminSessionStore {
    fn new() -> Self {
        Self {
            sessions: Mutex::new(HashMap::new()),
            counter: AtomicU64::new(1),
        }
    }

    fn create(&self) -> (String, u64) {
        let expires_at = now_unix() + 86_400;
        let token = format!(
            "{:x}{:x}",
            expires_at,
            self.counter.fetch_add(1, Ordering::Relaxed)
        );
        self.sessions
            .lock()
            .expect("admin sessions lock poisoned")
            .insert(token.clone(), expires_at);
        (token, expires_at)
    }

    fn expires_at(&self, token: &str) -> Option<u64> {
        let trimmed = token.trim();
        if trimmed.is_empty() {
            return None;
        }
        let mut sessions = self.sessions.lock().expect("admin sessions lock poisoned");
        let expires_at = sessions.get(trimmed).copied()?;
        if now_unix() >= expires_at {
            sessions.remove(trimmed);
            return None;
        }
        Some(expires_at)
    }

    fn delete(&self, token: &str) {
        let trimmed = token.trim();
        if trimmed.is_empty() {
            return;
        }
        self.sessions
            .lock()
            .expect("admin sessions lock poisoned")
            .remove(trimmed);
    }

    #[cfg(test)]
    fn clear(&self) {
        self.sessions
            .lock()
            .expect("admin sessions lock poisoned")
            .clear();
    }
}

static ADMIN_SESSIONS: OnceLock<AdminSessionStore> = OnceLock::new();
static ADMIN_STORE: OnceLock<AdminStore> = OnceLock::new();
static RESPONSE_COUNTER: AtomicU64 = AtomicU64::new(1);

fn admin_sessions() -> &'static AdminSessionStore {
    ADMIN_SESSIONS.get_or_init(AdminSessionStore::new)
}

fn admin_store() -> &'static AdminStore {
    ADMIN_STORE.get_or_init(AdminStore::from_env)
}

fn now_unix() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs()
}

fn response(
    status: &str,
    content_type: &str,
    headers: Vec<(String, String)>,
    body: String,
) -> String {
    let mut lines = vec![
        format!("HTTP/1.1 {status}"),
        format!("Content-Type: {content_type}"),
        format!("Content-Length: {}", body.len()),
        "Connection: close".to_string(),
    ];
    for (key, value) in headers {
        lines.push(format!("{key}: {value}"));
    }
    format!("{}\r\n\r\n{}", lines.join("\r\n"), body)
}

fn response_bytes(
    status: &str,
    content_type: &str,
    headers: Vec<(String, String)>,
    body: Vec<u8>,
) -> Vec<u8> {
    let mut lines = vec![
        format!("HTTP/1.1 {status}"),
        format!("Content-Type: {content_type}"),
        format!("Content-Length: {}", body.len()),
        "Connection: close".to_string(),
    ];
    for (key, value) in headers {
        lines.push(format!("{key}: {value}"));
    }
    let mut rendered = format!("{}\r\n\r\n", lines.join("\r\n")).into_bytes();
    rendered.extend_from_slice(&body);
    rendered
}

fn json_response(status: &str, headers: Vec<(String, String)>, body: String) -> String {
    response(status, "application/json; charset=utf-8", headers, body)
}

fn json_response_value(status: &str, headers: Vec<(String, String)>, body: Value) -> String {
    json_response(status, headers, body.to_string())
}

fn event_stream_response(status: &str, headers: Vec<(String, String)>, body: String) -> String {
    response(status, "text/event-stream", headers, body)
}

fn next_response_id(prefix: &str) -> String {
    format!(
        "{}_{:x}{:x}",
        prefix,
        now_unix(),
        RESPONSE_COUNTER.fetch_add(1, Ordering::Relaxed)
    )
}

fn json_error_response(status: &str, headers: Vec<(String, String)>, message: &str) -> String {
    json_response_value(status, headers, json!({"error": message}))
}

fn parse_request(raw: &[u8]) -> ParsedRequest {
    let split_index = raw
        .windows(4)
        .position(|window| window == b"\r\n\r\n");
    let (head_bytes, body_bytes) = match split_index {
        Some(index) => (&raw[..index], raw[index + 4..].to_vec()),
        None => (raw, Vec::new()),
    };
    let head = String::from_utf8_lossy(head_bytes).to_string();
    let mut lines = head.lines();
    let request_line = lines.next().unwrap_or("");
    let mut parts = request_line.split_whitespace();
    let method = parts.next().unwrap_or("").to_string();
    let path = parts.next().unwrap_or("/").to_string();
    let mut headers = HashMap::new();
    for line in lines {
        if let Some((name, value)) = line.split_once(':') {
            headers.insert(name.trim().to_ascii_lowercase(), value.trim().to_string());
        }
    }
    ParsedRequest {
        method,
        path,
        headers,
        body: String::from_utf8_lossy(&body_bytes).to_string(),
        body_bytes,
    }
}

fn parse_json_object_body(body: &str) -> Result<Value, String> {
    let value = serde_json::from_str::<Value>(body).map_err(|_| "invalid json".to_string())?;
    if !value.is_object() {
        return Err("invalid json".to_string());
    }
    Ok(value)
}

fn provider_from_payload(payload: &Value) -> Option<String> {
    payload
        .get("provider")
        .and_then(Value::as_str)
        .map(str::trim)
        .filter(|item| !item.is_empty())
        .map(ToString::to_string)
}

fn payload_string(payload: &Value, key: &str) -> String {
    payload
        .get(key)
        .and_then(Value::as_str)
        .unwrap_or("")
        .trim()
        .to_string()
}

fn payload_bool(payload: &Value, key: &str) -> bool {
    payload.get(key).and_then(Value::as_bool).unwrap_or(false)
}

fn payload_messages(payload: &Value) -> Vec<Message> {
    payload
        .get("messages")
        .and_then(Value::as_array)
        .map(|items| {
            items
                .iter()
                .filter_map(|item| {
                    let obj = item.as_object()?;
                    Some(Message {
                        role: obj
                            .get("role")
                            .and_then(Value::as_str)
                            .unwrap_or("user")
                            .trim()
                            .to_string(),
                        content: obj.get("content").cloned().unwrap_or(Value::Null),
                    })
                })
                .collect::<Vec<_>>()
        })
        .unwrap_or_default()
}

fn env_string(default: &str, keys: &[&str]) -> String {
    for key in keys {
        if let Ok(value) = std::env::var(key) {
            let trimmed = value.trim();
            if !trimmed.is_empty() {
                return trimmed.to_string();
            }
        }
    }
    default.trim().to_string()
}

fn current_zai_image_config(snapshot: &AdminConfig) -> ZAIImageRuntimeConfig {
    let config = &snapshot.providers.zai_image_config;
    ZAIImageRuntimeConfig {
        session_token: env_string(
            &config.session_token,
            &["NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN", "ZAI_IMAGE_SESSION_TOKEN"],
        ),
        api_url: env_string(&config.api_url, &["NEWPLATFORM2API_ZAI_IMAGE_API_URL"]),
    }
}

fn current_zai_tts_config(snapshot: &AdminConfig) -> ZAITTSRuntimeConfig {
    let config = &snapshot.providers.zai_tts_config;
    ZAITTSRuntimeConfig {
        token: env_string(
            &config.token,
            &["NEWPLATFORM2API_ZAI_TTS_TOKEN", "ZAI_TTS_TOKEN"],
        ),
        user_id: env_string(
            &config.user_id,
            &["NEWPLATFORM2API_ZAI_TTS_USER_ID", "ZAI_TTS_USER_ID"],
        ),
        api_url: env_string(&config.api_url, &["NEWPLATFORM2API_ZAI_TTS_API_URL"]),
    }
}

fn current_zai_ocr_config(snapshot: &AdminConfig) -> ZAIOCRRuntimeConfig {
    let config = &snapshot.providers.zai_ocr_config;
    ZAIOCRRuntimeConfig {
        token: env_string(
            &config.token,
            &["NEWPLATFORM2API_ZAI_OCR_TOKEN", "ZAI_OCR_TOKEN"],
        ),
        api_url: env_string(&config.api_url, &["NEWPLATFORM2API_ZAI_OCR_API_URL"]),
    }
}

fn provider_options_from_payload(
    payload: &Value,
) -> Result<Option<&serde_json::Map<String, Value>>, String> {
    match payload.get("provider_options") {
        None | Some(Value::Null) => Ok(None),
        Some(Value::Object(map)) => Ok(Some(map)),
        Some(_) => Err("provider_options must be an object".to_string()),
    }
}

fn option_string(value: Option<&Value>) -> String {
    match value {
        Some(Value::String(item)) => item.trim().to_string(),
        Some(Value::Number(item)) => item.to_string(),
        Some(Value::Bool(item)) => item.to_string(),
        _ => String::new(),
    }
}

fn option_bool(value: Option<&Value>, default: bool) -> bool {
    match value {
        None | Some(Value::Null) => default,
        Some(Value::Bool(item)) => *item,
        Some(Value::String(item)) => match item.trim().to_ascii_lowercase().as_str() {
            "1" | "true" | "yes" | "on" => true,
            "0" | "false" | "no" | "off" => false,
            _ => true,
        },
        Some(Value::Number(item)) => item.as_i64().map(|number| number != 0).unwrap_or(true),
        _ => true,
    }
}

fn option_int(value: Option<&Value>, default: i64) -> Result<i64, String> {
    match value {
        None | Some(Value::Null) => Ok(default),
        Some(Value::Number(item)) => {
            if let Some(parsed) = item.as_i64() {
                return Ok(parsed);
            }
            let parsed = item
                .as_f64()
                .ok_or_else(|| "invalid integer".to_string())?;
            if parsed.trunc() != parsed {
                return Err("invalid integer".to_string());
            }
            Ok(parsed as i64)
        }
        Some(Value::String(item)) => {
            let trimmed = item.trim();
            if trimmed.is_empty() {
                return Ok(default);
            }
            trimmed
                .parse::<i64>()
                .map_err(|_| "invalid integer".to_string())
        }
        _ => Err("invalid integer".to_string()),
    }
}

fn option_float(value: Option<&Value>, default: f64) -> Result<f64, String> {
    match value {
        None | Some(Value::Null) => Ok(default),
        Some(Value::Number(item)) => item.as_f64().ok_or_else(|| "invalid number".to_string()),
        Some(Value::String(item)) => {
            let trimmed = item.trim();
            if trimmed.is_empty() {
                return Ok(default);
            }
            trimmed
                .parse::<f64>()
                .map_err(|_| "invalid number".to_string())
        }
        _ => Err("invalid number".to_string()),
    }
}

fn image_size_settings(size: &str) -> Option<(&'static str, &'static str)> {
    match size.trim() {
        "1024x1024" => Some(("1:1", "1K")),
        "1024x1792" => Some(("9:16", "2K")),
        "1792x1024" => Some(("16:9", "2K")),
        _ => None,
    }
}

fn resolve_image_settings(
    payload: &Value,
    options: Option<&serde_json::Map<String, Value>>,
) -> Result<(String, String), String> {
    let ratio = {
        let value = option_string(payload.get("ratio"));
        if value.is_empty() {
            options
                .and_then(|map| map.get("ratio"))
                .map(|value| option_string(Some(value)))
                .unwrap_or_default()
        } else {
            value
        }
    };
    let resolution = {
        let value = option_string(payload.get("resolution"));
        if value.is_empty() {
            options
                .and_then(|map| map.get("resolution"))
                .map(|value| option_string(Some(value)))
                .unwrap_or_default()
        } else {
            value
        }
    };
    if !ratio.is_empty() && !resolution.is_empty() {
            return Ok((ratio, resolution));
    }
    let size = {
        let value = option_string(payload.get("size"));
        if value.is_empty() {
            options
                .and_then(|map| map.get("size"))
                .map(|value| option_string(Some(value)))
                .unwrap_or_default()
        } else {
            value
        }
    };
    if !size.is_empty() {
        let Some((ratio, resolution)) = image_size_settings(&size) else {
            return Err(format!("unsupported size: {size}"));
        };
        return Ok((ratio.to_string(), resolution.to_string()));
    }
    Ok((
        if ratio.is_empty() { "1:1".to_string() } else { ratio },
        if resolution.is_empty() {
            "1K".to_string()
        } else {
            resolution
        },
    ))
}

fn multimedia_provider_headers(provider: &str) -> Vec<(String, String)> {
    vec![(
        "X-Newplatform2API-Provider".to_string(),
        provider.to_string(),
    )]
}

fn current_zai_image_client(snapshot: &AdminConfig) -> Result<zai_image::ImageClient, String> {
    let config = current_zai_image_config(snapshot);
    if config.session_token.is_empty() {
        return Err("zai image is not configured".to_string());
    }
    let mut client = zai_image::ImageClient::new(&config.session_token);
    client.endpoint = if config.api_url.is_empty() {
        DEFAULT_ZAI_IMAGE_API_URL.to_string()
    } else {
        config.api_url
    };
    Ok(client)
}

fn current_zai_tts_client(snapshot: &AdminConfig) -> Result<zai_tts::TTSClient, String> {
    let config = current_zai_tts_config(snapshot);
    if config.token.is_empty() || config.user_id.is_empty() {
        return Err("zai tts is not configured".to_string());
    }
    let mut client = zai_tts::TTSClient::new(&config.token, &config.user_id);
    client.endpoint = if config.api_url.is_empty() {
        DEFAULT_ZAI_TTS_API_URL.to_string()
    } else {
        config.api_url
    };
    Ok(client)
}

fn current_zai_ocr_client(snapshot: &AdminConfig) -> Result<zai_ocr::OCRClient, String> {
    let config = current_zai_ocr_config(snapshot);
    if config.token.is_empty() {
        return Err("zai ocr is not configured".to_string());
    }
    let mut client = zai_ocr::OCRClient::new(&config.token);
    client.endpoint = if config.api_url.is_empty() {
        DEFAULT_ZAI_OCR_API_URL.to_string()
    } else {
        config.api_url
    };
    Ok(client)
}

fn find_bytes(haystack: &[u8], needle: &[u8]) -> Option<usize> {
    haystack.windows(needle.len()).position(|window| window == needle)
}

fn trim_crlf(mut input: &[u8]) -> &[u8] {
    while input.starts_with(b"\r\n") {
        input = &input[2..];
    }
    while input.ends_with(b"\r\n") {
        input = &input[..input.len() - 2];
    }
    input
}

fn multipart_boundary(content_type: &str) -> Option<String> {
    content_type.split(';').find_map(|part| {
        let trimmed = part.trim();
        let value = trimmed.strip_prefix("boundary=")?;
        Some(value.trim_matches('"').to_string())
    })
}

fn disposition_param(value: &str, key: &str) -> Option<String> {
    value.split(';').find_map(|segment| {
        let trimmed = segment.trim();
        let expected = format!("{key}=");
        let raw = trimmed.strip_prefix(&expected)?;
        Some(raw.trim_matches('"').to_string())
    })
}

fn multipart_file(request: &ParsedRequest) -> Result<(String, Vec<u8>), String> {
    let content_type = request
        .header("content-type")
        .ok_or_else(|| "content-type must be multipart/form-data".to_string())?;
    if !content_type.to_ascii_lowercase().contains("multipart/form-data") {
        return Err("content-type must be multipart/form-data".to_string());
    }
    let boundary = multipart_boundary(content_type)
        .ok_or_else(|| "invalid multipart form-data".to_string())?;
    let marker = format!("--{boundary}").into_bytes();
    let mut cursor = 0_usize;
    while let Some(offset) = find_bytes(&request.body_bytes[cursor..], &marker) {
        let marker_start = cursor + offset;
        let mut part_start = marker_start + marker.len();
        if request.body_bytes[part_start..].starts_with(b"--") {
            break;
        }
        if request.body_bytes[part_start..].starts_with(b"\r\n") {
            part_start += 2;
        }
        let Some(next_offset) = find_bytes(&request.body_bytes[part_start..], &marker) else {
            return Err("invalid multipart form-data".to_string());
        };
        let part_end = part_start + next_offset;
        let part = trim_crlf(&request.body_bytes[part_start..part_end]);
        cursor = part_end;
        if part.is_empty() {
            continue;
        }
        let Some(header_end) = find_bytes(part, b"\r\n\r\n") else {
            continue;
        };
        let headers = String::from_utf8_lossy(&part[..header_end]).to_string();
        let content = part[header_end + 4..].to_vec();
        let disposition = headers.lines().find_map(|line| {
            let Some((name, value)) = line.split_once(':') else {
                return None;
            };
            if name.trim().eq_ignore_ascii_case("content-disposition") {
                Some(value.trim().to_string())
            } else {
                None
            }
        });
        let Some(disposition) = disposition else {
            continue;
        };
        if disposition_param(&disposition, "name").as_deref() != Some("file") {
            continue;
        }
        let filename = disposition_param(&disposition, "filename")
            .filter(|value| !value.trim().is_empty())
            .unwrap_or_else(|| "upload.bin".to_string());
        return Ok((filename, content));
    }
    Err("file is required".to_string())
}

fn handle_openai_images_generation(request: &ParsedRequest, snapshot: &AdminConfig) -> String {
    let headers = multimedia_provider_headers("zai_image");
    let payload = match parse_json_object_body(&request.body) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err),
    };
    let prompt = option_string(payload.get("prompt"));
    if prompt.is_empty() {
        return json_error_response("400 Bad Request", headers, "prompt is required");
    }
    let n = match option_int(payload.get("n"), 1) {
        Ok(value) => value,
        Err(_) => return json_error_response("400 Bad Request", headers, "n must be an integer"),
    };
    if n != 1 {
        return json_error_response("400 Bad Request", headers, "only n=1 is supported");
    }
    let response_format = option_string(payload.get("response_format")).to_ascii_lowercase();
    if !response_format.is_empty() && response_format != "url" {
        return json_error_response(
            "400 Bad Request",
            headers,
            "only response_format=url is supported",
        );
    }
    let options = match provider_options_from_payload(&payload) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err),
    };
    let (ratio, resolution) = match resolve_image_settings(&payload, options) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err),
    };
    let rm_label_watermark = option_bool(
        payload
            .get("rm_label_watermark")
            .or_else(|| options.and_then(|map| map.get("rm_label_watermark"))),
        true,
    );
    let client = match current_zai_image_client(snapshot) {
        Ok(value) => value,
        Err(err) => return json_error_response("503 Service Unavailable", headers, &err),
    };
    match client.generate(&prompt, &ratio, &resolution, rm_label_watermark) {
        Ok(result) => {
            let created = if result.timestamp > 0 {
                result.timestamp
            } else {
                now_unix() as i64
            };
            let image = result.image;
            let size = if !image.size.trim().is_empty() {
                image.size.clone()
            } else if image.width > 0 && image.height > 0 {
                format!("{}x{}", image.width, image.height)
            } else {
                String::new()
            };
            json_response_value(
                "200 OK",
                headers,
                json!({
                    "created": created,
                    "data": [{
                        "url": image.image_url,
                        "revised_prompt": if image.prompt.trim().is_empty() { prompt } else { image.prompt },
                        "size": size,
                        "width": image.width,
                        "height": image.height,
                        "ratio": if image.ratio.trim().is_empty() { ratio } else { image.ratio },
                        "resolution": if image.resolution.trim().is_empty() { resolution } else { image.resolution },
                    }],
                }),
            )
        }
        Err(err) => json_error_response("502 Bad Gateway", headers, &err),
    }
}

fn handle_openai_audio_speech_bytes(request: &ParsedRequest, snapshot: &AdminConfig) -> Vec<u8> {
    let headers = multimedia_provider_headers("zai_tts");
    let payload = match parse_json_object_body(&request.body) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err).into_bytes(),
    };
    let text = option_string(payload.get("input"));
    let text = if text.is_empty() {
        option_string(payload.get("text"))
    } else {
        text
    };
    if text.is_empty() {
        return json_error_response("400 Bad Request", headers, "input is required").into_bytes();
    }
    let response_format = option_string(payload.get("response_format")).to_ascii_lowercase();
    if !response_format.is_empty() && response_format != "wav" {
        return json_error_response(
            "400 Bad Request",
            headers,
            "only response_format=wav is supported",
        )
        .into_bytes();
    }
    let options = match provider_options_from_payload(&payload) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err).into_bytes(),
    };
    let voice_id = option_string(payload.get("voice_id"));
    let voice_id = if voice_id.is_empty() {
        let voice = option_string(payload.get("voice"));
        if voice.is_empty() {
            options
                .and_then(|map| map.get("voice_id"))
                .map(|value| option_string(Some(value)))
                .filter(|value| !value.is_empty())
                .unwrap_or_else(|| "system_003".to_string())
        } else {
            voice
        }
    } else {
        voice_id
    };
    let voice_name = option_string(payload.get("voice_name"));
    let voice_name = if voice_name.is_empty() {
        options
            .and_then(|map| map.get("voice_name"))
            .map(|value| option_string(Some(value)))
            .filter(|value| !value.is_empty())
            .unwrap_or_else(|| "通用男声".to_string())
    } else {
        voice_name
    };
    let speed = match option_float(
        payload
            .get("speed")
            .or_else(|| options.and_then(|map| map.get("speed"))),
        1.0,
    ) {
        Ok(value) => value,
        Err(_) => {
            return json_error_response(
                "400 Bad Request",
                headers,
                "speed and volume must be numbers",
            )
            .into_bytes()
        }
    };
    let volume = match option_float(
        payload
            .get("volume")
            .or_else(|| options.and_then(|map| map.get("volume"))),
        1.0,
    ) {
        Ok(value) => value,
        Err(_) => {
            return json_error_response(
                "400 Bad Request",
                headers,
                "speed and volume must be numbers",
            )
            .into_bytes()
        }
    };
    let client = match current_zai_tts_client(snapshot) {
        Ok(value) => value,
        Err(err) => {
            return json_error_response("503 Service Unavailable", headers, &err).into_bytes()
        }
    };
    match client.synthesize(&text, &voice_id, &voice_name, speed, volume) {
        Ok(audio) => response_bytes("200 OK", "audio/wav", headers, audio),
        Err(err) => json_error_response("502 Bad Gateway", headers, &err).into_bytes(),
    }
}

fn handle_ocr_upload(request: &ParsedRequest, snapshot: &AdminConfig) -> String {
    let headers = multimedia_provider_headers("zai_ocr");
    let (filename, content) = match multipart_file(request) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", headers, &err),
    };
    let client = match current_zai_ocr_client(snapshot) {
        Ok(value) => value,
        Err(err) => return json_error_response("503 Service Unavailable", headers, &err),
    };
    match client.process_bytes(content, &filename) {
        Ok(result) => json_response_value(
            "200 OK",
            headers,
            json!({
                "id": result.task_id,
                "object": "ocr.result",
                "model": "zai-ocr",
                "status": result.status,
                "text": result.markdown_content,
                "markdown": result.markdown_content,
                "json": result.json_content,
                "layout": result.layout,
                "file": {
                    "name": result.file_name,
                    "size": result.file_size,
                    "type": result.file_type,
                    "url": result.file_url,
                    "created_at": result.created_at,
                }
            }),
        ),
        Err(err) => json_error_response("502 Bad Gateway", headers, &err),
    }
}

fn provider_headers(provider_id: &str) -> Vec<(String, String)> {
    vec![(
        "X-Newplatform2API-Provider".to_string(),
        provider_id.to_string(),
    )]
}

fn build_openai_request(payload: &Value, provider_hint: String) -> UnifiedRequest {
    UnifiedRequest {
        provider_hint,
        protocol: "openai",
        model: payload_string(payload, "model"),
        messages: payload_messages(payload),
        system: None,
        stream: payload_bool(payload, "stream"),
    }
}

fn build_anthropic_request(payload: &Value, provider_hint: String) -> UnifiedRequest {
    UnifiedRequest {
        provider_hint,
        protocol: "anthropic",
        model: payload_string(payload, "model"),
        messages: payload_messages(payload),
        system: payload.get("system").cloned(),
        stream: payload_bool(payload, "stream"),
    }
}

fn openai_sse_body(text: &str, completion_id: &str, model: &str) -> String {
    let chunk = json!({
        "id": completion_id,
        "object": "chat.completion.chunk",
        "model": model,
        "choices": [{
            "index": 0,
            "delta": {"role": "assistant", "content": text},
            "finish_reason": Value::Null
        }]
    });
    format!("data: {}\n\ndata: [DONE]\n\n", chunk)
}

fn anthropic_sse_body(text: &str, message_id: &str) -> String {
    let start = json!({"type": "message_start", "message": {"id": message_id}});
    let delta = json!({
        "type": "content_block_delta",
        "delta": {"type": "text_delta", "text": text}
    });
    let stop = json!({"type": "message_stop"});
    format!(
        "event: message_start\ndata: {}\n\nevent: content_block_delta\ndata: {}\n\nevent: message_stop\ndata: {}\n\n",
        start, delta, stop
    )
}

fn handle_openai_chat(
    request: &ParsedRequest,
    registry: &Registry,
    provider_hint_from_path: Option<&str>,
) -> String {
    let payload = match parse_json_object_body(&request.body) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", vec![], &err),
    };
    let provider_hint = provider_from_payload(&payload)
        .or_else(|| provider_hint_from_path.map(ToString::to_string));
    let provider = match registry.resolve(provider_hint.as_deref()) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", vec![], &err),
    };
    if !provider.capabilities().openai_compatible {
        return json_error_response(
            "400 Bad Request",
            provider_headers(provider.id()),
            "provider does not support /v1/chat/completions",
        );
    }
    let req = build_openai_request(&payload, provider_hint.unwrap_or_default());
    match provider.generate_reply(&req) {
        Ok(text) => {
            let completion_id = next_response_id("chatcmpl");
            let headers = provider_headers(provider.id());
            if req.stream {
                return event_stream_response(
                    "200 OK",
                    headers,
                    openai_sse_body(&text, &completion_id, &req.model),
                );
            }
            json_response_value(
                "200 OK",
                headers,
                json!({
                    "id": completion_id,
                    "object": "chat.completion",
                    "model": req.model,
                    "choices": [{
                        "index": 0,
                        "message": {"role": "assistant", "content": text},
                        "finish_reason": "stop"
                    }]
                }),
            )
        }
        Err(err) => json_error_response(
            "502 Bad Gateway",
            provider_headers(provider.id()),
            &err,
        ),
    }
}

fn handle_anthropic_messages(
    request: &ParsedRequest,
    registry: &Registry,
    provider_hint_from_path: Option<&str>,
) -> String {
    let payload = match parse_json_object_body(&request.body) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", vec![], &err),
    };
    let provider_hint = provider_from_payload(&payload)
        .or_else(|| provider_hint_from_path.map(ToString::to_string));
    let provider = match registry.resolve(provider_hint.as_deref()) {
        Ok(value) => value,
        Err(err) => return json_error_response("400 Bad Request", vec![], &err),
    };
    if !provider.capabilities().anthropic_compatible {
        return json_error_response(
            "400 Bad Request",
            provider_headers(provider.id()),
            "provider does not support /v1/messages",
        );
    }
    let req = build_anthropic_request(&payload, provider_hint.unwrap_or_default());
    match provider.generate_reply(&req) {
        Ok(text) => {
            let message_id = next_response_id("msg");
            let headers = provider_headers(provider.id());
            if req.stream {
                return event_stream_response(
                    "200 OK",
                    headers,
                    anthropic_sse_body(&text, &message_id),
                );
            }
            json_response_value(
                "200 OK",
                headers,
                json!({
                    "id": message_id,
                    "type": "message",
                    "role": "assistant",
                    "model": req.model,
                    "content": [{"type": "text", "text": text}],
                    "stop_reason": "end_turn"
                }),
            )
        }
        Err(err) => json_error_response(
            "502 Bad Gateway",
            provider_headers(provider.id()),
            &err,
        ),
    }
}

fn admin_cors_headers(request: &ParsedRequest) -> Vec<(String, String)> {
    let origin = request.header("origin").unwrap_or("*").to_string();
    let mut headers = vec![
        ("Access-Control-Allow-Origin".to_string(), origin.clone()),
        (
            "Access-Control-Allow-Headers".to_string(),
            "Content-Type, Authorization".to_string(),
        ),
        (
            "Access-Control-Allow-Methods".to_string(),
            "GET, POST, PUT, DELETE, OPTIONS".to_string(),
        ),
    ];
    if origin != "*" {
        headers.push((
            "Access-Control-Allow-Credentials".to_string(),
            "true".to_string(),
        ));
    }
    headers
}

fn path_only(path: &str) -> &str {
    path.split('?').next().unwrap_or(path)
}

fn admin_path_id(path: &str, prefix: &str) -> Option<String> {
    let raw = path.strip_prefix(prefix)?.trim();
    if raw.is_empty() || raw.contains('/') {
        return None;
    }
    Some(raw.to_string())
}

fn provider_from_path(path: &str) -> Option<&str> {
    path.split('?')
        .nth(1)
        .and_then(|query| query.split('&').find_map(|item| item.strip_prefix("provider=")))
}

fn admin_session_token(request: &ParsedRequest) -> Option<String> {
    if let Some(value) = request.header("authorization") {
        if let Some(token) = value.strip_prefix("Bearer ") {
            let trimmed = token.trim();
            if !trimmed.is_empty() {
                return Some(trimmed.to_string());
            }
        }
    }
    if let Some(cookie_header) = request.header("cookie") {
        for chunk in cookie_header.split(';') {
            if let Some((name, value)) = chunk.trim().split_once('=') {
                if name == ADMIN_SESSION_COOKIE && !value.trim().is_empty() {
                    return Some(value.trim().to_string());
                }
            }
        }
    }
    None
}

fn require_admin_token(request: &ParsedRequest) -> Result<(String, u64), String> {
    let token = admin_session_token(request).ok_or_else(|| {
        json_response(
            "401 Unauthorized",
            admin_cors_headers(request),
            "{\"error\":\"admin login required\"}".to_string(),
        )
    })?;
    let expires_at = admin_sessions().expires_at(&token).ok_or_else(|| {
        json_response(
            "401 Unauthorized",
            admin_cors_headers(request),
            "{\"error\":\"admin login required\"}".to_string(),
        )
    })?;
    Ok((token, expires_at))
}

fn shared_admin_features_json() -> &'static str {
    "{\"providers\":true,\"credentials\":true,\"providerState\":true,\"stats\":false,\"logs\":false,\"users\":false,\"configImportExport\":false}"
}

fn bool_count(value: bool) -> usize {
    if value {
        1
    } else {
        0
    }
}

fn provider_active_label(configured: bool) -> &'static str {
    if configured {
        "default"
    } else {
        ""
    }
}

fn active_kiro_id(accounts: &[KiroAccount]) -> String {
    accounts
        .iter()
        .find(|item| item.active)
        .map(|item| item.id.clone())
        .unwrap_or_default()
}

fn active_grok_id(tokens: &[GrokToken]) -> String {
    tokens
        .iter()
        .find(|item| item.active)
        .map(|item| item.id.clone())
        .unwrap_or_default()
}

fn admin_settings_json(settings: &AdminSettings) -> String {
    format!(
        "{{\"apiKey\":\"{}\",\"defaultProvider\":\"{}\",\"adminPasswordConfigured\":{}}}",
        json_escape(&settings.api_key),
        json_escape(&settings.default_provider),
        !settings.admin_password.trim().is_empty()
    )
}

fn cursor_config_json(config: &CursorRuntimeConfig) -> String {
    let fields = [
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
        format!("\"scriptUrl\":\"{}\"", json_escape(&config.script_url)),
        format!("\"cookie\":\"{}\"", json_escape(&config.cookie)),
        format!("\"xIsHuman\":\"{}\"", json_escape(&config.x_is_human)),
        format!("\"userAgent\":\"{}\"", json_escape(&config.user_agent)),
        format!("\"referer\":\"{}\"", json_escape(&config.referer)),
        format!("\"webglVendor\":\"{}\"", json_escape(&config.webgl_vendor)),
        format!(
            "\"webglRenderer\":\"{}\"",
            json_escape(&config.webgl_renderer)
        ),
    ];
    format!("{{{}}}", fields.join(","))
}

fn kiro_account_json(account: &KiroAccount) -> String {
    let fields = [
        format!("\"id\":\"{}\"", json_escape(&account.id)),
        format!("\"name\":\"{}\"", json_escape(&account.name)),
        format!(
            "\"accessToken\":\"{}\"",
            json_escape(&account.access_token)
        ),
        format!("\"machineId\":\"{}\"", json_escape(&account.machine_id)),
        format!(
            "\"preferredEndpoint\":\"{}\"",
            json_escape(&account.preferred_endpoint)
        ),
        format!("\"active\":{}", account.active),
    ];
    format!("{{{}}}", fields.join(","))
}

fn grok_token_json(token: &GrokToken) -> String {
    format!(
        "{{\"id\":\"{}\",\"name\":\"{}\",\"cookieToken\":\"{}\",\"active\":{}}}",
        json_escape(&token.id),
        json_escape(&token.name),
        json_escape(&token.cookie_token),
        token.active,
    )
}

fn grok_config_json(config: &GrokRuntimeConfig) -> String {
    let fields = [
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
        format!("\"proxyUrl\":\"{}\"", json_escape(&config.proxy_url)),
        format!("\"cfCookies\":\"{}\"", json_escape(&config.cf_cookies)),
        format!(
            "\"cfClearance\":\"{}\"",
            json_escape(&config.cf_clearance)
        ),
        format!("\"userAgent\":\"{}\"", json_escape(&config.user_agent)),
        format!("\"origin\":\"{}\"", json_escape(&config.origin)),
        format!("\"referer\":\"{}\"", json_escape(&config.referer)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn orchids_config_json(config: &OrchidsRuntimeConfig) -> String {
    let fields = [
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
        format!("\"clerkUrl\":\"{}\"", json_escape(&config.clerk_url)),
        format!(
            "\"clientCookie\":\"{}\"",
            json_escape(&config.client_cookie)
        ),
        format!("\"clientUat\":\"{}\"", json_escape(&config.client_uat)),
        format!("\"sessionId\":\"{}\"", json_escape(&config.session_id)),
        format!("\"projectId\":\"{}\"", json_escape(&config.project_id)),
        format!("\"userId\":\"{}\"", json_escape(&config.user_id)),
        format!("\"email\":\"{}\"", json_escape(&config.email)),
        format!("\"agentMode\":\"{}\"", json_escape(&config.agent_mode)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn web_config_json(config: &WebRuntimeConfig) -> String {
    let fields = [
        format!("\"baseUrl\":\"{}\"", json_escape(&config.base_url)),
        format!("\"type\":\"{}\"", json_escape(&config.type_name)),
        format!("\"apiKey\":\"{}\"", json_escape(&config.api_key)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn chatgpt_config_json(config: &ChatGPTRuntimeConfig) -> String {
    let fields = [
        format!("\"baseUrl\":\"{}\"", json_escape(&config.base_url)),
        format!("\"token\":\"{}\"", json_escape(&config.token)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn zai_image_config_json(config: &ZAIImageRuntimeConfig) -> String {
    let fields = [
        format!("\"sessionToken\":\"{}\"", json_escape(&config.session_token)),
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn zai_tts_config_json(config: &ZAITTSRuntimeConfig) -> String {
    let fields = [
        format!("\"token\":\"{}\"", json_escape(&config.token)),
        format!("\"userId\":\"{}\"", json_escape(&config.user_id)),
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn zai_ocr_config_json(config: &ZAIOCRRuntimeConfig) -> String {
    let fields = [
        format!("\"token\":\"{}\"", json_escape(&config.token)),
        format!("\"apiUrl\":\"{}\"", json_escape(&config.api_url)),
    ];
    format!("{{{}}}", fields.join(","))
}

fn admin_status_json(config: &AdminConfig) -> String {
    let cursor_configured = !config.providers.cursor_config.cookie.trim().is_empty();
    let orchids_configured = !config.providers.orchids_config.client_cookie.trim().is_empty();
    let web_configured = !config.providers.web_config.base_url.trim().is_empty()
        && !config.providers.web_config.type_name.trim().is_empty();
    let chatgpt_configured = !config.providers.chatgpt_config.base_url.trim().is_empty()
        && !config.providers.chatgpt_config.token.trim().is_empty();
    let zai_image_config = current_zai_image_config(config);
    let zai_tts_config = current_zai_tts_config(config);
    let zai_ocr_config = current_zai_ocr_config(config);
    let zai_image_configured = !zai_image_config.session_token.trim().is_empty();
    let zai_tts_configured = !zai_tts_config.token.trim().is_empty()
        && !zai_tts_config.user_id.trim().is_empty();
    let zai_ocr_configured = !zai_ocr_config.token.trim().is_empty();
    let kiro_count = config.providers.kiro_accounts.len();
    let grok_count = config.providers.grok_tokens.len();
    let providers = [
        format!(
            "\"cursor\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(cursor_configured),
            cursor_configured,
            provider_active_label(cursor_configured)
        ),
        format!(
            "\"kiro\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            kiro_count,
            kiro_count > 0,
            json_escape(&active_kiro_id(&config.providers.kiro_accounts))
        ),
        format!(
            "\"grok\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            grok_count,
            grok_count > 0,
            json_escape(&active_grok_id(&config.providers.grok_tokens))
        ),
        format!(
            "\"orchids\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(orchids_configured),
            orchids_configured,
            provider_active_label(orchids_configured)
        ),
        format!(
            "\"web\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(web_configured),
            web_configured,
            provider_active_label(web_configured)
        ),
        format!(
            "\"chatgpt\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(chatgpt_configured),
            chatgpt_configured,
            provider_active_label(chatgpt_configured)
        ),
        format!(
            "\"zaiImage\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(zai_image_configured),
            zai_image_configured,
            provider_active_label(zai_image_configured)
        ),
        format!(
            "\"zaiTTS\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(zai_tts_configured),
            zai_tts_configured,
            provider_active_label(zai_tts_configured)
        ),
        format!(
            "\"zaiOCR\":{{\"count\":{},\"configured\":{},\"active\":\"{}\"}}",
            bool_count(zai_ocr_configured),
            zai_ocr_configured,
            provider_active_label(zai_ocr_configured)
        ),
    ];
    format!(
        "{{\"project\":\"any2api-rust\",\"settings\":{},\"providers\":{{{}}}}}",
        admin_settings_json(&config.settings),
        providers.join(",")
    )
}

fn parse_settings_body(body: &str) -> Result<(String, String, String), ()> {
    if !body.trim_start().starts_with('{') {
        return Err(());
    }
    Ok((
        json_string_field(body, "apiKey").unwrap_or_default(),
        json_string_field(body, "defaultProvider").unwrap_or_default(),
        json_string_field(body, "adminPassword").unwrap_or_default(),
    ))
}

fn parse_kiro_account(body: &str) -> KiroAccount {
    KiroAccount {
        id: json_string_field(body, "id").unwrap_or_default(),
        name: json_string_field(body, "name").unwrap_or_default(),
        access_token: json_string_field(body, "accessToken").unwrap_or_default(),
        machine_id: json_string_field(body, "machineId").unwrap_or_default(),
        preferred_endpoint: json_string_field(body, "preferredEndpoint")
            .unwrap_or_default()
            .to_lowercase(),
        active: json_bool_field(body, "active").unwrap_or(false),
    }
}

fn parse_grok_token(body: &str) -> GrokToken {
    GrokToken {
        id: json_string_field(body, "id").unwrap_or_default(),
        name: json_string_field(body, "name").unwrap_or_default(),
        cookie_token: json_string_field(body, "cookieToken").unwrap_or_default(),
        active: json_bool_field(body, "active").unwrap_or(false),
    }
}

fn parse_kiro_account_body(body: &str) -> Result<KiroAccount, ()> {
    let value = parse_json_object_body(body).map_err(|_| ())?;
    Ok(parse_kiro_account(&value.to_string()))
}

fn parse_grok_token_body(body: &str) -> Result<GrokToken, ()> {
    let value = parse_json_object_body(body).map_err(|_| ())?;
    Ok(parse_grok_token(&value.to_string()))
}

fn parse_grok_config_body(body: &str) -> Result<GrokRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(GrokRuntimeConfig::from_json(&config))
}

fn parse_cursor_config_body(body: &str) -> Result<CursorRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(CursorRuntimeConfig::from_json(&config))
}

fn parse_orchids_config_body(body: &str) -> Result<OrchidsRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(OrchidsRuntimeConfig::from_json(&config))
}

fn parse_web_config_body(body: &str) -> Result<WebRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(WebRuntimeConfig::from_json(&config))
}

fn parse_chatgpt_config_body(body: &str) -> Result<ChatGPTRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(ChatGPTRuntimeConfig::from_json(&config))
}

fn parse_zai_image_config_body(body: &str) -> Result<ZAIImageRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(ZAIImageRuntimeConfig::from_json(&config))
}

fn parse_zai_tts_config_body(body: &str) -> Result<ZAITTSRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(ZAITTSRuntimeConfig::from_json(&config))
}

fn parse_zai_ocr_config_body(body: &str) -> Result<ZAIOCRRuntimeConfig, ()> {
    let config = json_object_field(body, "config").ok_or(())?;
    Ok(ZAIOCRRuntimeConfig::from_json(&config))
}

fn parse_kiro_accounts_body(body: &str) -> Result<Vec<KiroAccount>, ()> {
    let accounts = json_array_field(body, "accounts").ok_or(())?;
    Ok(json_array_objects(&accounts)
        .into_iter()
        .map(|item| parse_kiro_account(&item))
        .collect())
}

fn parse_grok_tokens_body(body: &str) -> Result<Vec<GrokToken>, ()> {
    let tokens = json_array_field(body, "tokens").ok_or(())?;
    Ok(json_array_objects(&tokens)
        .into_iter()
        .map(|item| parse_grok_token(&item))
        .collect())
}

fn admin_kiro_accounts_json(accounts: &[KiroAccount]) -> String {
    let body = accounts
        .iter()
        .map(kiro_account_json)
        .collect::<Vec<_>>()
        .join(",");
    format!("{{\"accounts\":[{body}]}}")
}

fn admin_grok_tokens_json(tokens: &[GrokToken]) -> String {
    let body = tokens
        .iter()
        .map(grok_token_json)
        .collect::<Vec<_>>()
        .join(",");
    format!("{{\"tokens\":[{body}]}}")
}

fn admin_store_error_response(request: &ParsedRequest, err: &str) -> String {
    let status = if err.starts_with("invalid ") {
        "400 Bad Request"
    } else {
        "500 Internal Server Error"
    };
    json_error_response(status, admin_cors_headers(request), err)
}

fn handle_admin_api(request: &ParsedRequest, method: &str, clean_path: &str) -> Option<String> {
    if !clean_path.starts_with("/admin/api/") {
        return None;
    }
    if let Err(response) = require_admin_token(request) {
        return Some(response);
    }
    let store = admin_store();
    let snapshot = store.snapshot();
    if method == "GET"
        && (clean_path == "/admin/api/providers/kiro/accounts"
            || clean_path == "/admin/api/providers/kiro/accounts/list")
    {
        return Some(json_response(
            "200 OK",
            admin_cors_headers(request),
            admin_kiro_accounts_json(&snapshot.providers.kiro_accounts),
        ));
    }
    if method == "GET" && clean_path.starts_with("/admin/api/providers/kiro/accounts/detail/") {
        let Some(account_id) = admin_path_id(clean_path, "/admin/api/providers/kiro/accounts/detail/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match store.kiro_account(&account_id) {
            Some(account) => json_response(
                "200 OK",
                admin_cors_headers(request),
                format!("{{\"account\":{}}}", kiro_account_json(&account)),
            ),
            None => json_error_response("404 Not Found", admin_cors_headers(request), "not found"),
        });
    }
    if method == "POST" && clean_path == "/admin/api/providers/kiro/accounts/create" {
        return Some(match parse_kiro_account_body(&request.body) {
            Ok(account) => match store.create_kiro_account(account) {
                Ok(account) => json_response(
                    "200 OK",
                    admin_cors_headers(request),
                    format!("{{\"account\":{}}}", kiro_account_json(&account)),
                ),
                Err(err) => admin_store_error_response(request, &err),
            },
            Err(()) => json_error_response(
                "400 Bad Request",
                admin_cors_headers(request),
                "invalid json",
            ),
        });
    }
    if method == "PUT" && clean_path.starts_with("/admin/api/providers/kiro/accounts/update/") {
        let Some(account_id) = admin_path_id(clean_path, "/admin/api/providers/kiro/accounts/update/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match parse_kiro_account_body(&request.body) {
            Ok(account) => match store.update_kiro_account(&account_id, account) {
                Ok(Some(account)) => json_response(
                    "200 OK",
                    admin_cors_headers(request),
                    format!("{{\"account\":{}}}", kiro_account_json(&account)),
                ),
                Ok(None) => json_error_response(
                    "404 Not Found",
                    admin_cors_headers(request),
                    "not found",
                ),
                Err(err) => admin_store_error_response(request, &err),
            },
            Err(()) => json_error_response(
                "400 Bad Request",
                admin_cors_headers(request),
                "invalid json",
            ),
        });
    }
    if method == "DELETE" && clean_path.starts_with("/admin/api/providers/kiro/accounts/delete/") {
        let Some(account_id) = admin_path_id(clean_path, "/admin/api/providers/kiro/accounts/delete/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match store.delete_kiro_account(&account_id) {
            Ok(true) => json_response(
                "200 OK",
                admin_cors_headers(request),
                "{\"ok\":true}".to_string(),
            ),
            Ok(false) => json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ),
            Err(err) => admin_store_error_response(request, &err),
        });
    }
    if method == "GET"
        && (clean_path == "/admin/api/providers/grok/tokens"
            || clean_path == "/admin/api/providers/grok/tokens/list")
    {
        return Some(json_response(
            "200 OK",
            admin_cors_headers(request),
            admin_grok_tokens_json(&snapshot.providers.grok_tokens),
        ));
    }
    if method == "GET" && clean_path.starts_with("/admin/api/providers/grok/tokens/detail/") {
        let Some(token_id) = admin_path_id(clean_path, "/admin/api/providers/grok/tokens/detail/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match store.grok_token(&token_id) {
            Some(token) => json_response(
                "200 OK",
                admin_cors_headers(request),
                format!("{{\"token\":{}}}", grok_token_json(&token)),
            ),
            None => json_error_response("404 Not Found", admin_cors_headers(request), "not found"),
        });
    }
    if method == "POST" && clean_path == "/admin/api/providers/grok/tokens/create" {
        return Some(match parse_grok_token_body(&request.body) {
            Ok(token) => match store.create_grok_token(token) {
                Ok(token) => json_response(
                    "200 OK",
                    admin_cors_headers(request),
                    format!("{{\"token\":{}}}", grok_token_json(&token)),
                ),
                Err(err) => admin_store_error_response(request, &err),
            },
            Err(()) => json_error_response(
                "400 Bad Request",
                admin_cors_headers(request),
                "invalid json",
            ),
        });
    }
    if method == "PUT" && clean_path.starts_with("/admin/api/providers/grok/tokens/update/") {
        let Some(token_id) = admin_path_id(clean_path, "/admin/api/providers/grok/tokens/update/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match parse_grok_token_body(&request.body) {
            Ok(token) => match store.update_grok_token(&token_id, token) {
                Ok(Some(token)) => json_response(
                    "200 OK",
                    admin_cors_headers(request),
                    format!("{{\"token\":{}}}", grok_token_json(&token)),
                ),
                Ok(None) => json_error_response(
                    "404 Not Found",
                    admin_cors_headers(request),
                    "not found",
                ),
                Err(err) => admin_store_error_response(request, &err),
            },
            Err(()) => json_error_response(
                "400 Bad Request",
                admin_cors_headers(request),
                "invalid json",
            ),
        });
    }
    if method == "DELETE" && clean_path.starts_with("/admin/api/providers/grok/tokens/delete/") {
        let Some(token_id) = admin_path_id(clean_path, "/admin/api/providers/grok/tokens/delete/") else {
            return Some(json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ));
        };
        return Some(match store.delete_grok_token(&token_id) {
            Ok(true) => json_response(
                "200 OK",
                admin_cors_headers(request),
                "{\"ok\":true}".to_string(),
            ),
            Ok(false) => json_error_response(
                "404 Not Found",
                admin_cors_headers(request),
                "not found",
            ),
            Err(err) => admin_store_error_response(request, &err),
        });
    }
    let response = match (method, clean_path) {
        ("GET", "/admin/api/status") => json_response(
            "200 OK",
            admin_cors_headers(request),
            admin_status_json(&snapshot),
        ),
        ("GET", "/admin/api/settings") => json_response(
            "200 OK",
            admin_cors_headers(request),
            admin_settings_json(&snapshot.settings),
        ),
        ("PUT", "/admin/api/settings") => match parse_settings_body(&request.body) {
            Ok((api_key, default_provider, admin_password)) => match store
                .update_settings(&api_key, &default_provider, &admin_password)
            {
                Ok(config) => json_response(
                    "200 OK",
                    admin_cors_headers(request),
                    admin_settings_json(&config.settings),
                ),
                Err(err) => json_response(
                    "500 Internal Server Error",
                    admin_cors_headers(request),
                    format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                ),
            },
            Err(()) => json_response(
                "400 Bad Request",
                admin_cors_headers(request),
                "{\"error\":\"invalid json\"}".to_string(),
            ),
        },
        ("GET", "/admin/api/providers/cursor/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                cursor_config_json(&snapshot.providers.cursor_config)
            ),
        ),
        ("PUT", "/admin/api/providers/cursor/config") => {
            match parse_cursor_config_body(&request.body) {
                Ok(config) => match store.replace_cursor_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            cursor_config_json(&config.providers.cursor_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("PUT", "/admin/api/providers/kiro/accounts") => {
            match parse_kiro_accounts_body(&request.body) {
                Ok(accounts) => match store.replace_kiro_accounts(accounts) {
                    Ok(config) => {
                        json_response(
                            "200 OK",
                            admin_cors_headers(request),
                            admin_kiro_accounts_json(&config.providers.kiro_accounts),
                        )
                    }
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("PUT", "/admin/api/providers/grok/tokens") => {
            match parse_grok_tokens_body(&request.body) {
                Ok(tokens) => match store.replace_grok_tokens(tokens) {
                    Ok(config) => {
                        json_response(
                            "200 OK",
                            admin_cors_headers(request),
                            admin_grok_tokens_json(&config.providers.grok_tokens),
                        )
                    }
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/grok/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                grok_config_json(&snapshot.providers.grok_config)
            ),
        ),
        ("PUT", "/admin/api/providers/grok/config") => {
            match parse_grok_config_body(&request.body) {
                Ok(config) => match store.replace_grok_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            grok_config_json(&config.providers.grok_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/orchids/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                orchids_config_json(&snapshot.providers.orchids_config)
            ),
        ),
        ("PUT", "/admin/api/providers/orchids/config") => {
            match parse_orchids_config_body(&request.body) {
                Ok(config) => match store.replace_orchids_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            orchids_config_json(&config.providers.orchids_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/web/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                web_config_json(&snapshot.providers.web_config)
            ),
        ),
        ("PUT", "/admin/api/providers/web/config") => {
            match parse_web_config_body(&request.body) {
                Ok(config) => match store.replace_web_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            web_config_json(&config.providers.web_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/chatgpt/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                chatgpt_config_json(&snapshot.providers.chatgpt_config)
            ),
        ),
        ("PUT", "/admin/api/providers/chatgpt/config") => {
            match parse_chatgpt_config_body(&request.body) {
                Ok(config) => match store.replace_chatgpt_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            chatgpt_config_json(&config.providers.chatgpt_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/zai/image/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                zai_image_config_json(&snapshot.providers.zai_image_config)
            ),
        ),
        ("PUT", "/admin/api/providers/zai/image/config") => {
            match parse_zai_image_config_body(&request.body) {
                Ok(config) => match store.replace_zai_image_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            zai_image_config_json(&config.providers.zai_image_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/zai/tts/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                zai_tts_config_json(&snapshot.providers.zai_tts_config)
            ),
        ),
        ("PUT", "/admin/api/providers/zai/tts/config") => {
            match parse_zai_tts_config_body(&request.body) {
                Ok(config) => match store.replace_zai_tts_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            zai_tts_config_json(&config.providers.zai_tts_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        ("GET", "/admin/api/providers/zai/ocr/config") => json_response(
            "200 OK",
            admin_cors_headers(request),
            format!(
                "{{\"config\":{}}}",
                zai_ocr_config_json(&snapshot.providers.zai_ocr_config)
            ),
        ),
        ("PUT", "/admin/api/providers/zai/ocr/config") => {
            match parse_zai_ocr_config_body(&request.body) {
                Ok(config) => match store.replace_zai_ocr_config(config) {
                    Ok(config) => json_response(
                        "200 OK",
                        admin_cors_headers(request),
                        format!(
                            "{{\"config\":{}}}",
                            zai_ocr_config_json(&config.providers.zai_ocr_config)
                        ),
                    ),
                    Err(err) => json_response(
                        "500 Internal Server Error",
                        admin_cors_headers(request),
                        format!("{{\"error\":\"{}\"}}", json_escape(&err)),
                    ),
                },
                Err(()) => json_response(
                    "400 Bad Request",
                    admin_cors_headers(request),
                    "{\"error\":\"invalid json\"}".to_string(),
                ),
            }
        },
        _ => json_response(
            "404 Not Found",
            admin_cors_headers(request),
            "{\"error\":\"not found\"}".to_string(),
        ),
    };
    Some(response)
}

fn handle_request(request: ParsedRequest) -> Vec<u8> {
    admin_store().sync_from_env();

    let method = request.method.as_str();
    let path = request.path.as_str();
    let clean_path = path_only(path);

    if (clean_path.starts_with("/api/admin/") || clean_path.starts_with("/admin/api/"))
        && method == "OPTIONS"
    {
        return response(
            "204 No Content",
            "text/plain; charset=utf-8",
            admin_cors_headers(&request),
            String::new(),
        )
        .into_bytes();
    }

    if method == "GET" && clean_path == "/api/admin/meta" {
        return json_response(
            "200 OK",
            admin_cors_headers(&request),
            format!(
                "{{\"backend\":{{\"language\":\"rust\",\"version\":\"{ADMIN_BACKEND_VERSION}\"}},\"auth\":{{\"mode\":\"{ADMIN_AUTH_MODE}\"}},\"features\":{}}}",
                shared_admin_features_json()
            ),
        )
        .into_bytes();
    }

    if method == "POST" && clean_path == "/api/admin/auth/login" {
        let Some(password) = json_string_field(&request.body, "password") else {
            return json_response(
                "400 Bad Request",
                admin_cors_headers(&request),
                "{\"error\":\"invalid json\"}".to_string(),
            )
            .into_bytes();
        };
        if password.trim() != admin_store().admin_password() {
            return json_response(
                "401 Unauthorized",
                admin_cors_headers(&request),
                "{\"error\":\"invalid admin password\"}".to_string(),
            )
            .into_bytes();
        }
        let (token, _) = admin_sessions().create();
        let mut headers = admin_cors_headers(&request);
        headers.push((
            "Set-Cookie".to_string(),
            format!(
                "{ADMIN_SESSION_COOKIE}={token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400"
            ),
        ));
        return json_response(
            "200 OK",
            headers,
            format!("{{\"ok\":true,\"token\":\"{}\"}}", json_escape(&token)),
        )
        .into_bytes();
    }

    if method == "GET" && clean_path == "/api/admin/auth/session" {
        let (_, expires_at) = match require_admin_token(&request) {
            Ok(value) => value,
            Err(response) => return response.into_bytes(),
        };
        return json_response(
            "200 OK",
            admin_cors_headers(&request),
            format!(
                "{{\"authenticated\":true,\"user\":{{\"id\":\"local-admin\",\"name\":\"Admin\",\"role\":\"admin\"}},\"expiresAt\":\"{}\"}}",
                expires_at
            ),
        )
        .into_bytes();
    }

    if method == "POST" && clean_path == "/api/admin/auth/logout" {
        let (token, _) = match require_admin_token(&request) {
            Ok(value) => value,
            Err(response) => return response.into_bytes(),
        };
        admin_sessions().delete(&token);
        let mut headers = admin_cors_headers(&request);
        headers.push((
            "Set-Cookie".to_string(),
            format!(
                "{ADMIN_SESSION_COOKIE}=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0"
            ),
        ));
        return json_response("200 OK", headers, "{\"ok\":true}".to_string()).into_bytes();
    }

    if let Some(response) = handle_admin_api(&request, method, clean_path) {
        return response.into_bytes();
    }

    let snapshot = admin_store().snapshot();
    let registry = default_registry(&snapshot.settings.default_provider, &snapshot);
    let provider_key = provider_from_path(path);

    if method == "GET" && path.starts_with("/health") {
        return json_response(
            "200 OK",
            vec![],
            "{\"status\":\"ok\",\"project\":\"any2api-rust\"}".to_string(),
        )
        .into_bytes();
    }

    if method == "GET" && path.starts_with("/v1/models") {
        return match registry.models(provider_key) {
            Ok(models) => {
                let data = models.iter().map(|item| format!(
                    "{{\"id\":\"{}\",\"object\":\"model\",\"owned_by\":\"{}\",\"provider\":\"{}\",\"upstream_model\":\"{}\"}}",
                    item.public_model, item.owned_by, item.provider, item.upstream_model
                )).collect::<Vec<_>>().join(",");
                json_response("200 OK", vec![], format!("{{\"object\":\"list\",\"data\":[{}]}}", data))
            }
            Err(err) => json_response(
                "400 Bad Request",
                vec![],
                format!("{{\"error\":\"{}\"}}", json_escape(&err)),
            ),
        }
        .into_bytes();
    }

    if method == "POST" && clean_path == "/v1/images/generations" {
        return handle_openai_images_generation(&request, &snapshot).into_bytes();
    }

    if method == "POST" && clean_path == "/v1/audio/speech" {
        return handle_openai_audio_speech_bytes(&request, &snapshot);
    }

    if method == "POST" && clean_path == "/v1/ocr" {
        return handle_ocr_upload(&request, &snapshot).into_bytes();
    }

    if method == "POST" && path.starts_with("/v1/chat/completions") {
        return handle_openai_chat(&request, &registry, provider_key).into_bytes();
    }

    if method == "POST" && path.starts_with("/v1/messages") {
        return handle_anthropic_messages(&request, &registry, provider_key).into_bytes();
    }

    json_response("404 Not Found", vec![], "{\"error\":\"not found\"}".to_string())
        .into_bytes()
}

fn handle_stream(mut stream: TcpStream, raw_request: &[u8]) {
    let payload = handle_request(parse_request(raw_request));
    let _ = stream.write_all(&payload);
}

pub fn run(addr: &str) -> std::io::Result<()> {
    let listener = TcpListener::bind(addr)?;
    println!("any2api-rust listening on http://{addr}");
    for incoming in listener.incoming() {
        let mut stream = incoming?;
        let mut buffer = [0_u8; 8192];
        let size = stream.read(&mut buffer)?;
        handle_stream(stream, &buffer[..size]);
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::env;
    use std::io::{Read, Write};
    use std::fs;
    use std::net::TcpListener;
    use std::path::PathBuf;
    use std::thread;

    static TEST_LOCK: OnceLock<Mutex<()>> = OnceLock::new();
    static TEST_COUNTER: AtomicU64 = AtomicU64::new(1);

    fn build_request(
        method: &str,
        path: &str,
        headers: &[(&str, &str)],
        body: &str,
    ) -> ParsedRequest {
        let mut parsed_headers = HashMap::new();
        for (name, value) in headers {
            parsed_headers.insert(name.to_ascii_lowercase(), (*value).to_string());
        }
        ParsedRequest {
            method: method.to_string(),
            path: path.to_string(),
            headers: parsed_headers,
            body: body.to_string(),
            body_bytes: body.as_bytes().to_vec(),
        }
    }

    fn build_bytes_request(
        method: &str,
        path: &str,
        headers: &[(&str, &str)],
        body: Vec<u8>,
    ) -> ParsedRequest {
        let mut parsed_headers = HashMap::new();
        for (name, value) in headers {
            parsed_headers.insert(name.to_ascii_lowercase(), (*value).to_string());
        }
        ParsedRequest {
            method: method.to_string(),
            path: path.to_string(),
            headers: parsed_headers,
            body: String::from_utf8_lossy(&body).to_string(),
            body_bytes: body,
        }
    }

    fn handle_request(request: ParsedRequest) -> String {
        String::from_utf8(super::handle_request(request)).expect("utf8 http response")
    }

    fn handle_request_bytes(request: ParsedRequest) -> Vec<u8> {
        super::handle_request(request)
    }

    fn response_token(response: &str) -> Option<String> {
        let marker = "\"token\":\"";
        let start = response.find(marker)? + marker.len();
        let rest = &response[start..];
        Some(rest.split('"').next()?.to_string())
    }

    fn response_json_string_field(response: &str, field: &str) -> Option<String> {
        let marker = format!("\"{field}\":\"");
        let start = response.find(&marker)? + marker.len();
        let rest = &response[start..];
        Some(rest.split('"').next()?.to_string())
    }

    fn split_http_response(response: &[u8]) -> (String, Vec<u8>) {
        let split_index = response
            .windows(4)
            .position(|window| window == b"\r\n\r\n")
            .expect("http response head/body separator");
        (
            String::from_utf8_lossy(&response[..split_index]).to_string(),
            response[split_index + 4..].to_vec(),
        )
    }

    fn test_lock() -> &'static Mutex<()> {
        TEST_LOCK.get_or_init(|| Mutex::new(()))
    }

    struct EnvVarGuard {
        key: &'static str,
        previous: Option<String>,
    }

    impl EnvVarGuard {
        fn set(key: &'static str, value: impl Into<String>) -> Self {
            let previous = env::var(key).ok();
            env::set_var(key, value.into());
            Self { key, previous }
        }
    }

    impl Drop for EnvVarGuard {
        fn drop(&mut self) {
            match &self.previous {
                Some(value) => env::set_var(self.key, value),
                None => env::remove_var(self.key),
            }
        }
    }

    struct MockResponse {
        status: &'static str,
        content_type: &'static str,
        headers: Vec<(String, String)>,
        body: Vec<u8>,
    }

    fn text_mock_response(status: &'static str, content_type: &'static str, body: &str) -> MockResponse {
        MockResponse {
            status,
            content_type,
            headers: vec![],
            body: body.as_bytes().to_vec(),
        }
    }

    fn spawn_mock_server<F>(expected_requests: usize, handler: F) -> String
    where
        F: Fn(&ParsedRequest) -> MockResponse + Send + 'static,
    {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind mock server");
        let addr = listener.local_addr().expect("mock server local addr");
        thread::spawn(move || {
            for _ in 0..expected_requests {
                let (mut stream, _) = listener.accept().expect("accept mock request");
                let mut buffer = [0_u8; 16 * 1024];
                let size = stream.read(&mut buffer).expect("read mock request");
                let request = parse_request(&buffer[..size]);
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
                let raw_head = format!("{}\r\n\r\n", head.join("\r\n"));
                stream
                    .write_all(raw_head.as_bytes())
                    .expect("write mock response head");
                stream
                    .write_all(&response.body)
                    .expect("write mock response body");
            }
        });
        format!("http://{addr}")
    }

    fn kiro_frame(event_type: &str, payload: &str) -> Vec<u8> {
        let mut headers = Vec::new();
        headers.push(11);
        headers.extend_from_slice(b":event-type");
        headers.push(7);
        headers.extend_from_slice(&(event_type.len() as u16).to_be_bytes());
        headers.extend_from_slice(event_type.as_bytes());
        let payload_bytes = payload.as_bytes();
        let total_length = 12 + headers.len() + payload_bytes.len() + 4;
        let mut frame = Vec::new();
        frame.extend_from_slice(&(total_length as u32).to_be_bytes());
        frame.extend_from_slice(&(headers.len() as u32).to_be_bytes());
        frame.extend_from_slice(&0_u32.to_be_bytes());
        frame.extend_from_slice(&headers);
        frame.extend_from_slice(payload_bytes);
        frame.extend_from_slice(&0_u32.to_be_bytes());
        frame
    }

    struct TestEnvGuard {
        previous_password: Option<String>,
        previous_provider: Option<String>,
        previous_store_path: Option<String>,
        previous_api_key: Option<String>,
        temp_dir: PathBuf,
    }

    impl TestEnvGuard {
        fn new(name: &str) -> Self {
            let temp_dir = env::temp_dir().join(format!(
                "any2api-rust-{name}-{}",
                TEST_COUNTER.fetch_add(1, Ordering::Relaxed)
            ));
            fs::create_dir_all(&temp_dir).expect("create rust test temp dir");
            let guard = Self {
                previous_password: env::var("NEWPLATFORM2API_ADMIN_PASSWORD").ok(),
                previous_provider: env::var("NEWPLATFORM2API_DEFAULT_PROVIDER").ok(),
                previous_store_path: env::var("NEWPLATFORM2API_ADMIN_STORE_PATH").ok(),
                previous_api_key: env::var("NEWPLATFORM2API_API_KEY").ok(),
                temp_dir,
            };
            env::set_var("NEWPLATFORM2API_ADMIN_PASSWORD", "changeme");
            env::set_var("NEWPLATFORM2API_DEFAULT_PROVIDER", "cursor");
            env::set_var(
                "NEWPLATFORM2API_ADMIN_STORE_PATH",
                guard.store_path().to_string_lossy().to_string(),
            );
            env::set_var("NEWPLATFORM2API_API_KEY", "");
            admin_sessions().clear();
            admin_store().sync_from_env();
            guard
        }

        fn store_path(&self) -> PathBuf {
            self.temp_dir.join("admin.json")
        }

        fn reload(&self) {
            admin_sessions().clear();
            admin_store().sync_from_env();
        }
    }

    impl Drop for TestEnvGuard {
        fn drop(&mut self) {
            admin_sessions().clear();
            match &self.previous_password {
                Some(value) => env::set_var("NEWPLATFORM2API_ADMIN_PASSWORD", value),
                None => env::remove_var("NEWPLATFORM2API_ADMIN_PASSWORD"),
            }
            match &self.previous_provider {
                Some(value) => env::set_var("NEWPLATFORM2API_DEFAULT_PROVIDER", value),
                None => env::remove_var("NEWPLATFORM2API_DEFAULT_PROVIDER"),
            }
            match &self.previous_store_path {
                Some(value) => env::set_var("NEWPLATFORM2API_ADMIN_STORE_PATH", value),
                None => env::remove_var("NEWPLATFORM2API_ADMIN_STORE_PATH"),
            }
            match &self.previous_api_key {
                Some(value) => env::set_var("NEWPLATFORM2API_API_KEY", value),
                None => env::remove_var("NEWPLATFORM2API_API_KEY"),
            }
            let _ = fs::remove_dir_all(&self.temp_dir);
        }
    }

    #[test]
    fn admin_meta_endpoint() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("meta");

        let response = handle_request(build_request(
            "GET",
            "/api/admin/meta",
            &[("Origin", "http://localhost:1420")],
            "",
        ));
        assert!(response.starts_with("HTTP/1.1 200 OK"));
        assert!(response.contains("\"language\":\"rust\""));
        assert!(response.contains("\"mode\":\"session_cookie\""));
        assert!(response.contains("\"providers\":true"));
        assert!(response.contains("\"credentials\":true"));
        assert!(response.contains("\"providerState\":true"));
        assert!(response.contains("Access-Control-Allow-Origin: http://localhost:1420"));
    }

    #[test]
    fn admin_login_session_logout_lifecycle() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("auth");

        let bad_login = handle_request(build_request(
            "POST",
            "/api/admin/auth/login",
            &[],
            "{\"password\":\"wrong\"}",
        ));
        assert!(bad_login.starts_with("HTTP/1.1 401 Unauthorized"));

        let login = handle_request(build_request(
            "POST",
            "/api/admin/auth/login",
            &[],
            "{\"password\":\"changeme\"}",
        ));
        assert!(login.starts_with("HTTP/1.1 200 OK"));
        assert!(login.contains("Set-Cookie: newplatform2api_admin_session="));
        let token = response_token(&login).expect("missing token in login response");

        let session = handle_request(build_request(
            "GET",
            "/api/admin/auth/session",
            &[("Authorization", &format!("Bearer {token}"))],
            "",
        ));
        assert!(session.starts_with("HTTP/1.1 200 OK"));
        assert!(session.contains("\"authenticated\":true"));
        assert!(session.contains("\"role\":\"admin\""));

        let logout = handle_request(build_request(
            "POST",
            "/api/admin/auth/logout",
            &[("Authorization", &format!("Bearer {token}"))],
            "",
        ));
        assert!(logout.starts_with("HTTP/1.1 200 OK"));
        assert!(logout.contains("\"ok\":true"));

        let session_after = handle_request(build_request(
            "GET",
            "/api/admin/auth/session",
            &[("Authorization", &format!("Bearer {token}"))],
            "",
        ));
        assert!(session_after.starts_with("HTTP/1.1 401 Unauthorized"));
        assert!(session_after.contains("admin login required"));
    }

    #[test]
    fn admin_api_requires_auth_and_supports_cors() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("cors");

        let unauth = handle_request(build_request("GET", "/admin/api/settings", &[], ""));
        assert!(unauth.starts_with("HTTP/1.1 401 Unauthorized"));
        assert!(unauth.contains("admin login required"));

        let options = handle_request(build_request(
            "OPTIONS",
            "/admin/api/settings",
            &[("Origin", "http://localhost:1420")],
            "",
        ));
        assert!(options.starts_with("HTTP/1.1 204 No Content"));
        assert!(options.contains("Access-Control-Allow-Origin: http://localhost:1420"));
        assert!(options.contains("Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS"));
    }

    #[test]
    fn settings_and_provider_roundtrip_persist_across_reload() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let env_guard = TestEnvGuard::new("roundtrip");

        let login = handle_request(build_request(
            "POST",
            "/api/admin/auth/login",
            &[],
            "{\"password\":\"changeme\"}",
        ));
        let token = response_token(&login).expect("missing token in login response");
        let auth_header = format!("Bearer {token}");

        let settings = handle_request(build_request(
            "GET",
            "/admin/api/settings",
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(settings.starts_with("HTTP/1.1 200 OK"));
        assert!(settings.contains("\"defaultProvider\":\"cursor\""));
        assert!(settings.contains("\"adminPasswordConfigured\":true"));

        let updated_settings = handle_request(build_request(
            "PUT",
            "/admin/api/settings",
            &[("Authorization", auth_header.as_str())],
            "{\"apiKey\":\"sk-rust\",\"defaultProvider\":\"grok\",\"adminPassword\":\"newpass\"}",
        ));
        assert!(updated_settings.starts_with("HTTP/1.1 200 OK"));
        assert!(updated_settings.contains("\"apiKey\":\"sk-rust\""));
        assert!(updated_settings.contains("\"defaultProvider\":\"grok\""));

        let cursor = handle_request(build_request(
            "PUT",
            "/admin/api/providers/cursor/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"apiUrl\":\"https://cursor.test/chat\",\"cookie\":\"cursor-cookie\"}}",
        ));
        assert!(cursor.starts_with("HTTP/1.1 200 OK"));
        assert!(cursor.contains("\"apiUrl\":\"https://cursor.test/chat\""));
        assert!(cursor.contains("\"cookie\":\"cursor-cookie\""));

        let kiro_main = handle_request(build_request(
            "POST",
            "/admin/api/providers/kiro/accounts/create",
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Main\",\"accessToken\":\"ak-1\",\"machineId\":\"machine-1\",\"active\":true}",
        ));
        assert!(kiro_main.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_main.contains("\"name\":\"Main\""));
        assert!(kiro_main.contains("\"active\":true"));
        let kiro_main_id = response_json_string_field(&kiro_main, "id").expect("missing kiro main id");

        let kiro_backup = handle_request(build_request(
            "POST",
            "/admin/api/providers/kiro/accounts/create",
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Backup\",\"accessToken\":\"ak-2\",\"machineId\":\"machine-2\",\"preferredEndpoint\":\"codewhisperer\",\"active\":false}",
        ));
        assert!(kiro_backup.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_backup.contains("\"name\":\"Backup\""));
        assert!(kiro_backup.contains("\"active\":false"));
        let kiro_backup_id = response_json_string_field(&kiro_backup, "id").expect("missing kiro backup id");

        let kiro_detail = handle_request(build_request(
            "GET",
            &format!("/admin/api/providers/kiro/accounts/detail/{kiro_main_id}"),
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(kiro_detail.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_detail.contains("\"name\":\"Main\""));
        assert!(kiro_detail.contains("\"machineId\":\"machine-1\""));

        let kiro_update = handle_request(build_request(
            "PUT",
            &format!("/admin/api/providers/kiro/accounts/update/{kiro_backup_id}"),
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Backup\",\"accessToken\":\"ak-2\",\"machineId\":\"machine-2\",\"preferredEndpoint\":\"amazonq\",\"active\":true}",
        ));
        assert!(kiro_update.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_update.contains("\"preferredEndpoint\":\"amazonq\""));
        assert!(kiro_update.contains("\"active\":true"));

        let kiro_list = handle_request(build_request(
            "GET",
            "/admin/api/providers/kiro/accounts/list",
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(kiro_list.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_list.contains(&format!("\"id\":\"{kiro_main_id}\"")));
        assert!(kiro_list.contains(&format!("\"id\":\"{kiro_backup_id}\"")));
        assert!(kiro_list.contains("\"active\":false"));
        assert!(kiro_list.contains("\"active\":true"));

        let kiro_delete = handle_request(build_request(
            "DELETE",
            &format!("/admin/api/providers/kiro/accounts/delete/{kiro_main_id}"),
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(kiro_delete.starts_with("HTTP/1.1 200 OK"));
        assert!(kiro_delete.contains("\"ok\":true"));

        let grok_primary = handle_request(build_request(
            "POST",
            "/admin/api/providers/grok/tokens/create",
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Primary\",\"cookieToken\":\"gt-1\",\"active\":true}",
        ));
        assert!(grok_primary.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_primary.contains("\"name\":\"Primary\""));
        let grok_primary_id = response_json_string_field(&grok_primary, "id").expect("missing grok primary id");

        let grok_secondary = handle_request(build_request(
            "POST",
            "/admin/api/providers/grok/tokens/create",
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Secondary\",\"cookieToken\":\"gt-2\",\"active\":false}",
        ));
        assert!(grok_secondary.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_secondary.contains("\"name\":\"Secondary\""));
        assert!(grok_secondary.contains("\"active\":false"));
        let grok_secondary_id = response_json_string_field(&grok_secondary, "id").expect("missing grok secondary id");

        let grok_detail = handle_request(build_request(
            "GET",
            &format!("/admin/api/providers/grok/tokens/detail/{grok_primary_id}"),
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(grok_detail.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_detail.contains("\"name\":\"Primary\""));
        assert!(grok_detail.contains("\"cookieToken\":\"gt-1\""));

        let grok_update = handle_request(build_request(
            "PUT",
            &format!("/admin/api/providers/grok/tokens/update/{grok_secondary_id}"),
            &[("Authorization", auth_header.as_str())],
            "{\"name\":\"Secondary\",\"cookieToken\":\"gt-2\",\"active\":true}",
        ));
        assert!(grok_update.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_update.contains("\"active\":true"));

        let grok_list = handle_request(build_request(
            "GET",
            "/admin/api/providers/grok/tokens/list",
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(grok_list.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_list.contains(&format!("\"id\":\"{grok_primary_id}\"")));
        assert!(grok_list.contains(&format!("\"id\":\"{grok_secondary_id}\"")));
        assert!(grok_list.contains("\"active\":false"));
        assert!(grok_list.contains("\"active\":true"));

        let grok_delete = handle_request(build_request(
            "DELETE",
            &format!("/admin/api/providers/grok/tokens/delete/{grok_primary_id}"),
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(grok_delete.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_delete.contains("\"ok\":true"));

        let grok_config = handle_request(build_request(
            "PUT",
            "/admin/api/providers/grok/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"apiUrl\":\"https://grok.test/chat\",\"proxyUrl\":\"http://127.0.0.1:7890\",\"cfCookies\":\"theme=dark\",\"cfClearance\":\"cf-token\",\"userAgent\":\"Mozilla/Test\",\"origin\":\"https://grok.test\",\"referer\":\"https://grok.test/\"}}",
        ));
        assert!(grok_config.starts_with("HTTP/1.1 200 OK"));
        assert!(grok_config.contains("\"apiUrl\":\"https://grok.test/chat\""));
        assert!(grok_config.contains("\"proxyUrl\":\"http://127.0.0.1:7890\""));
        assert!(grok_config.contains("\"cfClearance\":\"cf-token\""));

        let orchids = handle_request(build_request(
            "PUT",
            "/admin/api/providers/orchids/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"clientCookie\":\"orchids-cookie\",\"projectId\":\"project-1\",\"agentMode\":\"claude-sonnet-4.5\"}}",
        ));
        assert!(orchids.starts_with("HTTP/1.1 200 OK"));
        assert!(orchids.contains("\"clientCookie\":\"orchids-cookie\""));

        let web = handle_request(build_request(
            "PUT",
            "/admin/api/providers/web/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"baseUrl\":\"https://web.test\",\"type\":\"openai\",\"apiKey\":\"web-key\"}}",
        ));
        assert!(web.starts_with("HTTP/1.1 200 OK"));
        assert!(web.contains("\"baseUrl\":\"https://web.test\""));
        assert!(web.contains("\"type\":\"openai\""));
        assert!(web.contains("\"apiKey\":\"web-key\""));

        let chatgpt = handle_request(build_request(
            "PUT",
            "/admin/api/providers/chatgpt/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"baseUrl\":\"https://chatgpt.test\",\"token\":\"chatgpt-token\"}}",
        ));
        assert!(chatgpt.starts_with("HTTP/1.1 200 OK"));
        assert!(chatgpt.contains("\"baseUrl\":\"https://chatgpt.test\""));
        assert!(chatgpt.contains("\"token\":\"chatgpt-token\""));

        let zai_image = handle_request(build_request(
            "PUT",
            "/admin/api/providers/zai/image/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"sessionToken\":\"zai-image-session\",\"apiUrl\":\"https://image.test/generate\"}}",
        ));
        assert!(zai_image.starts_with("HTTP/1.1 200 OK"));
        assert!(zai_image.contains("\"sessionToken\":\"zai-image-session\""));
        assert!(zai_image.contains("\"apiUrl\":\"https://image.test/generate\""));

        let zai_tts = handle_request(build_request(
            "PUT",
            "/admin/api/providers/zai/tts/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"token\":\"zai-tts-token\",\"userId\":\"user-1\",\"apiUrl\":\"https://audio.test/tts\"}}",
        ));
        assert!(zai_tts.starts_with("HTTP/1.1 200 OK"));
        assert!(zai_tts.contains("\"token\":\"zai-tts-token\""));
        assert!(zai_tts.contains("\"userId\":\"user-1\""));

        let zai_ocr = handle_request(build_request(
            "PUT",
            "/admin/api/providers/zai/ocr/config",
            &[("Authorization", auth_header.as_str())],
            "{\"config\":{\"token\":\"zai-ocr-token\",\"apiUrl\":\"https://ocr.test/process\"}}",
        ));
        assert!(zai_ocr.starts_with("HTTP/1.1 200 OK"));
        assert!(zai_ocr.contains("\"token\":\"zai-ocr-token\""));
        assert!(zai_ocr.contains("\"apiUrl\":\"https://ocr.test/process\""));

        let status = handle_request(build_request(
            "GET",
            "/admin/api/status",
            &[("Authorization", auth_header.as_str())],
            "",
        ));
        assert!(status.starts_with("HTTP/1.1 200 OK"));
        assert!(status.contains("\"project\":\"any2api-rust\""));
        assert!(status.contains("\"defaultProvider\":\"grok\""));
        assert!(status.contains("\"cursor\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains(&format!("\"kiro\":{{\"count\":1,\"configured\":true,\"active\":\"{kiro_backup_id}\"}}")));
        assert!(status.contains(&format!("\"grok\":{{\"count\":1,\"configured\":true,\"active\":\"{grok_secondary_id}\"}}")));
        assert!(status.contains("\"orchids\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains("\"web\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains("\"chatgpt\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains("\"zaiImage\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains("\"zaiTTS\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));
        assert!(status.contains("\"zaiOCR\":{\"count\":1,\"configured\":true,\"active\":\"default\"}"));

        env_guard.reload();

        let relogin_bad = handle_request(build_request(
            "POST",
            "/api/admin/auth/login",
            &[],
            "{\"password\":\"changeme\"}",
        ));
        assert!(relogin_bad.starts_with("HTTP/1.1 401 Unauthorized"));

        let login_again = handle_request(build_request(
            "POST",
            "/api/admin/auth/login",
            &[],
            "{\"password\":\"newpass\"}",
        ));
        let token_again = response_token(&login_again).expect("missing token in second login response");
        let auth_header_again = format!("Bearer {token_again}");

        let persisted_settings = handle_request(build_request(
            "GET",
            "/admin/api/settings",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_settings.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_settings.contains("\"apiKey\":\"sk-rust\""));
        assert!(persisted_settings.contains("\"defaultProvider\":\"grok\""));

        let persisted_kiro = handle_request(build_request(
            "GET",
            "/admin/api/providers/kiro/accounts/list",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_kiro.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_kiro.contains("\"name\":\"Backup\""));
        assert!(!persisted_kiro.contains("\"name\":\"Main\""));

        let persisted_grok = handle_request(build_request(
            "GET",
            "/admin/api/providers/grok/tokens/list",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_grok.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_grok.contains("\"name\":\"Secondary\""));
        assert!(!persisted_grok.contains("\"name\":\"Primary\""));

        let persisted_grok_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/grok/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_grok_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_grok_config.contains("\"proxyUrl\":\"http://127.0.0.1:7890\""));
        assert!(persisted_grok_config.contains("\"cfClearance\":\"cf-token\""));

        let persisted_web_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/web/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_web_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_web_config.contains("\"baseUrl\":\"https://web.test\""));
        assert!(persisted_web_config.contains("\"type\":\"openai\""));

        let persisted_chatgpt_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/chatgpt/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_chatgpt_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_chatgpt_config.contains("\"baseUrl\":\"https://chatgpt.test\""));
        assert!(persisted_chatgpt_config.contains("\"token\":\"chatgpt-token\""));

        let persisted_zai_image_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/zai/image/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_zai_image_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_zai_image_config.contains("\"sessionToken\":\"zai-image-session\""));

        let persisted_zai_tts_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/zai/tts/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_zai_tts_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_zai_tts_config.contains("\"token\":\"zai-tts-token\""));
        assert!(persisted_zai_tts_config.contains("\"userId\":\"user-1\""));

        let persisted_zai_ocr_config = handle_request(build_request(
            "GET",
            "/admin/api/providers/zai/ocr/config",
            &[("Authorization", auth_header_again.as_str())],
            "",
        ));
        assert!(persisted_zai_ocr_config.starts_with("HTTP/1.1 200 OK"));
        assert!(persisted_zai_ocr_config.contains("\"token\":\"zai-ocr-token\""));

        let models = handle_request(build_request("GET", "/v1/models?provider=grok", &[], ""));
        assert!(models.starts_with("HTTP/1.1 200 OK"));
        assert!(models.contains("\"provider\":\"grok\""));
        assert!(models.contains("\"id\":\"grok-4\""));
    }

    #[test]
    fn openai_images_generation_uses_zai_image_provider() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("zai-image-upstream");

        let mock_url = spawn_mock_server(1, |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/image");
            assert_eq!(request.header("cookie"), Some("session=image-session"));
            assert!(request.body.contains("\"prompt\":\"draw a cat\""));
            assert!(request.body.contains("\"ratio\":\"16:9\""));
            assert!(request.body.contains("\"resolution\":\"2K\""));
            assert!(request.body.contains("\"rm_label_watermark\":true"));
            text_mock_response(
                "200 OK",
                "application/json",
                "{\"code\":0,\"message\":\"ok\",\"timestamp\":123,\"data\":{\"image\":{\"prompt\":\"draw a cat\",\"size\":\"1792x1024\",\"ratio\":\"16:9\",\"resolution\":\"2K\",\"image_url\":\"https://cdn.test/cat.png\",\"width\":1792,\"height\":1024}}}",
            )
        });

        admin_store()
            .replace_zai_image_config(ZAIImageRuntimeConfig {
                session_token: "image-session".to_string(),
                api_url: format!("{mock_url}/image"),
            })
            .expect("store zai image config");

        let response = handle_request(build_request(
            "POST",
            "/v1/images/generations",
            &[],
            "{\"prompt\":\"draw a cat\",\"size\":\"1792x1024\"}",
        ));
        assert!(response.starts_with("HTTP/1.1 200 OK"));
        assert!(response.contains("X-Newplatform2API-Provider: zai_image"));
        assert!(response.contains("\"url\":\"https://cdn.test/cat.png\""));
        assert!(response.contains("\"size\":\"1792x1024\""));
    }

    #[test]
    fn openai_audio_speech_returns_wav_bytes() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("zai-tts-upstream");

        let mock_url = spawn_mock_server(1, |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/tts");
            assert_eq!(request.header("authorization"), Some("Bearer tts-token"));
            assert!(request.body.contains("\"input_text\":\"hello world\""));
            assert!(request.body.contains("\"voice_id\":\"system_003\""));
            assert!(request.body.contains("\"voice_name\":\"通用男声\""));
            text_mock_response(
                "200 OK",
                "text/event-stream",
                "data: {\"audio\":\"T0s=\"}\n\ndata: [DONE]\n\n",
            )
        });

        admin_store()
            .replace_zai_tts_config(ZAITTSRuntimeConfig {
                token: "tts-token".to_string(),
                user_id: "user-tts".to_string(),
                api_url: format!("{mock_url}/tts"),
            })
            .expect("store zai tts config");

        let response = handle_request_bytes(build_request(
            "POST",
            "/v1/audio/speech",
            &[],
            "{\"input\":\"hello world\"}",
        ));
        let (head, body) = split_http_response(&response);
        assert!(head.starts_with("HTTP/1.1 200 OK"));
        assert!(head.contains("Content-Type: audio/wav"));
        assert!(head.contains("X-Newplatform2API-Provider: zai_tts"));
        assert_eq!(body, b"OK");
    }

    #[test]
    fn ocr_upload_uses_zai_ocr_provider() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("zai-ocr-upstream");

        let mock_url = spawn_mock_server(1, |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/ocr");
            assert_eq!(request.header("authorization"), Some("Bearer ocr-token"));
            assert!(request
                .header("content-type")
                .unwrap_or_default()
                .contains("multipart/form-data"));
            assert!(request.body.contains("filename=\"scan.txt\""));
            assert!(request.body.contains("hello ocr"));
            text_mock_response(
                "200 OK",
                "application/json",
                "{\"code\":0,\"message\":\"ok\",\"data\":{\"task_id\":\"ocr-1\",\"status\":\"succeeded\",\"file_name\":\"scan.txt\",\"file_size\":9,\"file_type\":\"text/plain\",\"file_url\":\"https://files.test/scan.txt\",\"created_at\":\"2026-01-01T00:00:00Z\",\"markdown_content\":\"hello ocr\",\"json_content\":\"{\\\"text\\\":\\\"hello ocr\\\"}\",\"layout\":[{\"page\":1}]}}",
            )
        });

        admin_store()
            .replace_zai_ocr_config(ZAIOCRRuntimeConfig {
                token: "ocr-token".to_string(),
                api_url: format!("{mock_url}/ocr"),
            })
            .expect("store zai ocr config");

        let boundary = "----any2api-ocr";
        let body = format!(
            "--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"scan.txt\"\r\nContent-Type: text/plain\r\n\r\nhello ocr\r\n--{boundary}--\r\n"
        )
        .into_bytes();
        let response = handle_request(build_bytes_request(
            "POST",
            "/v1/ocr",
            &[("Content-Type", &format!("multipart/form-data; boundary={boundary}"))],
            body,
        ));
        assert!(response.starts_with("HTTP/1.1 200 OK"));
        assert!(response.contains("X-Newplatform2API-Provider: zai_ocr"));
        assert!(response.contains("\"id\":\"ocr-1\""));
        assert!(response.contains("\"markdown\":\"hello ocr\""));
        assert!(response.contains("\"json\":{\"text\":\"hello ocr\"}"));
    }

    #[test]
    fn openai_chat_uses_grok_provider_real_upstream_shape() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("grok-upstream");

        admin_store()
            .replace_grok_tokens(vec![GrokToken {
                id: "grok-1".to_string(),
                name: "Main".to_string(),
                cookie_token: "cookie-token".to_string(),
                active: true,
            }])
            .expect("store grok token");

        let mock_url = spawn_mock_server(1, |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/grok");
            assert_eq!(request.header("origin"), Some("https://grok.test"));
            assert_eq!(request.header("referer"), Some("https://grok.test/"));
            assert_eq!(request.header("user-agent"), Some("Mozilla/Test"));
            assert!(request.header("cookie").unwrap_or_default().contains("theme=dark"));
            assert!(request.header("cookie").unwrap_or_default().contains("cf_clearance=cf-1"));
            assert!(request.body.contains("\"message\":\"reply exactly OK\""));
            text_mock_response(
                "200 OK",
                "application/json",
                "{\"result\":{\"response\":{\"token\":\"OK\"}}}\n",
            )
        });

        admin_store()
            .replace_grok_config(GrokRuntimeConfig {
                api_url: format!("{mock_url}/grok"),
                proxy_url: String::new(),
                cf_cookies: "theme=dark".to_string(),
                cf_clearance: "cf-1".to_string(),
                user_agent: "Mozilla/Test".to_string(),
                origin: "https://grok.test".to_string(),
                referer: "https://grok.test/".to_string(),
            })
            .expect("store grok config");

        let completion = handle_request(build_request(
            "POST",
            "/v1/chat/completions",
            &[],
            "{\"provider\":\"grok\",\"model\":\"grok-4\",\"messages\":[{\"role\":\"user\",\"content\":\"reply exactly OK\"}]}",
        ));
        assert!(completion.starts_with("HTTP/1.1 200 OK"));
        assert!(completion.contains("X-Newplatform2API-Provider: grok"));
        assert!(completion.contains("\"content\":\"OK\""));
    }

    #[test]
    fn openai_chat_uses_cursor_provider_real_upstream_shape() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("cursor-upstream");

        let mock_url = spawn_mock_server(1, |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/cursor");
            assert_eq!(request.header("x-is-human"), Some("human-token"));
            assert_eq!(request.header("cookie"), Some("cursor-cookie"));
            assert!(request.body.contains("\"trigger\":\"submit-message\""));
            assert!(request.body.contains("anthropic/claude-sonnet-4.6"));
            assert!(request.body.contains("reply exactly OK"));
            text_mock_response("200 OK", "application/json", "{\"text\":\"OK\"}")
        });

        admin_store()
            .replace_cursor_config(CursorRuntimeConfig {
                api_url: format!("{mock_url}/cursor"),
                script_url: format!("{mock_url}/script"),
                cookie: "cursor-cookie".to_string(),
                x_is_human: "human-token".to_string(),
                user_agent: String::new(),
                referer: String::new(),
                webgl_vendor: String::new(),
                webgl_renderer: String::new(),
            })
            .expect("store cursor config");

        let completion = handle_request(build_request(
            "POST",
            "/v1/chat/completions",
            &[],
            "{\"provider\":\"cursor\",\"model\":\"claude-sonnet-4.6\",\"messages\":[{\"role\":\"user\",\"content\":\"reply exactly OK\"}]}",
        ));
        assert!(completion.starts_with("HTTP/1.1 200 OK"));
        assert!(completion.contains("X-Newplatform2API-Provider: cursor"));
        assert!(completion.contains("\"content\":\"OK\""));
    }

    #[test]
    fn openai_chat_uses_kiro_provider_real_upstream_shape() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("kiro-upstream");

        admin_store()
            .replace_kiro_accounts(vec![KiroAccount {
                id: "kiro-1".to_string(),
                name: "Main".to_string(),
                access_token: "kiro-token".to_string(),
                machine_id: "machine-1".to_string(),
                preferred_endpoint: "codewhisperer".to_string(),
                active: true,
            }])
            .expect("store kiro account");

        let frame = kiro_frame(
            "assistantResponseEvent",
            "{\"content\":\"OK\"}",
        );
        let mock_url = spawn_mock_server(1, move |request| {
            assert_eq!(request.method, "POST");
            assert_eq!(request.path, "/kiro");
            assert!(request.body.contains("reply exactly OK"));
            MockResponse {
                status: "200 OK",
                content_type: "application/octet-stream",
                headers: vec![],
                body: frame.clone(),
            }
        });
        let _kiro_api = EnvVarGuard::set(
            "NEWPLATFORM2API_KIRO_CODEWHISPERER_URL",
            format!("{mock_url}/kiro"),
        );
        let _kiro_alt = EnvVarGuard::set(
            "NEWPLATFORM2API_KIRO_AMAZONQ_URL",
            format!("{mock_url}/kiro-alt"),
        );

        let completion = handle_request(build_request(
            "POST",
            "/v1/chat/completions",
            &[],
            "{\"provider\":\"kiro\",\"model\":\"claude-sonnet-4.6\",\"messages\":[{\"role\":\"user\",\"content\":\"reply exactly OK\"}]}",
        ));
        assert!(completion.starts_with("HTTP/1.1 200 OK"));
        assert!(completion.contains("X-Newplatform2API-Provider: kiro"));
        assert!(completion.contains("\"content\":\"OK\""));
    }

    #[test]
    fn anthropic_messages_use_orchids_provider_real_upstream_shape() {
        let _serial = test_lock().lock().expect("test lock poisoned");
        let _env = TestEnvGuard::new("orchids-upstream");

        admin_store()
            .replace_orchids_config(OrchidsRuntimeConfig {
                api_url: String::new(),
                clerk_url: String::new(),
                client_cookie: "client-cookie".to_string(),
                client_uat: "123".to_string(),
                session_id: String::new(),
                project_id: "project-1".to_string(),
                user_id: String::new(),
                email: String::new(),
                agent_mode: "claude-sonnet-4.5".to_string(),
            })
            .expect("store orchids config");

        let mock_url = spawn_mock_server(3, |request| {
            if request.path.starts_with("/v1/client/sessions/sess-1/tokens") {
                return text_mock_response(
                    "200 OK",
                    "application/json",
                    "{\"jwt\":\"orchids-jwt\"}",
                );
            }
            if request.path.starts_with("/v1/client") {
                return text_mock_response(
                    "200 OK",
                    "application/json",
                    "{\"response\":{\"last_active_session_id\":\"sess-1\",\"sessions\":[{\"user\":{\"id\":\"user-1\",\"email_addresses\":[{\"email_address\":\"u@example.com\"}]}}]}}",
                );
            }
            assert_eq!(request.path, "/agent");
            assert!(request.body.contains("reply exactly OK"));
            text_mock_response(
                "200 OK",
                "text/event-stream",
                "data: {\"type\":\"model\",\"event\":{\"type\":\"text-delta\",\"delta\":\"OK\"}}\n\n",
            )
        });
        let _orchids_api = EnvVarGuard::set("NEWPLATFORM2API_ORCHIDS_API_URL", format!("{mock_url}/agent"));
        let _orchids_clerk = EnvVarGuard::set("NEWPLATFORM2API_ORCHIDS_CLERK_URL", mock_url.clone());

        let completion = handle_request(build_request(
            "POST",
            "/v1/messages",
            &[],
            "{\"provider\":\"orchids\",\"model\":\"claude-sonnet-4.5\",\"messages\":[{\"role\":\"user\",\"content\":\"reply exactly OK\"}]}",
        ));
        assert!(completion.starts_with("HTTP/1.1 200 OK"));
        assert!(completion.contains("X-Newplatform2API-Provider: orchids"));
        assert!(completion.contains("\"text\":\"OK\""));
    }
}