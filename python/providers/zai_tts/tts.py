"""Z.ai TTS API client — converts text to speech via SSE audio stream."""

from __future__ import annotations

import base64
import json
from dataclasses import dataclass, field
from typing import Any
from urllib.request import Request, urlopen


DEFAULT_ENDPOINT = "https://audio.z.ai/api/v1/z-audio/tts/create"
DEFAULT_AUTH_ENDPOINT = "https://audio.z.ai/api/v1/z-audio/auth/"
DEFAULT_TIMEOUT = 120


@dataclass
class AuthResponse:
    """Parsed auth API response."""
    code: int = 0
    message: str = ""
    user_id: str = ""
    auth_token: str = ""
    name: str = ""
    profile_image_url: str = ""
    timestamp: int = 0
    raw: dict[str, Any] = field(default_factory=dict)


class TTSClient:
    """Client for the Z.ai TTS API."""

    def __init__(self, token: str = "", user_id: str = "",
                 endpoint: str = DEFAULT_ENDPOINT,
                 auth_endpoint: str = DEFAULT_AUTH_ENDPOINT,
                 timeout: int = DEFAULT_TIMEOUT):
        self.token = token
        self.user_id = user_id
        self.endpoint = endpoint
        self.auth_endpoint = auth_endpoint
        self.timeout = timeout

    def authenticate(self, code: str) -> AuthResponse:
        """Exchange an OAuth code for a token. Auto-sets token and user_id."""
        payload = json.dumps({"code": code}).encode("utf-8")
        req = Request(self.auth_endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "*/*")
        req.add_header("Origin", "https://audio.z.ai")
        req.add_header("Referer", "https://audio.z.ai/")

        with urlopen(req, timeout=self.timeout) as resp:
            raw = json.loads(resp.read().decode("utf-8"))

        d = raw.get("data", {}) if isinstance(raw.get("data"), dict) else {}
        result = AuthResponse(
            code=raw.get("code", 0), message=raw.get("message", ""),
            user_id=d.get("user_id", ""), auth_token=d.get("auth_token", ""),
            name=d.get("name", ""), profile_image_url=d.get("profile_image_url", ""),
            timestamp=raw.get("timestamp", 0), raw=raw,
        )
        if result.auth_token:
            self.token = result.auth_token
        if result.user_id:
            self.user_id = result.user_id
        return result

    @classmethod
    def from_code(cls, code: str, **kwargs: Any) -> tuple[TTSClient, AuthResponse]:
        """Create a client by authenticating with an OAuth code."""
        client = cls(**kwargs)
        auth = client.authenticate(code)
        return client, auth

    def synthesize(self, text: str, voice_id: str = "system_003",
                   voice_name: str = "通用男声", speed: float = 1,
                   volume: float = 1) -> bytes:
        """Convert text to speech. Returns WAV audio bytes."""
        payload = json.dumps({
            "voice_name": voice_name, "voice_id": voice_id,
            "user_id": self.user_id, "input_text": text,
            "speed": speed, "volume": volume,
        }).encode("utf-8")

        req = Request(self.endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "text/event-stream")
        req.add_header("Authorization", f"Bearer {self.token}")
        req.add_header("Origin", "https://audio.z.ai")
        req.add_header("Referer", "https://audio.z.ai/")

        with urlopen(req, timeout=self.timeout) as resp:
            return self._read_sse_audio(resp)

    @staticmethod
    def _read_sse_audio(resp: Any) -> bytes:
        """Read SSE events and collect base64 audio chunks into raw bytes."""
        audio_data = bytearray()
        for raw_line in resp:
            line = raw_line.decode("utf-8", errors="replace").strip()
            if not line.startswith("data: "):
                continue
            data = line[6:]
            if data == "[DONE]":
                break
            try:
                chunk = json.loads(data)
            except json.JSONDecodeError:
                continue
            audio_b64 = chunk.get("audio", "")
            if audio_b64:
                audio_data.extend(base64.b64decode(audio_b64))
        return bytes(audio_data)
