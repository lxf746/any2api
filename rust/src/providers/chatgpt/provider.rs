use std::time::Duration;

use reqwest::blocking::Client;
use serde_json::{json, Value};

use crate::admin_store::ChatGPTRuntimeConfig;
use crate::providers::common::{content_text, env_int, normalize_messages, split_system_messages, DEFAULT_MAX_INPUT_LENGTH};
use crate::registry::Provider;
use crate::types::{ModelInfo, ProviderCapabilities, UnifiedRequest};

const DEFAULT_TIMEOUT_SECONDS: u64 = 60;

pub struct ChatGPTProvider {
    config: ChatGPTRuntimeConfig,
    client: Client,
}

impl ChatGPTProvider {
    pub fn new(config: ChatGPTRuntimeConfig) -> Self {
        let timeout = Duration::from_secs(env_int("NEWPLATFORM2API_CHATGPT_TIMEOUT", DEFAULT_TIMEOUT_SECONDS));
        let client = Client::builder()
            .timeout(timeout)
            .build()
            .expect("build chatgpt reqwest client");
        Self { config: normalize_chatgpt_runtime_config(config), client }
    }

    fn chat_url(&self) -> String {
        format!("{}/v1/chat/completions", self.config.base_url)
    }

    fn map_model(&self, model: &str) -> String {
        let trimmed = model.trim();
        if trimmed.is_empty() {
            "gpt-4.1".to_string()
        } else {
            trimmed.to_string()
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
            .map_err(|err| format!("decode chatgpt response: {err}"))?;
        if !payload.is_object() {
            return Err("chatgpt upstream returned invalid response".to_string());
        }
        if !payload["error"].is_null() {
            return Err(format!("chatgpt upstream error: {}", payload["error"]));
        }
        let Some(choice) = payload["choices"].as_array().and_then(|items| items.first()) else {
            return Err("chatgpt upstream returned no choices".to_string());
        };
        let text = content_text(&choice["message"]["content"]);
        if text.is_empty() {
            return Err("chatgpt upstream returned empty content".to_string());
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
                return Err(format!("chatgpt upstream error: {}", value["error"]));
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
            return Err("chatgpt upstream returned empty content".to_string());
        }
        Ok(text)
    }
}

impl Provider for ChatGPTProvider {
    fn id(&self) -> &'static str {
        "chatgpt"
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
            provider: "chatgpt",
            public_model: "gpt-4.1",
            upstream_model: "gpt-4.1",
            owned_by: "openai",
        }]
    }

    fn build_upstream_preview(&self, req: &UnifiedRequest) -> String {
        format!(
            "url={} auth=bearer token protocol={} model={} configured={} token_set={}",
            self.chat_url(),
            req.protocol,
            self.map_model(&req.model),
            !self.config.base_url.is_empty() && !self.config.token.is_empty(),
            !self.config.token.is_empty(),
        )
    }

    fn generate_reply(&self, req: &UnifiedRequest) -> Result<String, String> {
        if self.config.base_url.is_empty() {
            return Err("chatgpt base url is not configured".to_string());
        }
        if self.config.token.is_empty() {
            return Err("chatgpt token is not configured".to_string());
        }
        let response = self
            .client
            .post(self.chat_url())
            .header("Authorization", format!("Bearer {}", self.config.token))
            .header("Content-Type", "application/json")
            .header("Accept", if req.stream { "text/event-stream" } else { "application/json" })
            .body(self.build_payload(req).to_string())
            .send()
            .map_err(|err| format!("chatgpt upstream request failed: {err}"))?;
        let status = response.status();
        let content_type = response
            .headers()
            .get(reqwest::header::CONTENT_TYPE)
            .and_then(|value| value.to_str().ok())
            .unwrap_or_default()
            .to_string();
        let body = response
            .text()
            .map_err(|err| format!("read chatgpt upstream body: {err}"))?;
        if !status.is_success() {
            return Err(format!("chatgpt upstream error: status={} body={}", status.as_u16(), body.trim()));
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

fn normalize_chatgpt_runtime_config(mut config: ChatGPTRuntimeConfig) -> ChatGPTRuntimeConfig {
    config.base_url = config.base_url.trim().trim_end_matches('/').to_string();
    config.token = config.token.trim().to_string();
    if config.base_url.is_empty() {
        config.base_url = "http://127.0.0.1:5005".to_string();
    }
    config
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn chatgpt_sse_response_preserves_delta_whitespace() {
        let provider = ChatGPTProvider::new(ChatGPTRuntimeConfig::default());
        let body = "data: {\"choices\":[{\"delta\":{\"content\":\"chat \"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n";
        assert_eq!(provider.collect_sse_text(body).unwrap(), "chat ok");
    }
}