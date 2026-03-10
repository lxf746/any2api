from __future__ import annotations

import os
import unittest
from unittest.mock import patch

from server import current_server_host, current_server_port


class ServerEnvTests(unittest.TestCase):
    def test_prefers_newplatform_host_and_port(self):
        with patch.dict(os.environ, {
            "NEWPLATFORM2API_HOST": "0.0.0.0",
            "NEWPLATFORM2API_PORT": "9123",
            "HOST": "127.0.0.1",
            "PORT": "8100",
        }, clear=False):
            self.assertEqual(current_server_host(), "0.0.0.0")
            self.assertEqual(current_server_port(), 9123)

    def test_falls_back_to_host_and_port(self):
        with patch.dict(os.environ, {
            "HOST": "0.0.0.0",
            "PORT": "9100",
        }, clear=True):
            self.assertEqual(current_server_host(), "0.0.0.0")
            self.assertEqual(current_server_port(), 9100)

    def test_uses_defaults_when_port_is_invalid(self):
        with patch.dict(os.environ, {"NEWPLATFORM2API_PORT": "not-a-number"}, clear=True):
            self.assertEqual(current_server_host(), "127.0.0.1")
            self.assertEqual(current_server_port(), 8100)