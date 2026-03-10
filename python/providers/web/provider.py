from __future__ import annotations

import json
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from gateway_types import ModelInfo, ProviderCapabilities, UnifiedRequest
from ..common import (
    DEFAULT_TIMEOUT_SECONDS,
    ProviderRequestError,
    SnapshotGetter,
    StoreBackedProvider,
    _content_text,
    _env_int,
    _normalize_messages,
    _read_http_error,
    _split_system_messages,
    _string,
)


def _event_text(value: Any) -> str:
    if isinstance(value, str):
        return value
    return _content_text(value)


class WebProvider(StoreBackedProvider):
    def __init__(self, snapshot_getter: SnapshotGetter | None):
        super().__init__(
            "web",
            ProviderCapabilities(openai_compatible=True),
            [ModelInfo("web", "claude-sonnet-4.5", "claude-sonnet-4.5", "claude")],
            snapshot_getter,
        )

    def build_upstream_preview(self, req: UnifiedRequest) -> dict:
        config = self._config()
        return {
            "url": self._chat_url(config),
            "auth": "bearer api key (optional)",
            "live_enabled": True,
            "configured": bool(config["baseUrl"] and config["type"]),
            "type": config["type"],
            "api_key_set": bool(config["apiKey"]),
            "mapped_model": self._map_model(req.model, config["type"]),
            "message_count": len(req.messages),
        }

    def generate_reply(self, req: UnifiedRequest) -> str:
        config = self._config()
        if not config["baseUrl"]:
            raise ProviderRequestError("web base url is not configured")
        if not config["type"]:
            raise ProviderRequestError("web type is not configured")
        payload = self._build_payload(req, config, stream=bool(req.stream))
        request = Request(
            self._chat_url(config),
            data=json.dumps(payload).encode("utf-8"),
            headers=self._headers(config, bool(req.stream)),
            method="POST",
        )
        try:
            with urlopen(request, timeout=_env_int("NEWPLATFORM2API_WEB_TIMEOUT", DEFAULT_TIMEOUT_SECONDS)) as response:
                raw = response.read().decode("utf-8", errors="replace")
                content_type = response.headers.get("Content-Type", "")
                if "text/event-stream" in content_type.lower() or raw.lstrip().startswith("data:"):
                    return self._collect_sse_text(raw)
                return self._collect_json_text(raw)
        except HTTPError as exc:
            raise ProviderRequestError(f"web upstream error: status={exc.code} body={_read_http_error(exc)}") from exc
        except URLError as exc:
            raise ProviderRequestError(f"web upstream request failed: {exc.reason}") from exc

    def _config(self) -> dict[str, str]:
        providers = self._snapshot().get("providers", {})
        raw = providers.get("webConfig") if isinstance(providers, dict) and isinstance(providers.get("webConfig"), dict) else {}
        return {
            "baseUrl": _string(raw.get("baseUrl")).rstrip("/") or "http://127.0.0.1:9000",
            "type": _string(raw.get("type")).strip("/") or "claude",
            "apiKey": _string(raw.get("apiKey")),
        }

    def _headers(self, config: dict[str, str], stream: bool) -> dict[str, str]:
        headers = {
            "Accept": "text/event-stream" if stream else "application/json",
            "Content-Type": "application/json",
        }
        if config["apiKey"]:
            headers["Authorization"] = f"Bearer {config['apiKey']}"
        return headers

    def _build_payload(self, req: UnifiedRequest, config: dict[str, str], stream: bool) -> dict[str, Any]:
        normalized = _normalize_messages(req.messages)
        system, messages = _split_system_messages(normalized)
        if req.system is not None:
            system = "\n\n".join(part for part in [system, _content_text(req.system)] if part).strip()
        if system:
            messages = [{"role": "system", "content": system}, *messages]
        return {
            "model": self._map_model(req.model, config["type"]),
            "messages": messages,
            "stream": stream,
        }

    def _chat_url(self, config: dict[str, str]) -> str:
        return f"{config['baseUrl']}/{config['type']}/v1/chat/completions"

    def _map_model(self, model: str, type_name: str) -> str:
        trimmed = _string(model)
        if trimmed:
            return trimmed
        type_lower = _string(type_name).lower()
        if type_lower == "claude":
            return "claude-sonnet-4.5"
        if type_lower == "openai":
            return "gpt-4.1"
        return type_name or "claude-sonnet-4.5"

    def _collect_json_text(self, raw: str) -> str:
        try:
            payload = json.loads(raw)
        except json.JSONDecodeError as exc:
            raise ProviderRequestError(f"decode web response: {exc}") from exc
        if not isinstance(payload, dict):
            raise ProviderRequestError("web upstream returned invalid response")
        if payload.get("error") is not None:
            raise ProviderRequestError(f"web upstream error: {json.dumps(payload.get('error'), ensure_ascii=False)}")
        choices = payload.get("choices")
        if not isinstance(choices, list) or not choices:
            raise ProviderRequestError("web upstream returned no choices")
        message = choices[0].get("message") if isinstance(choices[0], dict) else {}
        text = _content_text(message.get("content") if isinstance(message, dict) else None)
        if not text:
            raise ProviderRequestError("web upstream returned empty content")
        return text

    def _collect_sse_text(self, raw: str) -> str:
        parts: list[str] = []
        for line in raw.splitlines():
            line = line.strip()
            if not line.startswith("data:"):
                continue
            data = line.removeprefix("data:").strip()
            if not data or data == "[DONE]":
                continue
            try:
                payload = json.loads(data)
            except json.JSONDecodeError:
                continue
            if not isinstance(payload, dict):
                continue
            if payload.get("error") is not None:
                raise ProviderRequestError(f"web upstream error: {json.dumps(payload.get('error'), ensure_ascii=False)}")
            choices = payload.get("choices")
            if not isinstance(choices, list):
                continue
            for choice in choices:
                if not isinstance(choice, dict):
                    continue
                delta = choice.get("delta") if isinstance(choice.get("delta"), dict) else {}
                message = choice.get("message") if isinstance(choice.get("message"), dict) else {}
                text = _event_text(delta.get("content")) or _event_text(message.get("content"))
                if text:
                    parts.append(text)
        text = "".join(parts).strip()
        if not text:
            raise ProviderRequestError("web upstream returned empty content")
        return text