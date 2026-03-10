use std::time::Duration;

use reqwest::blocking::Client;
use serde_json::{json, Value};

use crate::admin_store::WebRuntimeConfig;
use crate::providers::common::{content_text, env_int, normalize_messages, split_system_messages, DEFAULT_MAX_INPUT_LENGTH};
use crate::registry::Provider;
use crate::types::{ModelInfo, ProviderCapabilities, UnifiedRequest};

const DEFAULT_TIMEOUT_SECONDS: u64 = 60;

pub struct WebProvider {
    config: WebRuntimeConfig,
    client: Client,
}

impl WebProvider {
    pub fn new(config: WebRuntimeConfig) -> Self {
        let timeout = Duration::from_secs(env_int("NEWPLATFORM2API_WEB_TIMEOUT", DEFAULT_TIMEOUT_SECONDS));
        let client = Client::builder()
            .timeout(timeout)
            .build()
            .expect("build web reqwest client");
        Self { config: normalize_web_runtime_config(config), client }
    }

    fn chat_url(&self) -> String {
        format!("{}/{}/v1/chat/completions", self.config.base_url, self.config.type_name)
    }

    fn map_model(&self, model: &str) -> String {
        let trimmed = model.trim();
        if !trimmed.is_empty() {
            return trimmed.to_string();
        }
        match self.config.type_name.to_ascii_lowercase().as_str() {
            "claude" => "claude-sonnet-4.5".to_string(),
            "openai" => "gpt-4.1".to_string(),
            other if !other.is_empty() => other.to_string(),
            _ => "claude-sonnet-4.5".to_string(),
        }
    }

    fn build_payload(&self, req: &UnifiedRequest) -> Value {
        let normalized = normalize_messages(req, DEFAULT_MAX_INPUT_LENGTH);
        let (system, non_system) = split_system_messages(&normalized);
        let mut messages = non_system
            .iter()
            .map(|message| json!({
                "role": message.role.as_str(),
                "content": content_text(&message.content),
            }))
            .collect::<Vec<_>>();
        if !system.is_empty() {
            messages.insert(0, json!({"role": "system", "content": system}));
        }
        json!({
            "model": self.map_model(&req.model),
            "messages": messages,
            "stream": req.stream,
        })
    }

    fn collect_json_text(&self, body: &str) -> Result<String, String> {
        let payload: Value = serde_json::from_str(body)
            .map_err(|err| format!("decode web response: {err}"))?;
        if !payload.is_object() {
            return Err("web upstream returned invalid response".to_string());
        }
        if !payload["error"].is_null() {
            return Err(format!("web upstream error: {}", payload["error"]));
        }
        let Some(choice) = payload["choices"].as_array().and_then(|items| items.first()) else {
            return Err("web upstream returned no choices".to_string());
        };
        let text = content_text(&choice["message"]["content"]);
        if text.is_empty() {
            return Err("web upstream returned empty content".to_string());
        }
        Ok(text)
    }

    fn collect_sse_text(&self, body: &str) -> Result<String, String> {
        let mut output = String::new();
        for raw_line in body.lines() {
            let line = raw_line.trim();
            if !line.starts_with("data:") {
                continue;
            }
            let payload = line.trim_start_matches("data:").trim();
            if payload.is_empty() || payload == "[DONE]" {
                continue;
            }
            let value: Value = match serde_json::from_str(payload) {
                Ok(value) => value,
                Err(_) => continue,
            };
            if !value["error"].is_null() {
                return Err(format!("web upstream error: {}", value["error"]));
            }
            let Some(choices) = value["choices"].as_array() else {
                continue;
            };
            for choice in choices {
                let delta = event_content_text(&choice["delta"]["content"]);
                let message = event_content_text(&choice["message"]["content"]);
                if !delta.is_empty() {
                    output.push_str(&delta);
                } else if !message.is_empty() {
                    output.push_str(&message);
                }
            }
        }
        let text = output.trim().to_string();
        if text.is_empty() {
            return Err("web upstream returned empty content".to_string());
        }
        Ok(text)
    }
}

impl Provider for WebProvider {
    fn id(&self) -> &'static str {
        "web"
    }

    fn capabilities(&self) -> ProviderCapabilities {
        ProviderCapabilities {
            openai_compatible: true,
            anthropic_compatible: false,
            tools: false,
            images: false,
            multi_account: false,
        }
    }

    fn models(&self) -> Vec<ModelInfo> {
        vec![ModelInfo {
            provider: "web",
            public_model: "claude-sonnet-4.5",
            upstream_model: "claude-sonnet-4.5",
            owned_by: "claude",
        }]
    }

    fn build_upstream_preview(&self, req: &UnifiedRequest) -> String {
        format!(
            "url={} auth=bearer api key (optional) protocol={} type={} model={} configured={} api_key_set={}",
            self.chat_url(),
            req.protocol,
            self.config.type_name,
            self.map_model(&req.model),
            !self.config.base_url.is_empty() && !self.config.type_name.is_empty(),
            !self.config.api_key.is_empty(),
        )
    }

    fn generate_reply(&self, req: &UnifiedRequest) -> Result<String, String> {
        if self.config.base_url.is_empty() {
            return Err("web base url is not configured".to_string());
        }
        if self.config.type_name.is_empty() {
            return Err("web type is not configured".to_string());
        }
        let mut request = self
            .client
            .post(self.chat_url())
            .header("Content-Type", "application/json")
            .header("Accept", if req.stream { "text/event-stream" } else { "application/json" })
            .body(self.build_payload(req).to_string());
        if !self.config.api_key.is_empty() {
            request = request.header("Authorization", format!("Bearer {}", self.config.api_key));
        }
        let response = request
            .send()
            .map_err(|err| format!("web upstream request failed: {err}"))?;
        let status = response.status();
        let content_type = response
            .headers()
            .get(reqwest::header::CONTENT_TYPE)
            .and_then(|value| value.to_str().ok())
            .unwrap_or_default()
            .to_string();
        let body = response
            .text()
            .map_err(|err| format!("read web upstream body: {err}"))?;
        if !status.is_success() {
            return Err(format!("web upstream error: status={} body={}", status.as_u16(), body.trim()));
        }
        if content_type.contains("text/event-stream") || body.trim_start().starts_with("data:") {
            return self.collect_sse_text(&body);
        }
        self.collect_json_text(&body)
    }
}

fn event_content_text(content: &Value) -> String {
    match content {
        Value::String(text) => text.clone(),
        _ => content_text(content),
    }
}

fn normalize_web_runtime_config(mut config: WebRuntimeConfig) -> WebRuntimeConfig {
    config.base_url = config.base_url.trim().trim_end_matches('/').to_string();
    config.type_name = config.type_name.trim().trim_matches('/').to_string();
    config.api_key = config.api_key.trim().to_string();
    if config.base_url.is_empty() {
        config.base_url = "http://127.0.0.1:9000".to_string();
    }
    if config.type_name.is_empty() {
        config.type_name = "claude".to_string();
    }
    config
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn web_sse_response_preserves_delta_whitespace() {
        let provider = WebProvider::new(WebRuntimeConfig::default());
        let body = "data: {\"choices\":[{\"delta\":{\"content\":\"chat \"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n";
        assert_eq!(provider.collect_sse_text(body).unwrap(), "chat ok");
    }
}