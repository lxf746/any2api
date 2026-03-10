import json
import unittest
from io import BytesIO
from urllib.error import HTTPError
from unittest.mock import patch

from gateway_types import UnifiedRequest
from providers import ChatGPTProvider, GrokProvider, KiroProvider, OrchidsProvider, WebProvider


class FakeResponse:
    def __init__(self, body: bytes, headers: dict[str, str] | None = None):
        self._body = body
        self.headers = headers or {}

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


def _encode_kiro_stream(event_type: str, payload: dict) -> bytes:
    payload_bytes = json.dumps(payload).encode("utf-8")
    name = b":event-type"
    value = event_type.encode("utf-8")
    headers = bytes([len(name)]) + name + bytes([7]) + len(value).to_bytes(2, "big") + value
    total_length = 12 + len(headers) + len(payload_bytes) + 4
    prelude = total_length.to_bytes(4, "big") + len(headers).to_bytes(4, "big") + b"\x00\x00\x00\x00"
    return prelude + headers + payload_bytes + b"\x00\x00\x00\x00"


class ProviderTests(unittest.TestCase):
    def test_kiro_generate_reply_parses_eventstream_and_uses_snapshot_account(self):
        provider = KiroProvider(lambda: {"providers": {"kiroAccounts": [{"accessToken": "test-token", "machineId": "machine-1", "active": True}]}})
        req = UnifiedRequest(messages=[{"role": "system", "content": "follow rules"}, {"role": "user", "content": "hi"}], model="claude-sonnet-4.6")

        def fake_urlopen(request, timeout=0):
            headers = {key.lower(): value for key, value in request.header_items()}
            body = json.loads(request.data.decode("utf-8"))
            self.assertEqual(headers["authorization"], "Bearer test-token")
            self.assertEqual(headers["x-amz-target"], "AmazonCodeWhispererStreamingService.GenerateAssistantResponse")
            self.assertEqual(body["conversationState"]["currentMessage"]["userInputMessage"]["content"], "follow rules\n\nhi")
            return FakeResponse(_encode_kiro_stream("assistantResponseEvent", {"content": "kiro ok"}))

        with patch("providers.kiro.provider.urlopen", side_effect=fake_urlopen):
            self.assertEqual(provider.generate_reply(req), "kiro ok")

    def test_grok_generate_reply_builds_cookie_and_parses_json_stream(self):
        provider = GrokProvider(lambda: {"providers": {
            "grokConfig": {"proxyUrl": "http://127.0.0.1:7890", "cfCookies": "theme=dark", "cfClearance": "cf-1", "origin": "https://grok.test", "referer": "https://grok.test/"},
            "grokTokens": [{"cookieToken": "test-token", "active": True}],
        }})
        req = UnifiedRequest(messages=[{"role": "system", "content": "be precise"}, {"role": "assistant", "content": "earlier reply"}, {"role": "user", "content": "latest question"}], model="grok-4")
        stream = "\n".join([
            json.dumps({"result": {"response": {"token": "hello "}}}),
            json.dumps({"result": {"response": {"token": "world"}}}),
        ]).encode("utf-8")

        def fake_urlopen(request, timeout=0):
            headers = {key.lower(): value for key, value in request.header_items()}
            body = json.loads(request.data.decode("utf-8"))
            self.assertIn("sso=test-token", headers["cookie"])
            self.assertIn("theme=dark", headers["cookie"])
            self.assertIn("cf_clearance=cf-1", headers["cookie"])
            self.assertEqual(headers["origin"], "https://grok.test")
            self.assertEqual(headers["referer"], "https://grok.test/")
            self.assertEqual(body["modelName"], "grok-4")
            self.assertEqual(body["message"], "system: be precise\n\nassistant: earlier reply\n\nlatest question")
            return FakeResponse(stream)

        with patch("providers.grok.provider.urlopen", side_effect=fake_urlopen), patch(
            "providers.grok.provider.build_opener", return_value=FakeOpener(fake_urlopen)
        ):
            self.assertEqual(provider.generate_reply(req), "hello world")

    def test_grok_generate_reply_retries_retry_after(self):
        provider = GrokProvider(lambda: {"providers": {"grokTokens": [{"cookieToken": "test-token", "active": True}]}})
        req = UnifiedRequest(messages=[{"role": "user", "content": "hi"}], model="grok-4")
        attempts = {"count": 0}
        slept: list[float] = []
        provider._sleep = lambda delay: slept.append(delay)

        def fake_urlopen(request, timeout=0):
            attempts["count"] += 1
            if attempts["count"] == 1:
                raise HTTPError(request.full_url, 429, "rate limited", {"Retry-After": "1"}, BytesIO(b"busy"))
            return FakeResponse(b'{"result":{"response":{"token":"retried"}}}\n')

        with patch("providers.grok.provider.urlopen", side_effect=fake_urlopen):
            self.assertEqual(provider.generate_reply(req), "retried")
        self.assertEqual(attempts["count"], 2)
        self.assertEqual(slept, [1.0])

    def test_orchids_generate_reply_fetches_clerk_token_and_collects_sse(self):
        provider = OrchidsProvider(lambda: {"providers": {"orchidsConfig": {"clientCookie": "client-cookie", "clientUat": "100", "projectId": "project-1", "agentMode": "claude-sonnet-4.5"}}})
        req = UnifiedRequest(messages=[{"role": "system", "content": "be precise"}, {"role": "user", "content": "write code"}], model="claude-sonnet-4.5")
        seen_urls: list[str] = []

        def fake_urlopen(request, timeout=0):
            seen_urls.append(request.full_url)
            headers = {key.lower(): value for key, value in request.header_items()}
            if request.get_method() == "GET":
                self.assertIn("/v1/client", request.full_url)
                self.assertEqual(headers["cookie"], "__client=client-cookie")
                return FakeResponse(json.dumps({"response": {"sessions": [{"user": {"id": "user-1", "email_addresses": [{"email_address": "user@example.com"}]}}], "last_active_session_id": "sess-1"}}).encode("utf-8"))
            if "/tokens" in request.full_url:
                self.assertIn("__client=client-cookie", headers["cookie"])
                self.assertIn("__client_uat=100", headers["cookie"])
                return FakeResponse(json.dumps({"jwt": "orchids-jwt"}).encode("utf-8"))
            body = json.loads(request.data.decode("utf-8"))
            self.assertEqual(headers["authorization"], "Bearer orchids-jwt")
            self.assertEqual(body["projectId"], "project-1")
            self.assertIn("<client_system>\nbe precise\n</client_system>", body["prompt"])
            self.assertIn("<user_request>\nwrite code\n</user_request>", body["prompt"])
            return FakeResponse(b'data: {"type":"model","event":{"type":"text-delta","delta":"hello "}}\n\ndata: {"type":"model","event":{"type":"text-delta","delta":"orchids"}}\n\n')

        with patch("providers.orchids.provider.urlopen", side_effect=fake_urlopen):
            self.assertEqual(provider.generate_reply(req), "hello orchids")
        self.assertEqual(len(seen_urls), 3)

    def test_web_generate_reply_posts_openai_shape_and_collects_json(self):
        provider = WebProvider(lambda: {"providers": {"webConfig": {"baseUrl": "https://web.test", "type": "openai", "apiKey": "web-key"}}})
        req = UnifiedRequest(messages=[{"role": "system", "content": "follow rules"}, {"role": "user", "content": "hi"}], model="")

        def fake_urlopen(request, timeout=0):
            headers = {key.lower(): value for key, value in request.header_items()}
            body = json.loads(request.data.decode("utf-8"))
            self.assertEqual(request.full_url, "https://web.test/openai/v1/chat/completions")
            self.assertEqual(headers["authorization"], "Bearer web-key")
            self.assertEqual(headers["accept"], "application/json")
            self.assertEqual(body["model"], "gpt-4.1")
            self.assertEqual(body["messages"][0], {"role": "system", "content": "follow rules"})
            self.assertEqual(body["messages"][1], {"role": "user", "content": "hi"})
            return FakeResponse(b'{"choices":[{"message":{"content":"web ok"}}]}', {"Content-Type": "application/json"})

        with patch("providers.web.provider.urlopen", side_effect=fake_urlopen):
            self.assertEqual(provider.generate_reply(req), "web ok")

    def test_chatgpt_generate_reply_collects_sse(self):
        provider = ChatGPTProvider(lambda: {"providers": {"chatgptConfig": {"baseUrl": "https://chatgpt.test", "token": "chatgpt-token"}}})
        req = UnifiedRequest(messages=[{"role": "user", "content": "hi"}], model="", stream=True)

        def fake_urlopen(request, timeout=0):
            headers = {key.lower(): value for key, value in request.header_items()}
            body = json.loads(request.data.decode("utf-8"))
            self.assertEqual(request.full_url, "https://chatgpt.test/v1/chat/completions")
            self.assertEqual(headers["authorization"], "Bearer chatgpt-token")
            self.assertEqual(headers["accept"], "text/event-stream")
            self.assertTrue(body["stream"])
            self.assertEqual(body["model"], "gpt-4.1")
            return FakeResponse(
                b'data: {"choices":[{"delta":{"content":"chat "}}]}\n\ndata: {"choices":[{"delta":{"content":"ok"}}]}\n\n',
                {"Content-Type": "text/event-stream"},
            )

        with patch("providers.chatgpt.provider.urlopen", side_effect=fake_urlopen):
            self.assertEqual(provider.generate_reply(req), "chat ok")


if __name__ == "__main__":
    unittest.main()