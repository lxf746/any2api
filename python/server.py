from __future__ import annotations

import json
import os
import secrets
import threading
from email.parser import BytesParser
from email.policy import default
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from datetime import datetime, timedelta, timezone
from urllib.parse import parse_qs, unquote, urlparse

from admin_store import AdminStore, default_admin_store_path
from gateway_types import UnifiedRequest
from providers import ProviderRegistry, ProviderRequestError, default_registry
from providers.zai_image import ImageClient
from providers.zai_ocr import OCRClient
from providers.zai_tts import TTSClient


ADMIN_SESSION_COOKIE = "newplatform2api_admin_session"
ADMIN_AUTH_MODE = "session_cookie"
ADMIN_BACKEND_VERSION = "0.1.0"


class AdminSessionStore:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._sessions: dict[str, datetime] = {}

    def create(self) -> tuple[str, datetime]:
        token = secrets.token_hex(24)
        expires_at = datetime.now(timezone.utc) + timedelta(hours=24)
        with self._lock:
            self._sessions[token] = expires_at
        return token, expires_at

    def expires_at(self, token: str) -> datetime | None:
        token = token.strip()
        if not token:
            return None
        with self._lock:
            expires_at = self._sessions.get(token)
            if expires_at is None:
                return None
            if datetime.now(timezone.utc) >= expires_at:
                self._sessions.pop(token, None)
                return None
            return expires_at

    def delete(self, token: str) -> None:
        token = token.strip()
        if not token:
            return
        with self._lock:
            self._sessions.pop(token, None)


def shared_admin_features() -> dict[str, bool]:
    return {
        "providers": True,
        "credentials": True,
        "providerState": True,
        "stats": False,
        "logs": False,
        "users": False,
        "configImportExport": False,
    }


def current_admin_password() -> str:
    return os.getenv("NEWPLATFORM2API_ADMIN_PASSWORD", "changeme").strip() or "changeme"


def current_api_key() -> str:
    return os.getenv("NEWPLATFORM2API_API_KEY", "").strip()


def current_zai_image_session_token() -> str:
    return env_string("", "NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN", "ZAI_IMAGE_SESSION_TOKEN")


def current_zai_tts_token() -> str:
    return env_string("", "NEWPLATFORM2API_ZAI_TTS_TOKEN", "ZAI_TTS_TOKEN")


def current_zai_tts_user_id() -> str:
    return env_string("", "NEWPLATFORM2API_ZAI_TTS_USER_ID", "ZAI_TTS_USER_ID")


def current_zai_ocr_token() -> str:
    return env_string("", "NEWPLATFORM2API_ZAI_OCR_TOKEN", "ZAI_OCR_TOKEN")


IMAGE_SIZE_MAP = {
    "1024x1024": ("1:1", "1K"),
    "1024x1792": ("9:16", "2K"),
    "1792x1024": ("16:9", "2K"),
}


def env_string(default: str, *keys: str) -> str:
    for key in keys:
        value = os.getenv(key, "").strip()
        if value:
            return value
    return default


def env_int(default: int, *keys: str) -> int:
    for key in keys:
        value = os.getenv(key, "").strip()
        if not value:
            continue
        try:
            return int(value)
        except ValueError:
            continue
    return default


def current_server_host() -> str:
    return env_string("127.0.0.1", "NEWPLATFORM2API_HOST", "HOST")


def current_server_port() -> int:
    return env_int(8100, "NEWPLATFORM2API_PORT", "PORT")


class AppHandler(BaseHTTPRequestHandler):
    registry: ProviderRegistry = default_registry()
    sessions = AdminSessionStore()
    store: AdminStore | None = None

    def log_message(self, format: str, *args: object) -> None:  # noqa: A003
        return

    def do_OPTIONS(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if self._is_admin_api_path(parsed.path):
            return self._empty(204, self._admin_cors_headers())
        self._empty(404)

    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path == "/health":
            return self._json(200, {"status": "ok", "project": "any2api-python"})
        if parsed.path == "/api/admin/meta":
            return self._admin_meta()
        if parsed.path == "/api/admin/auth/session":
            return self._admin_session()
        if parsed.path == "/admin/api/status":
            if not self._require_admin():
                return
            return self._admin_status()
        if parsed.path == "/admin/api/settings":
            if not self._require_admin():
                return
            return self._admin_settings()
        if parsed.path == "/admin/api/providers/cursor/config":
            if not self._require_admin():
                return
            return self._admin_cursor_config()
        if parsed.path in {"/admin/api/providers/kiro/accounts", "/admin/api/providers/kiro/accounts/list"}:
            if not self._require_admin():
                return
            return self._admin_kiro_accounts()
        if parsed.path.startswith("/admin/api/providers/kiro/accounts/detail/"):
            if not self._require_admin():
                return
            account_id = self._admin_path_id(parsed.path, "/admin/api/providers/kiro/accounts/detail/")
            if account_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_kiro_account(account_id)
        if parsed.path == "/admin/api/providers/grok/config":
            if not self._require_admin():
                return
            return self._admin_grok_config()
        if parsed.path in {"/admin/api/providers/grok/tokens", "/admin/api/providers/grok/tokens/list"}:
            if not self._require_admin():
                return
            return self._admin_grok_tokens()
        if parsed.path.startswith("/admin/api/providers/grok/tokens/detail/"):
            if not self._require_admin():
                return
            token_id = self._admin_path_id(parsed.path, "/admin/api/providers/grok/tokens/detail/")
            if token_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_grok_token(token_id)
        if parsed.path == "/admin/api/providers/orchids/config":
            if not self._require_admin():
                return
            return self._admin_orchids_config()
        if parsed.path == "/admin/api/providers/web/config":
            if not self._require_admin():
                return
            return self._admin_web_config()
        if parsed.path == "/admin/api/providers/chatgpt/config":
            if not self._require_admin():
                return
            return self._admin_chatgpt_config()
        if parsed.path == "/admin/api/providers/zai/image/config":
            if not self._require_admin():
                return
            return self._admin_zai_image_config()
        if parsed.path == "/admin/api/providers/zai/tts/config":
            if not self._require_admin():
                return
            return self._admin_zai_tts_config()
        if parsed.path == "/admin/api/providers/zai/ocr/config":
            if not self._require_admin():
                return
            return self._admin_zai_ocr_config()
        if parsed.path == "/v1/models":
            if not self._require_api_key():
                return
            provider = parse_qs(parsed.query).get("provider", [None])[0]
            models = self.registry.models(provider)
            data = [{"id": item["public_model"], "object": "model", "owned_by": item["owned_by"], "provider": item["provider"], "upstream_model": item["upstream_model"]} for item in models]
            return self._json(200, {"object": "list", "data": data})
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path == "/api/admin/auth/logout":
            return self._admin_logout()

        if parsed.path == "/admin/api/providers/kiro/accounts/create":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_create_kiro_account(payload)

        if parsed.path == "/admin/api/providers/grok/tokens/create":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_create_grok_token(payload)

        if parsed.path == "/v1/ocr":
            if not self._require_api_key():
                return
            form = self._read_multipart_form()
            if form is None:
                return
            fields, files = form
            return self._ocr_upload(fields, files)

        if parsed.path not in {
            "/api/admin/auth/login",
            "/v1/chat/completions",
            "/v1/messages",
            "/v1/images/generations",
            "/v1/audio/speech",
        }:
            return self._json(404, {"error": "not found"})

        payload = self._read_json_object()
        if payload is None:
            return

        if parsed.path == "/api/admin/auth/login":
            return self._admin_login(payload)

        if not self._require_api_key():
            return

        if parsed.path == "/v1/chat/completions":
            return self._openai_chat(payload)
        if parsed.path == "/v1/messages":
            return self._anthropic_messages(payload)
        if parsed.path == "/v1/images/generations":
            return self._openai_images_generation(payload)
        if parsed.path == "/v1/audio/speech":
            return self._openai_audio_speech(payload)
        self._json(404, {"error": "not found"})

    def do_PUT(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path.startswith("/admin/api/providers/kiro/accounts/update/"):
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            account_id = self._admin_path_id(parsed.path, "/admin/api/providers/kiro/accounts/update/")
            if account_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_update_kiro_account(account_id, payload)

        if parsed.path.startswith("/admin/api/providers/grok/tokens/update/"):
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            token_id = self._admin_path_id(parsed.path, "/admin/api/providers/grok/tokens/update/")
            if token_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_update_grok_token(token_id, payload)

        if parsed.path == "/admin/api/settings":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_settings(payload)
        if parsed.path == "/admin/api/providers/cursor/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_cursor_config(payload)
        if parsed.path == "/admin/api/providers/kiro/accounts":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_kiro_accounts(payload)
        if parsed.path == "/admin/api/providers/grok/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_grok_config(payload)
        if parsed.path == "/admin/api/providers/grok/tokens":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_grok_tokens(payload)
        if parsed.path == "/admin/api/providers/orchids/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_orchids_config(payload)
        if parsed.path == "/admin/api/providers/web/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_web_config(payload)
        if parsed.path == "/admin/api/providers/chatgpt/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_chatgpt_config(payload)
        if parsed.path == "/admin/api/providers/zai/image/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_zai_image_config(payload)
        if parsed.path == "/admin/api/providers/zai/tts/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_zai_tts_config(payload)
        if parsed.path == "/admin/api/providers/zai/ocr/config":
            if not self._require_admin():
                return
            payload = self._read_json_object()
            if payload is None:
                return
            return self._admin_zai_ocr_config(payload)
        self._json(404, {"error": "not found"})

    def do_DELETE(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        if parsed.path.startswith("/admin/api/providers/kiro/accounts/delete/"):
            if not self._require_admin():
                return
            account_id = self._admin_path_id(parsed.path, "/admin/api/providers/kiro/accounts/delete/")
            if account_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_delete_kiro_account(account_id)
        if parsed.path.startswith("/admin/api/providers/grok/tokens/delete/"):
            if not self._require_admin():
                return
            token_id = self._admin_path_id(parsed.path, "/admin/api/providers/grok/tokens/delete/")
            if token_id is None:
                return self._json(404, {"error": "not found"})
            return self._admin_delete_grok_token(token_id)
        self._json(404, {"error": "not found"})

    def _admin_meta(self) -> None:
        self._json(200, {
            "backend": {"language": "python", "version": ADMIN_BACKEND_VERSION},
            "auth": {"mode": ADMIN_AUTH_MODE},
            "features": shared_admin_features(),
        })

    def _admin_login(self, payload: dict) -> None:
        if str(payload.get("password", "")).strip() != self._store().admin_password():
            return self._json(401, {"error": "invalid admin password"})
        token, _ = self.sessions.create()
        self._json(200, {"ok": True, "token": token}, extra_headers={
            "Set-Cookie": f"{ADMIN_SESSION_COOKIE}={token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400",
        })

    def _admin_session(self) -> None:
        token = self._admin_session_token()
        expires_at = self.sessions.expires_at(token)
        if expires_at is None:
            return self._json(401, {"error": "admin login required"})
        self._json(200, {
            "authenticated": True,
            "user": {"id": "local-admin", "name": "Admin", "role": "admin"},
            "expiresAt": expires_at.isoformat().replace("+00:00", "Z"),
        })

    def _admin_logout(self) -> None:
        self.sessions.delete(self._admin_session_token())
        self._json(200, {"ok": True}, extra_headers={
            "Set-Cookie": f"{ADMIN_SESSION_COOKIE}=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0",
        })

    def _admin_status(self) -> None:
        data = self._store().snapshot()
        settings = self._admin_settings_body(data)
        providers = data["providers"]
        cursor_config = providers["cursorConfig"]
        orchids_config = providers["orchidsConfig"]
        web_config = providers["webConfig"]
        chatgpt_config = providers["chatgptConfig"]
        kiro_accounts = providers["kiroAccounts"]
        grok_tokens = providers["grokTokens"]
        self._json(200, {
            "project": "any2api-python",
            "settings": settings,
            "providers": {
                "cursor": self._single_provider_status(bool(cursor_config.get("cookie"))),
                "kiro": self._multi_provider_status(kiro_accounts, "accessToken"),
                "grok": self._multi_provider_status(grok_tokens, "cookieToken"),
                "orchids": self._single_provider_status(bool(orchids_config.get("clientCookie"))),
                "web": self._single_provider_status(bool(web_config.get("baseUrl") and web_config.get("type"))),
                "chatgpt": self._single_provider_status(bool(chatgpt_config.get("baseUrl") and chatgpt_config.get("token"))),
                "zaiImage": self._single_provider_status(bool(self._zai_image_session_token())),
                "zaiTTS": self._single_provider_status(bool(self._zai_tts_token() and self._zai_tts_user_id())),
                "zaiOCR": self._single_provider_status(bool(self._zai_ocr_token())),
            },
        })

    def _admin_settings(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, self._admin_settings_body(self._store().snapshot()))
        data = self._store().update_settings(
            payload.get("apiKey", ""),
            payload.get("defaultProvider", ""),
            payload.get("adminPassword", ""),
        )
        self.registry.default_provider = data["settings"]["defaultProvider"]
        self._json(200, self._admin_settings_body(data))

    def _admin_cursor_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["cursorConfig"]})
        data = self._store().replace_cursor_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["cursorConfig"]})

    def _admin_kiro_accounts(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"accounts": self._store().snapshot()["providers"]["kiroAccounts"]})
        data = self._store().replace_kiro_accounts(payload.get("accounts", []))
        self._json(200, {"accounts": data["providers"]["kiroAccounts"]})

    def _admin_kiro_account(self, account_id: str) -> None:
        account = self._store().kiro_account(account_id)
        if account is None:
            return self._json(404, {"error": "not found"})
        self._json(200, {"account": account})

    def _admin_create_kiro_account(self, payload: dict) -> None:
        try:
            account = self._store().create_kiro_account(payload)
        except ValueError as exc:
            return self._json(400, {"error": str(exc)})
        self._json(200, {"account": account})

    def _admin_update_kiro_account(self, account_id: str, payload: dict) -> None:
        try:
            account = self._store().update_kiro_account(account_id, payload)
        except ValueError as exc:
            return self._json(400, {"error": str(exc)})
        if account is None:
            return self._json(404, {"error": "not found"})
        self._json(200, {"account": account})

    def _admin_delete_kiro_account(self, account_id: str) -> None:
        if not self._store().delete_kiro_account(account_id):
            return self._json(404, {"error": "not found"})
        self._json(200, {"ok": True})

    def _admin_grok_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["grokConfig"]})
        data = self._store().replace_grok_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["grokConfig"]})

    def _admin_grok_tokens(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"tokens": self._store().snapshot()["providers"]["grokTokens"]})
        data = self._store().replace_grok_tokens(payload.get("tokens", []))
        self._json(200, {"tokens": data["providers"]["grokTokens"]})

    def _admin_grok_token(self, token_id: str) -> None:
        token = self._store().grok_token(token_id)
        if token is None:
            return self._json(404, {"error": "not found"})
        self._json(200, {"token": token})

    def _admin_create_grok_token(self, payload: dict) -> None:
        try:
            token = self._store().create_grok_token(payload)
        except ValueError as exc:
            return self._json(400, {"error": str(exc)})
        self._json(200, {"token": token})

    def _admin_update_grok_token(self, token_id: str, payload: dict) -> None:
        try:
            token = self._store().update_grok_token(token_id, payload)
        except ValueError as exc:
            return self._json(400, {"error": str(exc)})
        if token is None:
            return self._json(404, {"error": "not found"})
        self._json(200, {"token": token})

    def _admin_delete_grok_token(self, token_id: str) -> None:
        if not self._store().delete_grok_token(token_id):
            return self._json(404, {"error": "not found"})
        self._json(200, {"ok": True})

    def _admin_orchids_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["orchidsConfig"]})
        data = self._store().replace_orchids_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["orchidsConfig"]})

    def _admin_web_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["webConfig"]})
        data = self._store().replace_web_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["webConfig"]})

    def _admin_chatgpt_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["chatgptConfig"]})
        data = self._store().replace_chatgpt_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["chatgptConfig"]})

    def _admin_zai_image_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["zaiImageConfig"]})
        data = self._store().replace_zai_image_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["zaiImageConfig"]})

    def _admin_zai_tts_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["zaiTTSConfig"]})
        data = self._store().replace_zai_tts_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["zaiTTSConfig"]})

    def _admin_zai_ocr_config(self, payload: dict | None = None) -> None:
        if payload is None:
            return self._json(200, {"config": self._store().snapshot()["providers"]["zaiOCRConfig"]})
        data = self._store().replace_zai_ocr_config(payload.get("config", {}))
        self._json(200, {"config": data["providers"]["zaiOCRConfig"]})

    def _admin_session_token(self) -> str:
        authorization = self.headers.get("Authorization", "").strip()
        if authorization.startswith("Bearer "):
            token = authorization.removeprefix("Bearer ").strip()
            if token:
                return token
        cookie_header = self.headers.get("Cookie", "")
        for chunk in cookie_header.split(";"):
            name, _, value = chunk.strip().partition("=")
            if name == ADMIN_SESSION_COOKIE and value.strip():
                return value.strip()
        return ""

    def _admin_cors_headers(self) -> dict[str, str]:
        origin = self.headers.get("Origin", "*") or "*"
        headers = {
            "Access-Control-Allow-Origin": origin,
            "Access-Control-Allow-Headers": "Content-Type, Authorization",
            "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
        }
        if origin != "*":
            headers["Access-Control-Allow-Credentials"] = "true"
        return headers

    def _admin_path_id(self, path: str, prefix: str) -> str | None:
        raw = path.removeprefix(prefix).strip()
        if not raw or "/" in raw:
            return None
        decoded = unquote(raw).strip()
        return decoded or None

    def _require_admin(self) -> bool:
        token = self._admin_session_token()
        if self.sessions.expires_at(token) is None:
            self._json(401, {"error": "admin login required"})
            return False
        return True

    def _bearer_token(self) -> str:
        authorization = self.headers.get("Authorization", "").strip()
        if authorization.startswith("Bearer "):
            return authorization.removeprefix("Bearer ").strip()
        return ""

    def _require_api_key(self) -> bool:
        api_key = str(self._store().snapshot()["settings"].get("apiKey", "")).strip()
        if not api_key:
            return True
        if self._bearer_token() == api_key:
            return True
        self._json(401, {"error": {"message": "Missing or invalid authorization", "type": "authentication_error"}})
        return False

    def _is_admin_api_path(self, path: str) -> bool:
        return path.startswith("/api/admin/") or path.startswith("/admin/api/")

    def _read_request_body(self) -> bytes:
        length = int(self.headers.get("Content-Length", "0"))
        return self.rfile.read(length) if length else b""

    def _read_json_object(self) -> dict | None:
        raw = self._read_request_body() or b"{}"
        try:
            payload = json.loads(raw or b"{}")
        except json.JSONDecodeError:
            self._json(400, {"error": "invalid json"})
            return None
        if not isinstance(payload, dict):
            self._json(400, {"error": "invalid json"})
            return None
        return payload

    def _read_multipart_form(self) -> tuple[dict[str, str], dict[str, dict[str, object]]] | None:
        content_type = self.headers.get("Content-Type", "")
        if "multipart/form-data" not in content_type:
            self._json(400, {"error": "content-type must be multipart/form-data"})
            return None
        raw = self._read_request_body()
        message = BytesParser(policy=default).parsebytes(
            f"Content-Type: {content_type}\r\nMIME-Version: 1.0\r\n\r\n".encode("utf-8") + raw
        )
        if not message.is_multipart():
            self._json(400, {"error": "invalid multipart form-data"})
            return None
        fields: dict[str, str] = {}
        files: dict[str, dict[str, object]] = {}
        for part in message.iter_parts():
            name = part.get_param("name", header="content-disposition")
            if not name:
                continue
            payload = part.get_payload(decode=True) or b""
            filename = part.get_filename()
            if filename is None:
                charset = part.get_content_charset() or "utf-8"
                fields[name] = payload.decode(charset, errors="replace")
                continue
            files[name] = {
                "filename": filename,
                "content": payload,
                "content_type": part.get_content_type(),
            }
        return fields, files

    def _store(self) -> AdminStore:
        if self.store is None:
            raise RuntimeError("admin store not initialized")
        return self.store

    def _admin_settings_body(self, data: dict) -> dict[str, object]:
        settings = data["settings"]
        return {
            "apiKey": settings["apiKey"],
            "defaultProvider": settings["defaultProvider"],
            "adminPasswordConfigured": bool(str(settings.get("adminPassword", "")).strip()),
        }

    def _single_provider_status(self, configured: bool) -> dict[str, object]:
        return {"count": 1 if configured else 0, "configured": configured, "active": "default" if configured else ""}

    def _multi_provider_status(self, items: list[dict], secret_field: str) -> dict[str, object]:
        active = ""
        configured = False
        for item in items:
            if str(item.get(secret_field, "")).strip():
                configured = True
            if item.get("active"):
                active = str(item.get("id", "")).strip()
        return {"count": len(items), "configured": configured, "active": active}

    def _provider_options(self, payload: dict) -> dict:
        options = payload.get("provider_options", {})
        if options is None:
            return {}
        if not isinstance(options, dict):
            self._json(400, {"error": "provider_options must be an object"})
            return {}
        return options

    def _coerce_bool(self, value: object, default: bool) -> bool:
        if value is None:
            return default
        if isinstance(value, bool):
            return value
        if isinstance(value, str):
            normalized = value.strip().lower()
            if normalized in {"1", "true", "yes", "on"}:
                return True
            if normalized in {"0", "false", "no", "off"}:
                return False
        return bool(value)

    def _coerce_float(self, value: object, default: float) -> float:
        if value is None or value == "":
            return default
        return float(value)

    def _zai_image_session_token(self) -> str:
        config = self._store().snapshot()["providers"]["zaiImageConfig"]
        return str(config.get("sessionToken", "")).strip() or current_zai_image_session_token()

    def _zai_tts_token(self) -> str:
        config = self._store().snapshot()["providers"]["zaiTTSConfig"]
        return str(config.get("token", "")).strip() or current_zai_tts_token()

    def _zai_tts_user_id(self) -> str:
        config = self._store().snapshot()["providers"]["zaiTTSConfig"]
        return str(config.get("userId", "")).strip() or current_zai_tts_user_id()

    def _zai_ocr_token(self) -> str:
        config = self._store().snapshot()["providers"]["zaiOCRConfig"]
        return str(config.get("token", "")).strip() or current_zai_ocr_token()

    def _zai_image_client(self) -> ImageClient | None:
        token = self._zai_image_session_token()
        if not token:
            self._json(503, {"error": "zai image is not configured"})
            return None
        return ImageClient(session_token=token)

    def _zai_tts_client(self) -> TTSClient | None:
        token = self._zai_tts_token()
        user_id = self._zai_tts_user_id()
        if not token or not user_id:
            self._json(503, {"error": "zai tts is not configured"})
            return None
        return TTSClient(token=token, user_id=user_id)

    def _zai_ocr_client(self) -> OCRClient | None:
        token = self._zai_ocr_token()
        if not token:
            self._json(503, {"error": "zai ocr is not configured"})
            return None
        return OCRClient(token=token)

    def _resolve_image_settings(self, payload: dict, options: dict) -> tuple[str, str] | None:
        ratio = str(payload.get("ratio") or options.get("ratio") or "").strip()
        resolution = str(payload.get("resolution") or options.get("resolution") or "").strip()
        if ratio and resolution:
            return ratio, resolution
        size = str(payload.get("size") or options.get("size") or "").strip()
        if size:
            mapped = IMAGE_SIZE_MAP.get(size)
            if mapped is None:
                self._json(400, {"error": f"unsupported size: {size}"})
                return None
            return mapped
        return ratio or "1:1", resolution or "1K"

    def _openai_images_generation(self, payload: dict) -> None:
        prompt = str(payload.get("prompt", "")).strip()
        if not prompt:
            return self._json(400, {"error": "prompt is required"})
        try:
            n = int(payload.get("n", 1) or 1)
        except (TypeError, ValueError):
            return self._json(400, {"error": "n must be an integer"})
        if n != 1:
            return self._json(400, {"error": "only n=1 is supported"})
        response_format = str(payload.get("response_format", "url") or "url").strip().lower()
        if response_format not in {"", "url"}:
            return self._json(400, {"error": "only response_format=url is supported"})
        options = self._provider_options(payload)
        if payload.get("provider_options") is not None and not isinstance(payload.get("provider_options"), dict):
            return
        image_settings = self._resolve_image_settings(payload, options)
        if image_settings is None:
            return
        ratio, resolution = image_settings
        rm_label_watermark = self._coerce_bool(
            payload.get("rm_label_watermark", options.get("rm_label_watermark")),
            True,
        )
        client = self._zai_image_client()
        if client is None:
            return
        try:
            result = client.generate(
                prompt,
                ratio=ratio,
                resolution=resolution,
                rm_label_watermark=rm_label_watermark,
            )
        except Exception as exc:
            return self._json(502, {"error": str(exc)}, extra_headers={"X-Newplatform2API-Provider": "zai_image"})
        created = result.timestamp or int(datetime.now(timezone.utc).timestamp())
        image = result.image
        size = image.size or (f"{image.width}x{image.height}" if image.width and image.height else "")
        self._json(200, {
            "created": created,
            "data": [{
                "url": image.image_url,
                "revised_prompt": image.prompt or prompt,
                "size": size,
                "width": image.width,
                "height": image.height,
                "ratio": image.ratio or ratio,
                "resolution": image.resolution or resolution,
            }],
        }, extra_headers={"X-Newplatform2API-Provider": "zai_image"})

    def _openai_audio_speech(self, payload: dict) -> None:
        text = str(payload.get("input") or payload.get("text") or "").strip()
        if not text:
            return self._json(400, {"error": "input is required"})
        response_format = str(payload.get("response_format", "wav") or "wav").strip().lower()
        if response_format != "wav":
            return self._json(400, {"error": "only response_format=wav is supported"})
        options = self._provider_options(payload)
        if payload.get("provider_options") is not None and not isinstance(payload.get("provider_options"), dict):
            return
        voice_id = str(payload.get("voice_id") or payload.get("voice") or options.get("voice_id") or "system_003").strip() or "system_003"
        voice_name = str(payload.get("voice_name") or options.get("voice_name") or "通用男声").strip() or "通用男声"
        try:
            speed = self._coerce_float(payload.get("speed", options.get("speed")), 1.0)
            volume = self._coerce_float(payload.get("volume", options.get("volume")), 1.0)
        except (TypeError, ValueError):
            return self._json(400, {"error": "speed and volume must be numbers"})
        client = self._zai_tts_client()
        if client is None:
            return
        try:
            audio = client.synthesize(text, voice_id=voice_id, voice_name=voice_name, speed=speed, volume=volume)
        except Exception as exc:
            return self._json(502, {"error": str(exc)}, extra_headers={"X-Newplatform2API-Provider": "zai_tts"})
        self._bytes(200, audio, "audio/wav", extra_headers={"X-Newplatform2API-Provider": "zai_tts"})

    def _ocr_upload(self, fields: dict[str, str], files: dict[str, dict[str, object]]) -> None:
        del fields
        upload = files.get("file")
        if upload is None:
            return self._json(400, {"error": "file is required"})
        filename = str(upload.get("filename") or "upload.bin")
        content = upload.get("content")
        if not isinstance(content, (bytes, bytearray)):
            return self._json(400, {"error": "invalid file content"})
        client = self._zai_ocr_client()
        if client is None:
            return
        try:
            result = client.process_bytes(bytes(content), filename)
        except Exception as exc:
            return self._json(502, {"error": str(exc)}, extra_headers={"X-Newplatform2API-Provider": "zai_ocr"})
        self._json(200, {
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
            },
        }, extra_headers={"X-Newplatform2API-Provider": "zai_ocr"})

    def _openai_chat(self, payload: dict) -> None:
        provider = self._resolve(payload.get("provider"))
        if provider is None:
            return
        req = UnifiedRequest(
            provider_hint=payload.get("provider", ""),
            protocol="openai",
            model=payload.get("model", ""),
            messages=payload.get("messages", []),
            tools=payload.get("tools", []),
            stream=bool(payload.get("stream")),
        )
        text = self._generate_reply(provider, req)
        if text is None:
            return
        completion_id = f"chatcmpl_{secrets.token_hex(8)}"
        if req.stream:
            return self._sse_openai(text, provider.provider_id(), completion_id)
        self._json(200, {"id": completion_id, "object": "chat.completion", "model": req.model, "choices": [{"index": 0, "message": {"role": "assistant", "content": text}, "finish_reason": "stop"}]}, extra_headers={"X-Newplatform2API-Provider": provider.provider_id()})

    def _anthropic_messages(self, payload: dict) -> None:
        provider = self._resolve(payload.get("provider"))
        if provider is None:
            return
        req = UnifiedRequest(
            provider_hint=payload.get("provider", ""),
            protocol="anthropic",
            model=payload.get("model", ""),
            messages=payload.get("messages", []),
            system=payload.get("system"),
            tools=payload.get("tools", []),
            stream=bool(payload.get("stream")),
        )
        text = self._generate_reply(provider, req)
        if text is None:
            return
        message_id = f"msg_{secrets.token_hex(8)}"
        if req.stream:
            return self._sse_anthropic(text, provider.provider_id(), message_id)
        self._json(200, {"id": message_id, "type": "message", "role": "assistant", "model": req.model, "content": [{"type": "text", "text": text}], "stop_reason": "end_turn"}, extra_headers={"X-Newplatform2API-Provider": provider.provider_id()})

    def _generate_reply(self, provider, req: UnifiedRequest) -> str | None:
        try:
            return provider.generate_reply(req)
        except ProviderRequestError as exc:
            self._json(502, {"error": str(exc)}, extra_headers={"X-Newplatform2API-Provider": provider.provider_id()})
            return None

    def _resolve(self, provider_id: str | None):
        try:
            return self.registry.resolve(provider_id)
        except KeyError as exc:
            self._json(400, {"error": str(exc)})
            return None

    def _json(self, status: int, body: dict, extra_headers: dict[str, str] | None = None) -> None:
        encoded = json.dumps(body).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(encoded)))
        headers = dict(extra_headers or {})
        if self._is_admin_api_path(urlparse(self.path).path):
            headers = {**self._admin_cors_headers(), **headers}
        for key, value in headers.items():
            self.send_header(key, value)
        self.end_headers()
        self.wfile.write(encoded)

    def _bytes(self, status: int, body: bytes, content_type: str, extra_headers: dict[str, str] | None = None) -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        for key, value in (extra_headers or {}).items():
            self.send_header(key, value)
        self.end_headers()
        self.wfile.write(body)

    def _empty(self, status: int, extra_headers: dict[str, str] | None = None) -> None:
        self.send_response(status)
        headers = extra_headers or {}
        for key, value in headers.items():
            self.send_header(key, value)
        self.send_header("Content-Length", "0")
        self.end_headers()

    def _sse_openai(self, text: str, provider_id: str, completion_id: str) -> None:
        body = f'data: {{"id":"{completion_id}","object":"chat.completion.chunk","choices":[{{"index":0,"delta":{{"role":"assistant","content":{json.dumps(text)}}},"finish_reason":null}}]}}\n\ndata: [DONE]\n\n'.encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("X-Newplatform2API-Provider", provider_id)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _sse_anthropic(self, text: str, provider_id: str, message_id: str) -> None:
        chunks = (
            f"event: message_start\ndata: {{\"type\":\"message_start\",\"message\":{{\"id\":\"{message_id}\"}}}}\n\n"
            f"event: content_block_delta\ndata: {{\"type\":\"content_block_delta\",\"delta\":{{\"type\":\"text_delta\",\"text\":{json.dumps(text)}}}}}\n\n"
            "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
        ).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("X-Newplatform2API-Provider", provider_id)
        self.send_header("Content-Length", str(len(chunks)))
        self.end_headers()
        self.wfile.write(chunks)


def create_server(host: str | None = None, port: int | None = None) -> ThreadingHTTPServer:
    bind_host = current_server_host() if host is None else host
    bind_port = current_server_port() if port is None else port
    store = AdminStore(
        default_admin_store_path(),
        os.getenv("NEWPLATFORM2API_DEFAULT_PROVIDER", "cursor"),
        current_admin_password(),
        current_api_key(),
    )
    AppHandler.store = store
    AppHandler.registry = default_registry(store.snapshot()["settings"]["defaultProvider"], snapshot_getter=store.snapshot)
    AppHandler.sessions = AdminSessionStore()
    return ThreadingHTTPServer((bind_host, bind_port), AppHandler)


if __name__ == "__main__":
    host = current_server_host()
    port = current_server_port()
    server = create_server(host, port)
    print(f"any2api-python listening on http://{host}:{port}")
    server.serve_forever()
