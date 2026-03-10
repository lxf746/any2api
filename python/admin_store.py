from __future__ import annotations

import json
import os
import threading
import time
from copy import deepcopy
from pathlib import Path
from typing import Any


CURSOR_FIELDS = ("apiUrl", "scriptUrl", "cookie", "xIsHuman", "userAgent", "referer", "webglVendor", "webglRenderer")
GROK_FIELDS = ("apiUrl", "proxyUrl", "cfCookies", "cfClearance", "userAgent", "origin", "referer")
ORCHIDS_FIELDS = ("apiUrl", "clerkUrl", "clientCookie", "clientUat", "sessionId", "projectId", "userId", "email", "agentMode")
WEB_FIELDS = ("baseUrl", "type", "apiKey")
CHATGPT_FIELDS = ("baseUrl", "token")
ZAI_IMAGE_FIELDS = ("sessionToken",)
ZAI_TTS_FIELDS = ("token", "userId")
ZAI_OCR_FIELDS = ("token",)


def default_admin_store_path() -> Path:
    override = os.getenv("NEWPLATFORM2API_ADMIN_STORE_PATH", "").strip()
    if override:
        return Path(override).expanduser()
    return Path(__file__).resolve().parent / "data" / "admin.json"


def _string(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def _normalize_object(value: Any, fields: tuple[str, ...]) -> dict[str, str]:
    raw = value if isinstance(value, dict) else {}
    return {field: _string(raw.get(field)) for field in fields}


def _next_item_id(prefix: str, existing_ids: set[str]) -> str:
    while True:
        candidate = f"{prefix}-{time.time_ns()}-{len(existing_ids)}"
        if candidate not in existing_ids:
            return candidate


def _find_item_index(items: list[dict[str, Any]], item_id: str) -> int:
    target = _string(item_id)
    if not target:
        return -1
    for index, item in enumerate(items):
        if _string(item.get("id")) == target:
            return index
    return -1


def _default_grok_config() -> dict[str, str]:
    return {
        "apiUrl": os.getenv("NEWPLATFORM2API_GROK_API_URL", "https://grok.com/rest/app-chat/conversations/new").strip() or "https://grok.com/rest/app-chat/conversations/new",
        "proxyUrl": os.getenv("NEWPLATFORM2API_GROK_PROXY_URL", "").strip(),
        "cfCookies": os.getenv("NEWPLATFORM2API_GROK_CF_COOKIES", "").strip(),
        "cfClearance": os.getenv("NEWPLATFORM2API_GROK_CF_CLEARANCE", "").strip(),
        "userAgent": os.getenv("NEWPLATFORM2API_GROK_USER_AGENT", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").strip() or "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
        "origin": os.getenv("NEWPLATFORM2API_GROK_ORIGIN", "https://grok.com").strip() or "https://grok.com",
        "referer": os.getenv("NEWPLATFORM2API_GROK_REFERER", "https://grok.com/").strip() or "https://grok.com/",
    }


def _normalize_grok_config(value: Any) -> dict[str, str]:
    raw = value if isinstance(value, dict) else {}
    defaults = _default_grok_config()
    return {field: _string(raw.get(field, defaults[field])) for field in GROK_FIELDS}


def _default_web_config() -> dict[str, str]:
    return {
        "baseUrl": os.getenv("NEWPLATFORM2API_WEB_BASE_URL", "http://127.0.0.1:9000").strip() or "http://127.0.0.1:9000",
        "type": os.getenv("NEWPLATFORM2API_WEB_TYPE", "claude").strip().strip("/") or "claude",
        "apiKey": os.getenv("NEWPLATFORM2API_WEB_API_KEY", "").strip(),
    }


def _normalize_web_config(value: Any) -> dict[str, str]:
    raw = value if isinstance(value, dict) else {}
    defaults = _default_web_config()
    return {
        "baseUrl": _string(raw.get("baseUrl", defaults["baseUrl"])),
        "type": _string(raw.get("type", defaults["type"])).strip("/") or defaults["type"],
        "apiKey": _string(raw.get("apiKey", defaults["apiKey"])),
    }


def _default_chatgpt_config() -> dict[str, str]:
    return {
        "baseUrl": os.getenv("NEWPLATFORM2API_CHATGPT_BASE_URL", "http://127.0.0.1:5005").strip() or "http://127.0.0.1:5005",
        "token": os.getenv("NEWPLATFORM2API_CHATGPT_TOKEN", "").strip(),
    }


def _normalize_chatgpt_config(value: Any) -> dict[str, str]:
    raw = value if isinstance(value, dict) else {}
    defaults = _default_chatgpt_config()
    return {
        "baseUrl": _string(raw.get("baseUrl", defaults["baseUrl"])),
        "token": _string(raw.get("token", defaults["token"])),
    }


def _normalize_kiro_accounts(value: Any) -> list[dict[str, Any]]:
    raw_items = value if isinstance(value, list) else []
    items: list[dict[str, Any]] = []
    existing_ids: set[str] = set()
    active_set = False
    for item in raw_items:
        if not isinstance(item, dict):
            continue
        access_token = _string(item.get("accessToken"))
        machine_id = _string(item.get("machineId"))
        if not access_token and not machine_id:
            continue
        item_id = _string(item.get("id"))
        if not item_id or item_id in existing_ids:
            item_id = _next_item_id("kiro", existing_ids)
        existing_ids.add(item_id)
        normalized = {
            "id": item_id,
            "name": _string(item.get("name")) or f"Kiro Account {len(items) + 1}",
            "accessToken": access_token,
            "machineId": machine_id,
            "preferredEndpoint": _string(item.get("preferredEndpoint")).lower(),
            "active": bool(item.get("active")) and not active_set,
        }
        if normalized["active"]:
            active_set = True
        items.append(normalized)
    if items and not active_set:
        items[0]["active"] = True
    return items


def _normalize_grok_tokens(value: Any) -> list[dict[str, Any]]:
    raw_items = value if isinstance(value, list) else []
    items: list[dict[str, Any]] = []
    existing_ids: set[str] = set()
    active_set = False
    for item in raw_items:
        if not isinstance(item, dict):
            continue
        cookie_token = _string(item.get("cookieToken"))
        if not cookie_token:
            continue
        item_id = _string(item.get("id"))
        if not item_id or item_id in existing_ids:
            item_id = _next_item_id("grok", existing_ids)
        existing_ids.add(item_id)
        normalized = {
            "id": item_id,
            "name": _string(item.get("name")) or f"Grok Token {len(items) + 1}",
            "cookieToken": cookie_token,
            "active": bool(item.get("active")) and not active_set,
        }
        if normalized["active"]:
            active_set = True
        items.append(normalized)
    if items and not active_set:
        items[0]["active"] = True
    return items


def _normalize_data(raw: Any, default_provider: str, admin_password: str, api_key: str) -> dict[str, Any]:
    source = raw if isinstance(raw, dict) else {}
    settings = source.get("settings") if isinstance(source.get("settings"), dict) else {}
    providers = source.get("providers") if isinstance(source.get("providers"), dict) else {}
    normalized_password = _string(settings.get("adminPassword")) or admin_password
    normalized_provider = _string(settings.get("defaultProvider")) or default_provider
    return {
        "settings": {
            "adminPassword": normalized_password,
            "apiKey": _string(settings.get("apiKey")) or api_key,
            "defaultProvider": normalized_provider,
        },
        "providers": {
            "cursorConfig": _normalize_object(providers.get("cursorConfig"), CURSOR_FIELDS),
            "kiroAccounts": _normalize_kiro_accounts(providers.get("kiroAccounts")),
            "grokConfig": _normalize_grok_config(providers.get("grokConfig")),
            "grokTokens": _normalize_grok_tokens(providers.get("grokTokens")),
            "orchidsConfig": _normalize_object(providers.get("orchidsConfig"), ORCHIDS_FIELDS),
            "webConfig": _normalize_web_config(providers.get("webConfig")),
            "chatgptConfig": _normalize_chatgpt_config(providers.get("chatgptConfig")),
            "zaiImageConfig": _normalize_object(providers.get("zaiImageConfig"), ZAI_IMAGE_FIELDS),
            "zaiTTSConfig": _normalize_object(providers.get("zaiTTSConfig"), ZAI_TTS_FIELDS),
            "zaiOCRConfig": _normalize_object(providers.get("zaiOCRConfig"), ZAI_OCR_FIELDS),
        },
    }


class AdminStore:
    def __init__(self, path: str | Path, default_provider: str, admin_password: str, api_key: str = "") -> None:
        self._path = Path(path)
        self._lock = threading.RLock()
        self._defaults = {
            "default_provider": _string(default_provider) or "cursor",
            "admin_password": _string(admin_password) or "changeme",
            "api_key": _string(api_key),
        }
        self._data = _normalize_data({}, self._defaults["default_provider"], self._defaults["admin_password"], self._defaults["api_key"])
        self._load()

    def snapshot(self) -> dict[str, Any]:
        with self._lock:
            return deepcopy(self._data)

    def admin_password(self) -> str:
        with self._lock:
            return _string(self._data["settings"].get("adminPassword")) or self._defaults["admin_password"]

    def update_settings(self, api_key: str, default_provider: str, admin_password: str) -> dict[str, Any]:
        with self._lock:
            self._data["settings"]["apiKey"] = _string(api_key)
            self._data["settings"]["defaultProvider"] = _string(default_provider) or self._defaults["default_provider"]
            if _string(admin_password):
                self._data["settings"]["adminPassword"] = _string(admin_password)
            self._persist_locked()
            return deepcopy(self._data)

    def replace_cursor_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("cursorConfig", _normalize_object(config, CURSOR_FIELDS))

    def replace_kiro_accounts(self, accounts: Any) -> dict[str, Any]:
        return self._replace_provider("kiroAccounts", _normalize_kiro_accounts(accounts))

    def kiro_account(self, account_id: str) -> dict[str, Any] | None:
        with self._lock:
            items = self._data["providers"]["kiroAccounts"]
            index = _find_item_index(items, account_id)
            if index < 0:
                return None
            return deepcopy(items[index])

    def create_kiro_account(self, account: Any) -> dict[str, Any]:
        candidate = dict(account) if isinstance(account, dict) else {}
        with self._lock:
            items = deepcopy(self._data["providers"]["kiroAccounts"])
            if _find_item_index(items, candidate.get("id", "")) >= 0:
                candidate["id"] = ""
            prepared = _normalize_kiro_accounts([candidate])
            if not prepared:
                raise ValueError("invalid kiro account")
            if prepared[0].get("active"):
                for item in items:
                    item["active"] = False
            items.append(prepared[0])
            items = _normalize_kiro_accounts(items)
            index = _find_item_index(items, prepared[0]["id"])
            if index < 0:
                raise ValueError("invalid kiro account")
            self._data["providers"]["kiroAccounts"] = items
            self._persist_locked()
            return deepcopy(items[index])

    def update_kiro_account(self, account_id: str, account: Any) -> dict[str, Any] | None:
        candidate = dict(account) if isinstance(account, dict) else {}
        candidate["id"] = _string(account_id)
        with self._lock:
            items = deepcopy(self._data["providers"]["kiroAccounts"])
            index = _find_item_index(items, account_id)
            if index < 0:
                return None
            prepared = _normalize_kiro_accounts([candidate])
            if not prepared:
                raise ValueError("invalid kiro account")
            if prepared[0].get("active"):
                for item in items:
                    item["active"] = False
            items[index] = prepared[0]
            items = _normalize_kiro_accounts(items)
            updated_index = _find_item_index(items, account_id)
            if updated_index < 0:
                raise ValueError("invalid kiro account")
            self._data["providers"]["kiroAccounts"] = items
            self._persist_locked()
            return deepcopy(items[updated_index])

    def delete_kiro_account(self, account_id: str) -> bool:
        with self._lock:
            items = deepcopy(self._data["providers"]["kiroAccounts"])
            index = _find_item_index(items, account_id)
            if index < 0:
                return False
            del items[index]
            self._data["providers"]["kiroAccounts"] = _normalize_kiro_accounts(items)
            self._persist_locked()
            return True

    def replace_grok_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("grokConfig", _normalize_grok_config(config))

    def replace_grok_tokens(self, tokens: Any) -> dict[str, Any]:
        return self._replace_provider("grokTokens", _normalize_grok_tokens(tokens))

    def grok_token(self, token_id: str) -> dict[str, Any] | None:
        with self._lock:
            items = self._data["providers"]["grokTokens"]
            index = _find_item_index(items, token_id)
            if index < 0:
                return None
            return deepcopy(items[index])

    def create_grok_token(self, token: Any) -> dict[str, Any]:
        candidate = dict(token) if isinstance(token, dict) else {}
        with self._lock:
            items = deepcopy(self._data["providers"]["grokTokens"])
            if _find_item_index(items, candidate.get("id", "")) >= 0:
                candidate["id"] = ""
            prepared = _normalize_grok_tokens([candidate])
            if not prepared:
                raise ValueError("invalid grok token")
            if prepared[0].get("active"):
                for item in items:
                    item["active"] = False
            items.append(prepared[0])
            items = _normalize_grok_tokens(items)
            index = _find_item_index(items, prepared[0]["id"])
            if index < 0:
                raise ValueError("invalid grok token")
            self._data["providers"]["grokTokens"] = items
            self._persist_locked()
            return deepcopy(items[index])

    def update_grok_token(self, token_id: str, token: Any) -> dict[str, Any] | None:
        candidate = dict(token) if isinstance(token, dict) else {}
        candidate["id"] = _string(token_id)
        with self._lock:
            items = deepcopy(self._data["providers"]["grokTokens"])
            index = _find_item_index(items, token_id)
            if index < 0:
                return None
            prepared = _normalize_grok_tokens([candidate])
            if not prepared:
                raise ValueError("invalid grok token")
            if prepared[0].get("active"):
                for item in items:
                    item["active"] = False
            items[index] = prepared[0]
            items = _normalize_grok_tokens(items)
            updated_index = _find_item_index(items, token_id)
            if updated_index < 0:
                raise ValueError("invalid grok token")
            self._data["providers"]["grokTokens"] = items
            self._persist_locked()
            return deepcopy(items[updated_index])

    def delete_grok_token(self, token_id: str) -> bool:
        with self._lock:
            items = deepcopy(self._data["providers"]["grokTokens"])
            index = _find_item_index(items, token_id)
            if index < 0:
                return False
            del items[index]
            self._data["providers"]["grokTokens"] = _normalize_grok_tokens(items)
            self._persist_locked()
            return True

    def replace_orchids_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("orchidsConfig", _normalize_object(config, ORCHIDS_FIELDS))

    def replace_web_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("webConfig", _normalize_web_config(config))

    def replace_chatgpt_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("chatgptConfig", _normalize_chatgpt_config(config))

    def replace_zai_image_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("zaiImageConfig", _normalize_object(config, ZAI_IMAGE_FIELDS))

    def replace_zai_tts_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("zaiTTSConfig", _normalize_object(config, ZAI_TTS_FIELDS))

    def replace_zai_ocr_config(self, config: Any) -> dict[str, Any]:
        return self._replace_provider("zaiOCRConfig", _normalize_object(config, ZAI_OCR_FIELDS))

    def _replace_provider(self, key: str, value: Any) -> dict[str, Any]:
        with self._lock:
            self._data["providers"][key] = value
            self._persist_locked()
            return deepcopy(self._data)

    def _load(self) -> None:
        with self._lock:
            if not self._path.exists() or not self._path.read_text(encoding="utf-8").strip():
                self._persist_locked()
                return
            raw = json.loads(self._path.read_text(encoding="utf-8"))
            self._data = _normalize_data(raw, self._defaults["default_provider"], self._defaults["admin_password"], self._defaults["api_key"])

    def _persist_locked(self) -> None:
        self._path.parent.mkdir(parents=True, exist_ok=True)
        content = json.dumps(self._data, indent=2)
        self._path.write_text(f"{content}\n", encoding="utf-8")
