from __future__ import annotations

import json
import os
import re
import time
import uuid
from email.utils import parsedate_to_datetime
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.parse import urlparse
from urllib.request import ProxyHandler, Request, build_opener, urlopen

from gateway_types import ModelInfo, ProviderCapabilities, UnifiedRequest
from ..common import (
    DEFAULT_TIMEOUT_SECONDS,
    ProviderRequestError,
    SnapshotGetter,
    StoreBackedProvider,
    _content_text,
    _env_int,
    _nested_map,
    _normalize_incremental_chunk,
    _normalize_messages,
    _pick_active_item,
    _read_http_error,
    _string,
    _text_value,
)

DEFAULT_GROK_API_URL = "https://grok.com/rest/app-chat/conversations/new"
DEFAULT_GROK_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
DEFAULT_GROK_ORIGIN = "https://grok.com"
DEFAULT_GROK_REFERER = "https://grok.com/"
GROK_MAX_RETRIES = 3
GROK_RETRY_BUDGET_SECONDS = 12.0
GROK_RETRY_BACKOFF_BASE_SECONDS = 0.4
GROK_RETRY_BACKOFF_MAX_SECONDS = 4.0

_GROK_TOOL_USAGE_CARD_RE = re.compile(r"<xai:tool_usage_card[^>]*>.*?</xai:tool_usage_card>", re.DOTALL)
_GROK_TOOL_NAME_RE = re.compile(r"<xai:tool_name>(.*?)</xai:tool_name>", re.DOTALL)
_GROK_TOOL_ARGS_RE = re.compile(r"<xai:tool_args>(.*?)</xai:tool_args>", re.DOTALL)
_GROK_CDATA_RE = re.compile(r"<!\[CDATA\[(.*?)\]\]>", re.DOTALL)
_GROK_ROLLOUT_RE = re.compile(r"<rolloutId>.*?</rolloutId>", re.DOTALL)
_GROK_SPECIAL_TAG_RE = re.compile(r"</?xai:[^>]+>")


class GrokStreamFilter:
    def __init__(self) -> None:
        self.tool_card_open = False
        self.buffer = ""

    def filter(self, token: str) -> str:
        if not token:
            return ""
        start_tag = "<xai:tool_usage_card"
        end_tag = "</xai:tool_usage_card>"
        output: list[str] = []
        remaining = token
        while remaining:
            if self.tool_card_open:
                end_index = remaining.find(end_tag)
                if end_index == -1:
                    self.buffer += remaining
                    return "".join(output)
                end_pos = end_index + len(end_tag)
                self.buffer += remaining[:end_pos]
                summary = _summarize_grok_tool_card(self.buffer)
                if summary:
                    output.append(summary)
                    if not summary.endswith("\n"):
                        output.append("\n")
                self.buffer = ""
                self.tool_card_open = False
                remaining = remaining[end_pos:]
                continue
            start_index = remaining.find(start_tag)
            if start_index == -1:
                output.append(remaining)
                break
            if start_index > 0:
                output.append(remaining[:start_index])
            end_index = remaining[start_index:].find(end_tag)
            if end_index == -1:
                self.tool_card_open = True
                self.buffer += remaining[start_index:]
                break
            end_pos = start_index + end_index + len(end_tag)
            summary = _summarize_grok_tool_card(remaining[start_index:end_pos])
            if summary:
                output.append(summary)
                if not summary.endswith("\n"):
                    output.append("\n")
            remaining = remaining[end_pos:]
        return "".join(output)


class GrokProvider(StoreBackedProvider):
    def __init__(self, snapshot_getter: SnapshotGetter | None):
        super().__init__(
            "grok",
            ProviderCapabilities(openai_compatible=True, tools=True, images=True, multi_account=True),
            [ModelInfo("grok", "grok-4", "grok-4", "xai")],
            snapshot_getter,
        )
        self._sleep = time.sleep

    def build_upstream_preview(self, req: UnifiedRequest) -> dict:
        payload = self._build_payload(req)
        token = self._token()
        return {
            "url": self._api_url(),
            "auth": "grok sso cookie token",
            "live_enabled": True,
            "cookie_configured": bool(_string(token.get("cookieToken"))),
            "proxy_configured": bool(self._proxy_url()),
            "cf_configured": bool(self._cf_cookies() or self._cf_clearance()),
            "payload": {"model": payload["modelName"], "message_len": len(payload["message"]), "message_count": len(req.messages)},
        }

    def generate_reply(self, req: UnifiedRequest) -> str:
        token = self._token()
        cookie_token = _string(token.get("cookieToken"))
        if not cookie_token:
            raise ProviderRequestError("grok cookie token is not configured")
        total_delay = 0.0
        last_delay = GROK_RETRY_BACKOFF_BASE_SECONDS
        timeout = _env_int("NEWPLATFORM2API_GROK_TIMEOUT", DEFAULT_TIMEOUT_SECONDS)
        for attempt in range(GROK_MAX_RETRIES + 1):
            request = Request(self._api_url(), data=json.dumps(self._build_payload(req)).encode("utf-8"), headers=self._headers(cookie_token), method="POST")
            try:
                with self._open_request(request, timeout=timeout) as response:
                    return self._collect_text(response.read().decode("utf-8", errors="replace"))
            except HTTPError as exc:
                error = ProviderRequestError(f"grok upstream error: status={exc.code} body={_read_http_error(exc)}")
                if attempt == GROK_MAX_RETRIES or not _should_retry_grok_status(exc.code):
                    raise error from exc
                delay = _grok_retry_delay(last_delay, _parse_retry_after_seconds(exc.headers), exc.code, attempt)
                if total_delay + delay > GROK_RETRY_BUDGET_SECONDS:
                    raise error from exc
                total_delay += delay
                last_delay = max(delay, GROK_RETRY_BACKOFF_BASE_SECONDS)
                if delay > 0:
                    self._sleep(delay)
            except URLError as exc:
                error = ProviderRequestError(f"grok upstream request failed: {exc.reason}")
                if attempt == GROK_MAX_RETRIES or not _should_retry_grok_transport_error(exc):
                    raise error from exc
                delay = _grok_retry_delay(last_delay, 0.0, 0, attempt)
                if total_delay + delay > GROK_RETRY_BUDGET_SECONDS:
                    raise error from exc
                total_delay += delay
                last_delay = max(delay, GROK_RETRY_BACKOFF_BASE_SECONDS)
                if delay > 0:
                    self._sleep(delay)
        raise ProviderRequestError("grok upstream request failed after retries")

    def _token(self) -> dict[str, Any]:
        providers = self._snapshot().get("providers", {})
        tokens = providers.get("grokTokens") if isinstance(providers, dict) else []
        return _pick_active_item(tokens, "cookieToken")

    def _runtime_config(self) -> dict[str, Any]:
        providers = self._snapshot().get("providers", {})
        config = providers.get("grokConfig") if isinstance(providers, dict) else {}
        return config if isinstance(config, dict) else {}

    def _config_value(self, key: str, env_name: str, default: str = "") -> str:
        config = self._runtime_config()
        if key in config:
            return _string(config.get(key))
        return os.getenv(env_name, default).strip() or default

    def _api_url(self) -> str:
        return self._config_value("apiUrl", "NEWPLATFORM2API_GROK_API_URL", DEFAULT_GROK_API_URL)

    def _proxy_url(self) -> str:
        return self._config_value("proxyUrl", "NEWPLATFORM2API_GROK_PROXY_URL")

    def _cf_cookies(self) -> str:
        return self._config_value("cfCookies", "NEWPLATFORM2API_GROK_CF_COOKIES")

    def _cf_clearance(self) -> str:
        return self._config_value("cfClearance", "NEWPLATFORM2API_GROK_CF_CLEARANCE")

    def _headers(self, cookie_token: str) -> dict[str, str]:
        return {
            "Accept": "*/*",
            "Accept-Encoding": "gzip, deflate, br",
            "Accept-Language": "en-US,en;q=0.9",
            "Content-Type": "application/json",
            "Cookie": _build_grok_cookie_header(cookie_token, self._cf_cookies(), self._cf_clearance()),
            "Origin": self._config_value("origin", "NEWPLATFORM2API_GROK_ORIGIN", DEFAULT_GROK_ORIGIN),
            "Priority": "u=1, i",
            "Referer": self._config_value("referer", "NEWPLATFORM2API_GROK_REFERER", DEFAULT_GROK_REFERER),
            "Sec-Fetch-Dest": "empty",
            "Sec-Fetch-Mode": "cors",
            "Sec-Fetch-Site": _grok_sec_fetch_site(
                self._config_value("origin", "NEWPLATFORM2API_GROK_ORIGIN", DEFAULT_GROK_ORIGIN),
                self._config_value("referer", "NEWPLATFORM2API_GROK_REFERER", DEFAULT_GROK_REFERER),
            ),
            "User-Agent": self._config_value("userAgent", "NEWPLATFORM2API_GROK_USER_AGENT", DEFAULT_GROK_USER_AGENT),
            "X-Statsig-Id": uuid.uuid4().hex[:16],
            "X-XAI-Request-Id": uuid.uuid4().hex,
            "X-Requested-With": "XMLHttpRequest",
        }

    def _open_request(self, request: Request, timeout: int):
        proxy_url = self._proxy_url()
        if not proxy_url:
            return urlopen(request, timeout=timeout)
        scheme = urlparse(proxy_url).scheme.lower()
        if scheme not in ("http", "https"):
            raise ProviderRequestError(f"grok proxy scheme is not supported by python backend: {proxy_url}")
        opener = build_opener(ProxyHandler({"http": proxy_url, "https": proxy_url}))
        return opener.open(request, timeout=timeout)

    def _build_payload(self, req: UnifiedRequest) -> dict[str, Any]:
        return {
            "deviceEnvInfo": {
                "darkModeEnabled": False,
                "devicePixelRatio": 2,
                "screenWidth": 2056,
                "screenHeight": 1329,
                "viewportWidth": 2056,
                "viewportHeight": 1083,
            },
            "disableMemory": False,
            "disableSearch": False,
            "disableSelfHarmShortCircuit": False,
            "disableTextFollowUps": False,
            "enableImageGeneration": True,
            "enableImageStreaming": True,
            "enableSideBySide": True,
            "fileAttachments": [],
            "forceConcise": False,
            "forceSideBySide": False,
            "imageAttachments": [],
            "imageGenerationCount": 2,
            "isAsyncChat": False,
            "isReasoning": False,
            "message": self._flatten_messages(req),
            "modelName": _string(req.model) or "grok-4",
            "responseMetadata": {"requestModelDetails": {"modelId": _string(req.model) or "grok-4"}},
            "returnImageBytes": False,
            "returnRawGrokInXaiRequest": False,
            "sendFinalMetadata": True,
            "temporary": False,
            "toolOverrides": {},
        }

    def _flatten_messages(self, req: UnifiedRequest) -> str:
        normalized = _normalize_messages(req.messages)
        parts: list[dict[str, str]] = []
        for message in normalized:
            text = _content_text(message.get("content"))
            if not text:
                continue
            role = _string(message.get("role")).lower() or "user"
            parts.append({"role": role, "text": text})
        if not parts:
            return "."
        last_user_index = -1
        for index in range(len(parts) - 1, -1, -1):
            if parts[index]["role"] == "user":
                last_user_index = index
                break
        output: list[str] = []
        for index, part in enumerate(parts):
            output.append(part["text"] if index == last_user_index else f"{part['role']}: {part['text']}")
        return "\n\n".join(output)

    def _collect_text(self, raw: str) -> str:
        filter_state = GrokStreamFilter()
        last_message = ""
        token_seen = False
        parts: list[str] = []
        for line in raw.splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                payload = json.loads(line)
            except json.JSONDecodeError:
                continue
            response = _nested_map(payload, "result", "response")
            if response is None:
                continue
            token = _text_value(response.get("token"))
            if token != "":
                token_seen = True
                filtered = _strip_grok_artifacts(filter_state.filter(token))
                if filtered:
                    parts.append(filtered)
                continue
            if token_seen:
                continue
            model_response = response.get("modelResponse")
            if not isinstance(model_response, dict):
                continue
            message = _text_value(model_response.get("message"))
            if message == "":
                continue
            filtered = _strip_grok_artifacts(message)
            delta = _normalize_incremental_chunk(filtered, last_message)
            if delta:
                last_message = filtered
                parts.append(delta)
        return "".join(parts).strip()


def _build_grok_cookie_header(token: str, cf_cookies: str = "", cf_clearance: str = "") -> str:
    trimmed = _string(token)
    if not trimmed:
        return ""
    base = trimmed if ";" in trimmed else f"sso={trimmed.removeprefix('sso=')}; sso-rw={trimmed.removeprefix('sso=')}"
    cookies = _parse_cookie_map(base)
    for name, value in _parse_cookie_map(cf_cookies).items():
        cookies[name] = value
    if _string(cf_clearance):
        cookies["cf_clearance"] = _string(cf_clearance)
    return "; ".join(f"{name}={value}" for name, value in cookies.items())


def _parse_cookie_map(raw: str) -> dict[str, str]:
    cookies: dict[str, str] = {}
    for chunk in raw.split(";"):
        name, sep, value = chunk.strip().partition("=")
        if not sep or not name.strip():
            continue
        cookies[name.strip()] = value.strip()
    return cookies


def _parse_retry_after_seconds(headers: Any) -> float:
    if headers is None:
        return 0.0
    value = str(headers.get("Retry-After", "")).strip()
    if not value:
        return 0.0
    try:
        return max(float(int(value)), 0.0)
    except ValueError:
        pass
    try:
        parsed = parsedate_to_datetime(value)
    except (TypeError, ValueError, IndexError):
        return 0.0
    if parsed.tzinfo is None:
        return 0.0
    return max((parsed - parsed.now(parsed.tzinfo)).total_seconds(), 0.0)


def _grok_retry_delay(last_delay: float, retry_after: float, status_code: int, attempt: int) -> float:
    if retry_after > 0:
        return retry_after
    if status_code == 429:
        return min(max(last_delay, GROK_RETRY_BACKOFF_BASE_SECONDS) * 2, GROK_RETRY_BACKOFF_MAX_SECONDS)
    return min(GROK_RETRY_BACKOFF_BASE_SECONDS * (2 ** min(attempt, 4)), GROK_RETRY_BACKOFF_MAX_SECONDS)


def _should_retry_grok_status(status_code: int) -> bool:
    return status_code in {403, 408, 429, 502, 503, 504} or status_code >= 500


def _should_retry_grok_transport_error(exc: URLError) -> bool:
    message = _string(getattr(exc, "reason", "") or exc)
    if not message:
        return True
    return "unsupported protocol scheme" not in message.lower()


def _grok_sec_fetch_site(origin: str, referer: str) -> str:
    origin_host = urlparse(origin).netloc.lower()
    referer_host = urlparse(referer).netloc.lower()
    if not origin_host or not referer_host:
        return "same-origin"
    return "same-origin" if origin_host == referer_host else "cross-site"


def _summarize_grok_tool_card(raw: str) -> str:
    name_match = _GROK_TOOL_NAME_RE.search(raw)
    args_match = _GROK_TOOL_ARGS_RE.search(raw)
    name = _GROK_CDATA_RE.sub(r"\1", name_match.group(1)).strip() if name_match else ""
    args = _GROK_CDATA_RE.sub(r"\1", args_match.group(1)).strip() if args_match else ""
    if not name and not args:
        return ""
    if not args:
        return f"[{name}]"
    return f"[{name}] {args}"


def _strip_grok_artifacts(text: str) -> str:
    if not text:
        return ""
    cleaned = _GROK_TOOL_USAGE_CARD_RE.sub(lambda match: _summarize_grok_tool_card(match.group(0)), text)
    cleaned = _GROK_ROLLOUT_RE.sub("", cleaned)
    cleaned = _GROK_SPECIAL_TAG_RE.sub("", cleaned)
    return cleaned