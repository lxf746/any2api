"""Z.ai TTS (Text-to-Speech) API client.

Endpoint: POST https://audio.z.ai/api/v1/z-audio/tts/create
Auth: Bearer JWT token.
Response: SSE stream with {"audio":"<base64 WAV>"} chunks, ending with [DONE].
"""

from .tts import TTSClient, AuthResponse

__all__ = ["TTSClient", "AuthResponse"]
