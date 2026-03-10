"""Z.ai OCR API client.

Endpoint: POST https://ocr.z.ai/api/v1/z-ocr/tasks/process
Auth: Bearer JWT token, no signature required.
Request: multipart/form-data with "file" field.
"""

from .ocr import OCRClient, OCRResponse, AuthResponse

__all__ = ["OCRClient", "OCRResponse", "AuthResponse"]
