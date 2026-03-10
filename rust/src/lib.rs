pub mod admin_store;
pub mod providers;
pub mod registry;
pub mod server;
pub mod types;

#[cfg(test)]
mod tests {
    use crate::admin_store::AdminConfig;
    use crate::providers::default_registry;

    #[test]
    fn registry_contains_six_providers() {
        let registry = default_registry("cursor", &AdminConfig::default());
        assert_eq!(registry.provider_ids(), vec!["chatgpt", "cursor", "grok", "kiro", "orchids", "web"]);
    }
}
