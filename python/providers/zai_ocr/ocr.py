"""Z.ai OCR API client — uploads a file and returns structured OCR results."""

from __future__ import annotations

import json
import uuid
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from urllib.request import Request, urlopen
from io import BytesIO


DEFAULT_ENDPOINT = "https://ocr.z.ai/api/v1/z-ocr/tasks/process"
DEFAULT_AUTH_ENDPOINT = "https://ocr.z.ai/api/v1/z-ocr/auth/"
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
class OCRResponse:
    """Parsed OCR API response."""
    code: int = 0
    message: str = ""
    task_id: str = ""
    status: str = ""
    file_name: str = ""
    file_size: int = 0
    file_type: str = ""
    file_url: str = ""
    created_at: str = ""
    markdown_content: str = ""
    json_content: dict[str, Any] = field(default_factory=dict)
    layout: list[dict[str, Any]] = field(default_factory=list)
    data_info: Any = None
    timestamp: int = 0
    raw: dict[str, Any] = field(default_factory=dict)


class OCRClient:
    """Client for the Z.ai OCR API."""

    def __init__(self, token: str = "", endpoint: str = DEFAULT_ENDPOINT,
                 auth_endpoint: str = DEFAULT_AUTH_ENDPOINT, timeout: int = DEFAULT_TIMEOUT):
        self.token = token
        self.endpoint = endpoint
        self.auth_endpoint = auth_endpoint
        self.timeout = timeout

    def authenticate(self, code: str) -> AuthResponse:
        """Exchange an OAuth code for an auth token. Auto-sets self.token on success."""
        payload = json.dumps({"code": code}).encode("utf-8")
        req = Request(self.auth_endpoint, data=payload, method="POST")
        req.add_header("Content-Type", "application/json")
        req.add_header("Accept", "application/json, text/plain, */*")
        req.add_header("X-Request-ID", str(uuid.uuid4()))
        req.add_header("Origin", "https://ocr.z.ai")
        req.add_header("Referer", "https://ocr.z.ai/")

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
        if result.auth_token:
            self.token = result.auth_token
        return result

    @classmethod
    def from_code(cls, code: str, **kwargs: Any) -> tuple[OCRClient, AuthResponse]:
        """Create a client by authenticating with an OAuth code."""
        client = cls(**kwargs)
        auth = client.authenticate(code)
        return client, auth

    def process_file(self, file_path: str | Path) -> OCRResponse:
        """Upload a local file and return the OCR result."""
        path = Path(file_path)
        with open(path, "rb") as f:
            return self.process_bytes(f.read(), path.name)

    def process_bytes(self, data: bytes, filename: str) -> OCRResponse:
        """Upload raw bytes and return the OCR result."""
        boundary = uuid.uuid4().hex
        body = self._build_multipart(data, filename, boundary)

        req = Request(self.endpoint, data=body, method="POST")
        req.add_header("Content-Type", f"multipart/form-data; boundary={boundary}")
        req.add_header("Authorization", f"Bearer {self.token}")
        req.add_header("X-Request-ID", str(uuid.uuid4()))
        req.add_header("Accept", "application/json, text/plain, */*")
        req.add_header("Origin", "https://ocr.z.ai")
        req.add_header("Referer", "https://ocr.z.ai/")

        with urlopen(req, timeout=self.timeout) as resp:
            raw = json.loads(resp.read().decode("utf-8"))

        return self._parse_response(raw)

    @staticmethod
    def _build_multipart(data: bytes, filename: str, boundary: str) -> bytes:
        """Build a multipart/form-data body with a single file field."""
        buf = BytesIO()
        buf.write(f"--{boundary}\r\n".encode())
        buf.write(f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'.encode())
        buf.write(b"Content-Type: application/octet-stream\r\n\r\n")
        buf.write(data)
        buf.write(f"\r\n--{boundary}--\r\n".encode())
        return buf.getvalue()

    @staticmethod
    def _parse_response(raw: dict[str, Any]) -> OCRResponse:
        """Parse the raw JSON into an OCRResponse, including double-parsing json_content."""
        d = raw.get("data", {}) if isinstance(raw.get("data"), dict) else {}
        json_content_raw = d.get("json_content", "")
        json_content: dict[str, Any] = {}
        if isinstance(json_content_raw, str) and json_content_raw.strip():
            try:
                json_content = json.loads(json_content_raw)
            except json.JSONDecodeError:
                json_content = {"_raw": json_content_raw}

        return OCRResponse(
            code=raw.get("code", 0),
            message=raw.get("message", ""),
            task_id=d.get("task_id", ""),
            status=d.get("status", ""),
            file_name=d.get("file_name", ""),
            file_size=d.get("file_size", 0),
            file_type=d.get("file_type", ""),
            file_url=d.get("file_url", ""),
            created_at=d.get("created_at", ""),
            markdown_content=d.get("markdown_content", ""),
            json_content=json_content,
            layout=d.get("layout", []),
            data_info=d.get("data_info"),
            timestamp=d.get("timestamp", 0),
            raw=raw,
        )
