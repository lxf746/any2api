mod common;
mod chatgpt;
mod cursor;
mod grok;
mod kiro;
mod orchids;
mod web;
pub mod zai_ocr;
pub mod zai_image;
pub mod zai_tts;

use crate::admin_store::AdminConfig;
use crate::registry::Registry;

pub fn default_registry(default_provider: &str, snapshot: &AdminConfig) -> Registry {
    let mut registry = Registry::new(default_provider);
    registry.register(Box::new(cursor::CursorProvider::new(
        snapshot.providers.cursor_config.clone(),
    )));
    registry.register(Box::new(kiro::KiroProvider::new(
        snapshot.providers.kiro_accounts.clone(),
    )));
    registry.register(Box::new(grok::GrokProvider::new(
        snapshot.providers.grok_tokens.clone(),
        snapshot.providers.grok_config.clone(),
    )));
    registry.register(Box::new(orchids::OrchidsProvider::new(
        snapshot.providers.orchids_config.clone(),
    )));
    registry.register(Box::new(web::WebProvider::new(
        snapshot.providers.web_config.clone(),
    )));
    registry.register(Box::new(chatgpt::ChatGPTProvider::new(
        snapshot.providers.chatgpt_config.clone(),
    )));
    registry
}

pub use chatgpt::ChatGPTProvider;
pub use cursor::CursorProvider;
pub use grok::GrokProvider;
pub use kiro::KiroProvider;
pub use orchids::OrchidsProvider;
pub use web::WebProvider;
