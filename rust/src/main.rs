use std::env;

fn resolve_bind_addr<F>(get_env: F) -> String
where
    F: Fn(&str) -> Option<String>,
{
    for key in ["NEWPLATFORM2API_BIND_ADDR", "BIND_ADDR"] {
        if let Some(value) = get_env(key) {
            let trimmed = value.trim();
            if !trimmed.is_empty() {
                return trimmed.to_string();
            }
        }
    }

    let host = ["NEWPLATFORM2API_HOST", "HOST"]
        .into_iter()
        .filter_map(&get_env)
        .map(|value| value.trim().to_string())
        .find(|value| !value.is_empty())
        .unwrap_or_else(|| "127.0.0.1".to_string());

    let port = ["NEWPLATFORM2API_PORT", "PORT"]
        .into_iter()
        .filter_map(get_env)
        .map(|value| value.trim().to_string())
        .find_map(|value| value.parse::<u16>().ok())
        .unwrap_or(8101);

    format!("{host}:{port}")
}

fn current_bind_addr() -> String {
    resolve_bind_addr(|key| env::var(key).ok())
}

fn main() -> std::io::Result<()> {
    any2api_rust::server::run(&current_bind_addr())
}

#[cfg(test)]
mod tests {
    use std::collections::HashMap;

    use super::resolve_bind_addr;

    #[test]
    fn prefers_explicit_bind_addr() {
        let values = HashMap::from([
            ("NEWPLATFORM2API_BIND_ADDR", "0.0.0.0:9001"),
            ("NEWPLATFORM2API_HOST", "127.0.0.1"),
            ("NEWPLATFORM2API_PORT", "8101"),
        ]);
        assert_eq!(resolve_bind_addr(|key| values.get(key).map(|value| value.to_string())), "0.0.0.0:9001");
    }

    #[test]
    fn falls_back_to_host_and_port_env() {
        let values = HashMap::from([("HOST", "0.0.0.0"), ("PORT", "9100")]);
        assert_eq!(resolve_bind_addr(|key| values.get(key).map(|value| value.to_string())), "0.0.0.0:9100");
    }

    #[test]
    fn uses_local_defaults_when_env_missing() {
        let values = HashMap::<&str, &str>::new();
        assert_eq!(resolve_bind_addr(|key| values.get(key).map(|value| value.to_string())), "127.0.0.1:8101");
    }
}