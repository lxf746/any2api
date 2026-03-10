"""Z.ai Image Generation API client."""

from __future__ import annotations

import json
import random
import string
from dataclasses import dataclass, field
from typing import Any
from urllib.request import Request, urlopen


DEFAULT_ENDPOINT = "https://image.z.ai/api/proxy/images/generate"
DEFAULT_AUTH_ENDPOINT = "https://image.z.ai/api/v1/z-image/auth/"
DEFAULT_CALLBACK_ENDPOINT = "https://image.z.ai/api/auth/callback"
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


@dataclass
class ImageInfo:
    """Generated image details."""
    image_id: str = ""
    prompt: str = ""
    size: str = ""
    ratio: str = ""
    resolution: str = ""
    image_url: str = ""
    status: str = ""
    created_at: str = ""
    updated_at: str = ""
    width: int = 0
    height: int = 0


@dataclass
class ImageResponse:
    """Parsed image generation API response."""
    code: int = 0
    message: str = ""
    image: ImageInfo = field(default_factory=ImageInfo)
    timestamp: int = 0
    raw: dict[str, Any] = field(default_factory=dict)


class ImageClient:
    """Client for the Z.ai Image Generation API."""

    def __init__(self, session_token: str = "", endpoint: str = DEFAULT_ENDPOINT,
                 auth_endpoint: str = DEFAULT_AUTH_ENDPOINT,
                 callback_endpoint: str = DEFAULT_CALLBACK_ENDPOINT,
                 timeout: int = DEFAULT_TIMEOUT):
        self.session_token = session_token
        self.endpoint = endpoint
        self.auth_endpoint = auth_endpoint
        self.callback_endpoint = callback_endpoint
        self.timeout = timeout

    def authenticate(self, code: str) -> AuthResponse:
        """Full auth flow: code → token → session cookie. Auto-sets session_token."""
        # Step 1: exchange code for token
        payload = json.dumps({"code": code}).encode("utf-8")
        req = Request(self.auth_endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "*/*")
        req.add_header("X-Request-ID", _random_id(22))
        req.add_header("Origin", "https://image.z.ai")
        req.add_header("Referer", "https://image.z.ai/")

        with urlopen(req, timeout=self.timeout) as resp:
            raw = json.loads(resp.read().decode("utf-8"))

        d = raw.get("data", {}) if isinstance(raw.get("data"), dict) else {}
        result = AuthResponse(
            code=raw.get("code", 0),
            message=raw.get("message", ""),
            user_id=d.get("user_id", ""),
            auth_token=d.get("auth_token", ""),
            name=d.get("name", ""),
            profile_image_url=d.get("profile_image_url", ""),
            timestamp=raw.get("timestamp", 0),
            raw=raw,
        )
        if not result.auth_token:
            raise RuntimeError("auth returned empty token")

        # Step 2: register token as session cookie
        self._register_callback(result.auth_token)
        self.session_token = result.auth_token
        return result

    def _register_callback(self, token: str) -> None:
        """POST /api/auth/callback to register the session."""
        payload = json.dumps({"token": token}).encode("utf-8")
        req = Request(self.callback_endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "*/*")
        req.add_header("Origin", "https://image.z.ai")
        req.add_header("Referer", "https://image.z.ai/")
        with urlopen(req, timeout=self.timeout) as resp:
            resp.read()

    @classmethod
    def from_code(cls, code: str, **kwargs: Any) -> tuple[ImageClient, AuthResponse]:
        """Create a client by authenticating with an OAuth code."""
        client = cls(**kwargs)
        auth = client.authenticate(code)
        return client, auth

    def generate(self, prompt: str, ratio: str = "1:1", resolution: str = "1K",
                 rm_label_watermark: bool = True) -> ImageResponse:
        """Generate an image from a text prompt."""
        payload = json.dumps({
            "prompt": prompt,
            "ratio": ratio,
            "resolution": resolution,
            "rm_label_watermark": rm_label_watermark,
        }).encode("utf-8")

        req = Request(self.endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "*/*")
        req.add_header("X-Request-ID", _random_id(22))
        req.add_header("Origin", "https://image.z.ai")
        req.add_header("Referer", "https://image.z.ai/create")
        req.add_header("Cookie", f"session={self.session_token}")

        with urlopen(req, timeout=self.timeout) as resp:
            raw = json.loads(resp.read().decode("utf-8"))

        return self._parse_response(raw)

    @staticmethod
    def _parse_response(raw: dict[str, Any]) -> ImageResponse:
        d = raw.get("data", {}) if isinstance(raw.get("data"), dict) else {}
        img = d.get("image", {}) if isinstance(d.get("image"), dict) else {}
        return ImageResponse(
            code=raw.get("code", 0),
            message=raw.get("message", ""),
            image=ImageInfo(
                image_id=img.get("image_id", ""),
                prompt=img.get("prompt", ""),
                size=img.get("size", ""),
                ratio=img.get("ratio", ""),
                resolution=img.get("resolution", ""),
                image_url=img.get("image_url", ""),
                status=img.get("status", ""),
                created_at=img.get("created_at", ""),
                updated_at=img.get("updated_at", ""),
                width=img.get("width", 0),
                height=img.get("height", 0),
            ),
            timestamp=raw.get("timestamp", 0),
            raw=raw,
        )


def _random_id(length: int) -> str:
    chars = string.ascii_lowercase + string.digits
    return "".join(random.choices(chars, k=length))
