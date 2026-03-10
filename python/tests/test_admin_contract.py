from __future__ import annotations

import http.client
import json
import os
import threading
import tempfile
import unittest
from unittest.mock import patch

from providers.zai_image.image import ImageInfo, ImageResponse
from providers.zai_ocr import OCRResponse
from server import create_server


class FakeResponse:
    def __init__(self, body: bytes):
        self._body = body

    def read(self) -> bytes:
        return self._body

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False


class FakeOpener:
    def __init__(self, handler):
        self._handler = handler

    def open(self, request, timeout=0):
        return self._handler(request, timeout=timeout)


class AdminContractTests(unittest.TestCase):
    def setUp(self):
        self.previous_password = os.environ.get("NEWPLATFORM2API_ADMIN_PASSWORD")
        self.previous_provider = os.environ.get("NEWPLATFORM2API_DEFAULT_PROVIDER")
        self.previous_store_path = os.environ.get("NEWPLATFORM2API_ADMIN_STORE_PATH")
        self.previous_api_key = os.environ.get("NEWPLATFORM2API_API_KEY")
        self.tempdir = tempfile.TemporaryDirectory()
        os.environ["NEWPLATFORM2API_ADMIN_PASSWORD"] = "changeme"
        os.environ["NEWPLATFORM2API_DEFAULT_PROVIDER"] = "cursor"
        os.environ["NEWPLATFORM2API_ADMIN_STORE_PATH"] = os.path.join(self.tempdir.name, "admin.json")
        os.environ["NEWPLATFORM2API_API_KEY"] = ""
        self._start_server()

    def tearDown(self):
        self._stop_server()
        self.tempdir.cleanup()
        self._restore_env("NEWPLATFORM2API_ADMIN_PASSWORD", self.previous_password)
        self._restore_env("NEWPLATFORM2API_DEFAULT_PROVIDER", self.previous_provider)
        self._restore_env("NEWPLATFORM2API_ADMIN_STORE_PATH", self.previous_store_path)
        self._restore_env("NEWPLATFORM2API_API_KEY", self.previous_api_key)

    def _restore_env(self, key: str, value: str | None) -> None:
        if value is None:
            os.environ.pop(key, None)
        else:
            os.environ[key] = value

    def _start_server(self):
        self.server = create_server("127.0.0.1", 0)
        self.port = self.server.server_address[1]
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def _stop_server(self):
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)

    def restart_server(self):
        self._stop_server()
        self._start_server()

    def request(self, method: str, path: str, body: dict | None = None, headers: dict[str, str] | None = None):
        conn = http.client.HTTPConnection("127.0.0.1", self.port, timeout=5)
        payload = json.dumps(body).encode("utf-8") if body is not None else None
        actual_headers = {**(headers or {})}
        if payload is not None:
            actual_headers.setdefault("Content-Type", "application/json")
        conn.request(method, path, body=payload, headers=actual_headers)
        response = conn.getresponse()
        raw = response.read().decode("utf-8")
        conn.close()
        parsed = json.loads(raw) if raw else None
        return response, parsed

    def request_raw(self, method: str, path: str, body: bytes | None = None, headers: dict[str, str] | None = None):
        conn = http.client.HTTPConnection("127.0.0.1", self.port, timeout=5)
        conn.request(method, path, body=body, headers=headers or {})
        response = conn.getresponse()
        raw = response.read()
        conn.close()
        return response, raw

    def multipart_body(self, fields: dict[str, str], files: dict[str, tuple[str, bytes, str]], boundary: str = "----Any2ApiTestBoundary") -> tuple[bytes, dict[str, str]]:
        chunks: list[bytes] = []
        for name, value in fields.items():
            chunks.extend([
                f"--{boundary}\r\n".encode("utf-8"),
                f'Content-Disposition: form-data; name="{name}"\r\n\r\n'.encode("utf-8"),
                str(value).encode("utf-8"),
                b"\r\n",
            ])
        for name, (filename, content, content_type) in files.items():
            chunks.extend([
                f"--{boundary}\r\n".encode("utf-8"),
                f'Content-Disposition: form-data; name="{name}"; filename="{filename}"\r\n'.encode("utf-8"),
                f"Content-Type: {content_type}\r\n\r\n".encode("utf-8"),
                content,
                b"\r\n",
            ])
        chunks.append(f"--{boundary}--\r\n".encode("utf-8"))
        body = b"".join(chunks)
        return body, {"Content-Type": f"multipart/form-data; boundary={boundary}"}

    def login(self, password: str = "changeme") -> str:
        response, body = self.request("POST", "/api/admin/auth/login", {"password": password})
        self.assertEqual(response.status, 200)
        self.assertTrue(body["ok"])
        self.assertTrue(body["token"])
        return body["token"]

    def test_meta_endpoint(self):
        response, body = self.request("GET", "/api/admin/meta", headers={"Origin": "http://localhost:1420"})
        self.assertEqual(response.status, 200)
        self.assertEqual(body["backend"]["language"], "python")
        self.assertEqual(body["auth"]["mode"], "session_cookie")
        self.assertTrue(body["features"]["providers"])
        self.assertTrue(body["features"]["credentials"])
        self.assertTrue(body["features"]["providerState"])
        self.assertFalse(body["features"]["configImportExport"])
        self.assertEqual(response.getheader("Access-Control-Allow-Origin"), "http://localhost:1420")

    def test_login_session_logout_lifecycle(self):
        bad_response, _ = self.request("POST", "/api/admin/auth/login", {"password": "wrong"})
        self.assertEqual(bad_response.status, 401)

        login_response, login_body = self.request("POST", "/api/admin/auth/login", {"password": "changeme"})
        self.assertEqual(login_response.status, 200)
        self.assertTrue(login_body["ok"])
        token = login_body["token"]
        self.assertTrue(token)
        self.assertIn("newplatform2api_admin_session=", login_response.getheader("Set-Cookie") or "")

        session_response, session_body = self.request(
            "GET",
            "/api/admin/auth/session",
            headers={"Authorization": f"Bearer {token}"},
        )
        self.assertEqual(session_response.status, 200)
        self.assertTrue(session_body["authenticated"])
        self.assertEqual(session_body["user"]["role"], "admin")

        logout_response, logout_body = self.request(
            "POST",
            "/api/admin/auth/logout",
            headers={"Authorization": f"Bearer {token}"},
        )
        self.assertEqual(logout_response.status, 200)
        self.assertTrue(logout_body["ok"])

        after_response, after_body = self.request(
            "GET",
            "/api/admin/auth/session",
            headers={"Authorization": f"Bearer {token}"},
        )
        self.assertEqual(after_response.status, 401)
        self.assertEqual(after_body["error"], "admin login required")

    def test_admin_api_requires_auth_and_supports_cors(self):
        unauth_response, unauth_body = self.request("GET", "/admin/api/settings")
        self.assertEqual(unauth_response.status, 401)
        self.assertEqual(unauth_body["error"], "admin login required")

        options_response, _ = self.request("OPTIONS", "/admin/api/settings", headers={"Origin": "http://localhost:1420"})
        self.assertEqual(options_response.status, 204)
        self.assertEqual(options_response.getheader("Access-Control-Allow-Origin"), "http://localhost:1420")
        self.assertIn("PUT", options_response.getheader("Access-Control-Allow-Methods") or "")
        self.assertIn("DELETE", options_response.getheader("Access-Control-Allow-Methods") or "")

    def test_settings_and_provider_roundtrip_persist_across_restart(self):
        token = self.login()
        auth = {"Authorization": f"Bearer {token}"}

        settings_response, settings_body = self.request("GET", "/admin/api/settings", headers=auth)
        self.assertEqual(settings_response.status, 200)
        self.assertEqual(settings_body["defaultProvider"], "cursor")
        self.assertTrue(settings_body["adminPasswordConfigured"])

        update_settings_response, update_settings_body = self.request(
            "PUT",
            "/admin/api/settings",
            {"apiKey": "sk-python", "defaultProvider": "grok", "adminPassword": "newpass"},
            headers=auth,
        )
        self.assertEqual(update_settings_response.status, 200)
        self.assertEqual(update_settings_body["apiKey"], "sk-python")
        self.assertEqual(update_settings_body["defaultProvider"], "grok")
        self.assertTrue(update_settings_body["adminPasswordConfigured"])

        cursor_response, cursor_body = self.request(
            "PUT",
            "/admin/api/providers/cursor/config",
            {"config": {"apiUrl": "https://cursor.test/chat", "cookie": "cursor-cookie"}},
            headers=auth,
        )
        self.assertEqual(cursor_response.status, 200)
        self.assertEqual(cursor_body["config"]["apiUrl"], "https://cursor.test/chat")
        self.assertEqual(cursor_body["config"]["cookie"], "cursor-cookie")

        kiro_primary_response, kiro_primary_body = self.request(
            "POST",
            "/admin/api/providers/kiro/accounts/create",
            {"name": "Primary", "accessToken": "ak-1", "machineId": "machine-1", "preferredEndpoint": "amazonq", "active": True},
            headers=auth,
        )
        self.assertEqual(kiro_primary_response.status, 200)
        kiro_primary = kiro_primary_body["account"]
        self.assertTrue(kiro_primary["id"])
        self.assertTrue(kiro_primary["active"])

        kiro_backup_response, kiro_backup_body = self.request(
            "POST",
            "/admin/api/providers/kiro/accounts/create",
            {"name": "Backup", "accessToken": "ak-2", "machineId": "machine-2", "preferredEndpoint": "codewhisperer", "active": False},
            headers=auth,
        )
        self.assertEqual(kiro_backup_response.status, 200)
        kiro_backup = kiro_backup_body["account"]

        kiro_detail_response, kiro_detail_body = self.request(
            "GET",
            f"/admin/api/providers/kiro/accounts/detail/{kiro_primary['id']}",
            headers=auth,
        )
        self.assertEqual(kiro_detail_response.status, 200)
        self.assertEqual(kiro_detail_body["account"]["id"], kiro_primary["id"])

        kiro_update_response, kiro_update_body = self.request(
            "PUT",
            f"/admin/api/providers/kiro/accounts/update/{kiro_backup['id']}",
            {"name": "Backup Updated", "accessToken": "ak-2b", "machineId": "machine-2b", "preferredEndpoint": "codewhisperer", "active": True},
            headers=auth,
        )
        self.assertEqual(kiro_update_response.status, 200)
        self.assertEqual(kiro_update_body["account"]["id"], kiro_backup["id"])
        self.assertEqual(kiro_update_body["account"]["name"], "Backup Updated")
        self.assertTrue(kiro_update_body["account"]["active"])

        kiro_delete_response, kiro_delete_body = self.request(
            "DELETE",
            f"/admin/api/providers/kiro/accounts/delete/{kiro_primary['id']}",
            headers=auth,
        )
        self.assertEqual(kiro_delete_response.status, 200)
        self.assertTrue(kiro_delete_body["ok"])

        kiro_list_response, kiro_list_body = self.request(
            "GET",
            "/admin/api/providers/kiro/accounts/list",
            headers=auth,
        )
        self.assertEqual(kiro_list_response.status, 200)
        self.assertEqual(len(kiro_list_body["accounts"]), 1)
        self.assertEqual(kiro_list_body["accounts"][0]["id"], kiro_backup["id"])
        self.assertTrue(kiro_list_body["accounts"][0]["active"])

        grok_primary_response, grok_primary_body = self.request(
            "POST",
            "/admin/api/providers/grok/tokens/create",
            {"name": "Primary", "cookieToken": "gt-1", "active": False},
            headers=auth,
        )
        self.assertEqual(grok_primary_response.status, 200)
        grok_primary = grok_primary_body["token"]

        grok_secondary_response, grok_secondary_body = self.request(
            "POST",
            "/admin/api/providers/grok/tokens/create",
            {"name": "Secondary", "cookieToken": "gt-2", "active": True},
            headers=auth,
        )
        self.assertEqual(grok_secondary_response.status, 200)
        grok_secondary = grok_secondary_body["token"]

        grok_detail_response, grok_detail_body = self.request(
            "GET",
            f"/admin/api/providers/grok/tokens/detail/{grok_primary['id']}",
            headers=auth,
        )
        self.assertEqual(grok_detail_response.status, 200)
        self.assertEqual(grok_detail_body["token"]["id"], grok_primary["id"])

        grok_update_response, grok_update_body = self.request(
            "PUT",
            f"/admin/api/providers/grok/tokens/update/{grok_secondary['id']}",
            {"name": "Secondary Updated", "cookieToken": "gt-2b", "active": True},
            headers=auth,
        )
        self.assertEqual(grok_update_response.status, 200)
        self.assertEqual(grok_update_body["token"]["id"], grok_secondary["id"])
        self.assertEqual(grok_update_body["token"]["name"], "Secondary Updated")
        self.assertTrue(grok_update_body["token"]["active"])

        grok_delete_response, grok_delete_body = self.request(
            "DELETE",
            f"/admin/api/providers/grok/tokens/delete/{grok_primary['id']}",
            headers=auth,
        )
        self.assertEqual(grok_delete_response.status, 200)
        self.assertTrue(grok_delete_body["ok"])

        grok_list_response, grok_list_body = self.request(
            "GET",
            "/admin/api/providers/grok/tokens/list",
            headers=auth,
        )
        self.assertEqual(grok_list_response.status, 200)
        self.assertEqual(len(grok_list_body["tokens"]), 1)
        self.assertEqual(grok_list_body["tokens"][0]["id"], grok_secondary["id"])
        self.assertTrue(grok_list_body["tokens"][0]["active"])

        grok_config_response, grok_config_body = self.request(
            "PUT",
            "/admin/api/providers/grok/config",
            {"config": {
                "apiUrl": "https://grok.test/chat",
                "proxyUrl": "http://127.0.0.1:7890",
                "cfCookies": "theme=dark",
                "cfClearance": "cf-token",
                "userAgent": "Mozilla/Test",
                "origin": "https://grok.test",
                "referer": "https://grok.test/",
            }},
            headers=auth,
        )
        self.assertEqual(grok_config_response.status, 200)
        self.assertEqual(grok_config_body["config"]["apiUrl"], "https://grok.test/chat")
        self.assertEqual(grok_config_body["config"]["proxyUrl"], "http://127.0.0.1:7890")
        self.assertEqual(grok_config_body["config"]["cfClearance"], "cf-token")

        orchids_response, orchids_body = self.request(
            "PUT",
            "/admin/api/providers/orchids/config",
            {"config": {"clientCookie": "orchids-cookie", "projectId": "project-1", "agentMode": "claude-sonnet-4.5"}},
            headers=auth,
        )
        self.assertEqual(orchids_response.status, 200)
        self.assertEqual(orchids_body["config"]["clientCookie"], "orchids-cookie")

        web_response, web_body = self.request(
            "PUT",
            "/admin/api/providers/web/config",
            {"config": {"baseUrl": "https://web.test", "type": "openai", "apiKey": "web-key"}},
            headers=auth,
        )
        self.assertEqual(web_response.status, 200)
        self.assertEqual(web_body["config"]["baseUrl"], "https://web.test")
        self.assertEqual(web_body["config"]["type"], "openai")
        self.assertEqual(web_body["config"]["apiKey"], "web-key")

        chatgpt_response, chatgpt_body = self.request(
            "PUT",
            "/admin/api/providers/chatgpt/config",
            {"config": {"baseUrl": "https://chatgpt.test", "token": "chatgpt-token"}},
            headers=auth,
        )
        self.assertEqual(chatgpt_response.status, 200)
        self.assertEqual(chatgpt_body["config"]["baseUrl"], "https://chatgpt.test")
        self.assertEqual(chatgpt_body["config"]["token"], "chatgpt-token")

        zai_image_response, zai_image_body = self.request(
            "PUT",
            "/admin/api/providers/zai/image/config",
            {"config": {"sessionToken": "zai-image-session"}},
            headers=auth,
        )
        self.assertEqual(zai_image_response.status, 200)
        self.assertEqual(zai_image_body["config"]["sessionToken"], "zai-image-session")

        zai_tts_response, zai_tts_body = self.request(
            "PUT",
            "/admin/api/providers/zai/tts/config",
            {"config": {"token": "zai-tts-token", "userId": "tts-user-1"}},
            headers=auth,
        )
        self.assertEqual(zai_tts_response.status, 200)
        self.assertEqual(zai_tts_body["config"]["token"], "zai-tts-token")
        self.assertEqual(zai_tts_body["config"]["userId"], "tts-user-1")

        zai_ocr_response, zai_ocr_body = self.request(
            "PUT",
            "/admin/api/providers/zai/ocr/config",
            {"config": {"token": "zai-ocr-token"}},
            headers=auth,
        )
        self.assertEqual(zai_ocr_response.status, 200)
        self.assertEqual(zai_ocr_body["config"]["token"], "zai-ocr-token")

        status_response, status_body = self.request("GET", "/admin/api/status", headers=auth)
        self.assertEqual(status_response.status, 200)
        self.assertEqual(status_body["project"], "any2api-python")
        self.assertEqual(status_body["settings"]["defaultProvider"], "grok")
        self.assertTrue(status_body["providers"]["cursor"]["configured"])
        self.assertEqual(status_body["providers"]["cursor"]["active"], "default")
        self.assertEqual(status_body["providers"]["kiro"]["count"], 1)
        self.assertTrue(status_body["providers"]["kiro"]["configured"])
        self.assertEqual(status_body["providers"]["kiro"]["active"], kiro_backup["id"])
        self.assertEqual(status_body["providers"]["grok"]["count"], 1)
        self.assertEqual(status_body["providers"]["grok"]["active"], grok_secondary["id"])
        self.assertTrue(status_body["providers"]["orchids"]["configured"])
        self.assertEqual(status_body["providers"]["web"]["count"], 1)
        self.assertTrue(status_body["providers"]["web"]["configured"])
        self.assertEqual(status_body["providers"]["web"]["active"], "default")
        self.assertEqual(status_body["providers"]["chatgpt"]["count"], 1)
        self.assertTrue(status_body["providers"]["chatgpt"]["configured"])
        self.assertEqual(status_body["providers"]["chatgpt"]["active"], "default")
        self.assertEqual(status_body["providers"]["zaiImage"]["count"], 1)
        self.assertTrue(status_body["providers"]["zaiImage"]["configured"])
        self.assertEqual(status_body["providers"]["zaiImage"]["active"], "default")
        self.assertEqual(status_body["providers"]["zaiTTS"]["count"], 1)
        self.assertTrue(status_body["providers"]["zaiTTS"]["configured"])
        self.assertEqual(status_body["providers"]["zaiTTS"]["active"], "default")
        self.assertEqual(status_body["providers"]["zaiOCR"]["count"], 1)
        self.assertTrue(status_body["providers"]["zaiOCR"]["configured"])
        self.assertEqual(status_body["providers"]["zaiOCR"]["active"], "default")

        self.restart_server()

        relogin_bad, _ = self.request("POST", "/api/admin/auth/login", {"password": "changeme"})
        self.assertEqual(relogin_bad.status, 401)
        persisted_token = self.login("newpass")
        persisted_auth = {"Authorization": f"Bearer {persisted_token}"}

        persisted_settings_response, persisted_settings_body = self.request("GET", "/admin/api/settings", headers=persisted_auth)
        self.assertEqual(persisted_settings_response.status, 200)
        self.assertEqual(persisted_settings_body["apiKey"], "sk-python")
        self.assertEqual(persisted_settings_body["defaultProvider"], "grok")

        persisted_kiro_response, persisted_kiro_body = self.request("GET", "/admin/api/providers/kiro/accounts/list", headers=persisted_auth)
        self.assertEqual(persisted_kiro_response.status, 200)
        self.assertEqual(len(persisted_kiro_body["accounts"]), 1)
        self.assertEqual(persisted_kiro_body["accounts"][0]["id"], kiro_backup["id"])

        persisted_grok_config_response, persisted_grok_config_body = self.request("GET", "/admin/api/providers/grok/config", headers=persisted_auth)
        self.assertEqual(persisted_grok_config_response.status, 200)
        self.assertEqual(persisted_grok_config_body["config"]["proxyUrl"], "http://127.0.0.1:7890")
        self.assertEqual(persisted_grok_config_body["config"]["cfClearance"], "cf-token")

        persisted_web_config_response, persisted_web_config_body = self.request("GET", "/admin/api/providers/web/config", headers=persisted_auth)
        self.assertEqual(persisted_web_config_response.status, 200)
        self.assertEqual(persisted_web_config_body["config"]["baseUrl"], "https://web.test")
        self.assertEqual(persisted_web_config_body["config"]["type"], "openai")

        persisted_chatgpt_config_response, persisted_chatgpt_config_body = self.request("GET", "/admin/api/providers/chatgpt/config", headers=persisted_auth)
        self.assertEqual(persisted_chatgpt_config_response.status, 200)
        self.assertEqual(persisted_chatgpt_config_body["config"]["baseUrl"], "https://chatgpt.test")
        self.assertEqual(persisted_chatgpt_config_body["config"]["token"], "chatgpt-token")

        persisted_zai_image_response, persisted_zai_image_body = self.request("GET", "/admin/api/providers/zai/image/config", headers=persisted_auth)
        self.assertEqual(persisted_zai_image_response.status, 200)
        self.assertEqual(persisted_zai_image_body["config"]["sessionToken"], "zai-image-session")

        persisted_zai_tts_response, persisted_zai_tts_body = self.request("GET", "/admin/api/providers/zai/tts/config", headers=persisted_auth)
        self.assertEqual(persisted_zai_tts_response.status, 200)
        self.assertEqual(persisted_zai_tts_body["config"]["token"], "zai-tts-token")
        self.assertEqual(persisted_zai_tts_body["config"]["userId"], "tts-user-1")

        persisted_zai_ocr_response, persisted_zai_ocr_body = self.request("GET", "/admin/api/providers/zai/ocr/config", headers=persisted_auth)
        self.assertEqual(persisted_zai_ocr_response.status, 200)
        self.assertEqual(persisted_zai_ocr_body["config"]["token"], "zai-ocr-token")

        grok_stream = b'{"result":{"response":{"token":"live grok"}}}\n'
        with patch("providers.grok.provider.urlopen", return_value=FakeResponse(grok_stream)), patch(
            "providers.grok.provider.build_opener",
            return_value=FakeOpener(lambda request, timeout=0: FakeResponse(grok_stream)),
        ):
            models_response, models_body = self.request(
                "POST",
                "/v1/chat/completions",
                {"messages": [{"role": "user", "content": "hi"}]},
                headers={"Authorization": "Bearer sk-python"},
            )
        self.assertEqual(models_response.status, 200)
        self.assertEqual(models_response.getheader("X-Newplatform2API-Provider"), "grok")
        self.assertEqual(models_body["choices"][0]["message"]["content"], "live grok")
        self.assertTrue(models_body["id"].startswith("chatcmpl_"))
        self.assertNotIn("skeleton", models_body["choices"][0]["message"]["content"].lower())

    def test_public_routes_require_api_key_when_configured(self):
        token = self.login()
        auth = {"Authorization": f"Bearer {token}"}
        update_response, _ = self.request(
            "PUT",
            "/admin/api/settings",
            {"apiKey": "sk-python", "defaultProvider": "cursor", "adminPassword": "changeme"},
            headers=auth,
        )
        self.assertEqual(update_response.status, 200)

        with patch.dict(os.environ, {"NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN": "image-session"}, clear=False), patch("server.ImageClient") as image_client_cls:
            image_client_cls.return_value.generate.return_value = ImageResponse(
                image=ImageInfo(image_url="https://img.test/cat.png", prompt="cat", width=1024, height=1024),
                timestamp=1710000000,
            )
            denied_response, denied_body = self.request("POST", "/v1/images/generations", {"prompt": "cat"})
            self.assertEqual(denied_response.status, 401)
            self.assertEqual(denied_body["error"]["type"], "authentication_error")
            image_client_cls.assert_not_called()

            allowed_response, allowed_body = self.request(
                "POST",
                "/v1/images/generations",
                {"prompt": "cat"},
                headers={"Authorization": "Bearer sk-python"},
            )
            self.assertEqual(allowed_response.status, 200)
            self.assertEqual(allowed_body["data"][0]["url"], "https://img.test/cat.png")

    def test_images_generation_route_maps_parameters(self):
        with patch.dict(os.environ, {"NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN": "image-session"}, clear=False), patch("server.ImageClient") as image_client_cls:
            image_client_cls.return_value.generate.return_value = ImageResponse(
                image=ImageInfo(
                    image_url="https://img.test/cat.png",
                    prompt="a cat by the window",
                    ratio="1:1",
                    resolution="1K",
                    width=1024,
                    height=1024,
                ),
                timestamp=1710000000,
            )
            response, body = self.request("POST", "/v1/images/generations", {
                "model": "glm-image",
                "prompt": "a cat by the window",
                "size": "1024x1024",
                "provider_options": {"rm_label_watermark": False},
            })
        self.assertEqual(response.status, 200)
        self.assertEqual(response.getheader("X-Newplatform2API-Provider"), "zai_image")
        self.assertEqual(body["created"], 1710000000)
        self.assertEqual(body["data"][0]["url"], "https://img.test/cat.png")
        self.assertEqual(body["data"][0]["width"], 1024)
        self.assertEqual(body["data"][0]["ratio"], "1:1")
        image_client_cls.assert_called_once_with(session_token="image-session")
        image_client_cls.return_value.generate.assert_called_once_with(
            "a cat by the window",
            ratio="1:1",
            resolution="1K",
            rm_label_watermark=False,
        )

    def test_zai_admin_config_overrides_environment_for_public_routes(self):
        token = self.login()
        auth = {"Authorization": f"Bearer {token}"}

        self.request(
            "PUT",
            "/admin/api/providers/zai/image/config",
            {"config": {"sessionToken": "stored-image-session"}},
            headers=auth,
        )
        self.request(
            "PUT",
            "/admin/api/providers/zai/tts/config",
            {"config": {"token": "stored-tts-token", "userId": "stored-user"}},
            headers=auth,
        )
        self.request(
            "PUT",
            "/admin/api/providers/zai/ocr/config",
            {"config": {"token": "stored-ocr-token"}},
            headers=auth,
        )

        with patch.dict(os.environ, {
            "NEWPLATFORM2API_ZAI_IMAGE_SESSION_TOKEN": "env-image-session",
            "NEWPLATFORM2API_ZAI_TTS_TOKEN": "env-tts-token",
            "NEWPLATFORM2API_ZAI_TTS_USER_ID": "env-user",
            "NEWPLATFORM2API_ZAI_OCR_TOKEN": "env-ocr-token",
        }, clear=False), patch("server.ImageClient") as image_client_cls, patch("server.TTSClient") as tts_client_cls, patch("server.OCRClient") as ocr_client_cls:
            image_client_cls.return_value.generate.return_value = ImageResponse(
                image=ImageInfo(image_url="https://img.test/admin.png", prompt="cat"),
                timestamp=1710000001,
            )
            tts_client_cls.return_value.synthesize.return_value = b"RIFFadmin"
            ocr_client_cls.return_value.process_bytes.return_value = OCRResponse(
                task_id="task-admin",
                status="succeeded",
                file_name="admin.txt",
                file_size=5,
                file_type="text/plain",
                file_url="https://files.test/admin.txt",
                created_at="2025-01-01T00:00:00Z",
                markdown_content="hello",
                json_content={"text": "hello"},
                layout=[],
            )

            image_response, _ = self.request("POST", "/v1/images/generations", {"prompt": "cat"})
            self.assertEqual(image_response.status, 200)

            tts_response, _ = self.request_raw(
                "POST",
                "/v1/audio/speech",
                body=json.dumps({"input": "hello"}).encode("utf-8"),
                headers={"Content-Type": "application/json"},
            )
            self.assertEqual(tts_response.status, 200)

            body, headers = self.multipart_body({}, {"file": ("admin.txt", b"hello", "text/plain")})
            ocr_response, _ = self.request_raw("POST", "/v1/ocr", body=body, headers=headers)
            self.assertEqual(ocr_response.status, 200)

        image_client_cls.assert_called_once_with(session_token="stored-image-session")
        tts_client_cls.assert_called_once_with(token="stored-tts-token", user_id="stored-user")
        ocr_client_cls.assert_called_once_with(token="stored-ocr-token")

    def test_audio_speech_route_returns_binary_wav(self):
        audio_bytes = b"RIFFtest-wave"
        with patch.dict(os.environ, {
            "NEWPLATFORM2API_ZAI_TTS_TOKEN": "tts-token",
            "NEWPLATFORM2API_ZAI_TTS_USER_ID": "user-123",
        }, clear=False), patch("server.TTSClient") as tts_client_cls:
            tts_client_cls.return_value.synthesize.return_value = audio_bytes
            response, raw = self.request_raw(
                "POST",
                "/v1/audio/speech",
                body=json.dumps({
                    "model": "tts-1",
                    "input": "hello world",
                    "voice": "system_003",
                    "speed": 1.25,
                    "provider_options": {"volume": 0.75},
                }).encode("utf-8"),
                headers={"Content-Type": "application/json"},
            )
        self.assertEqual(response.status, 200)
        self.assertEqual(response.getheader("Content-Type"), "audio/wav")
        self.assertEqual(response.getheader("X-Newplatform2API-Provider"), "zai_tts")
        self.assertEqual(raw, audio_bytes)
        tts_client_cls.assert_called_once_with(token="tts-token", user_id="user-123")
        tts_client_cls.return_value.synthesize.assert_called_once_with(
            "hello world",
            voice_id="system_003",
            voice_name="通用男声",
            speed=1.25,
            volume=0.75,
        )

    def test_ocr_route_accepts_multipart_upload(self):
        with patch.dict(os.environ, {"NEWPLATFORM2API_ZAI_OCR_TOKEN": "ocr-token"}, clear=False), patch("server.OCRClient") as ocr_client_cls:
            ocr_client_cls.return_value.process_bytes.return_value = OCRResponse(
                task_id="task-1",
                status="succeeded",
                file_name="note.txt",
                file_size=5,
                file_type="text/plain",
                file_url="https://files.test/note.txt",
                created_at="2025-01-01T00:00:00Z",
                markdown_content="hello",
                json_content={"text": "hello"},
                layout=[{"type": "paragraph"}],
            )
            body, headers = self.multipart_body({}, {"file": ("note.txt", b"hello", "text/plain")})
            response, raw = self.request_raw("POST", "/v1/ocr", body=body, headers=headers)
        parsed = json.loads(raw.decode("utf-8"))
        self.assertEqual(response.status, 200)
        self.assertEqual(response.getheader("X-Newplatform2API-Provider"), "zai_ocr")
        self.assertEqual(parsed["id"], "task-1")
        self.assertEqual(parsed["text"], "hello")
        self.assertEqual(parsed["json"]["text"], "hello")
        self.assertEqual(parsed["file"]["name"], "note.txt")
        ocr_client_cls.assert_called_once_with(token="ocr-token")
        ocr_client_cls.return_value.process_bytes.assert_called_once_with(b"hello", "note.txt")


if __name__ == "__main__":
    unittest.main()