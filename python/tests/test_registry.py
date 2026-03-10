import unittest

from providers import default_registry


class RegistryTests(unittest.TestCase):
    def test_registry_contains_six_providers(self):
        registry = default_registry()
        self.assertEqual(registry.provider_ids(), ["chatgpt", "cursor", "grok", "kiro", "orchids", "web"])

    def test_provider_specific_models(self):
        registry = default_registry()
        kiro_models = registry.models("kiro")
        self.assertEqual(len(kiro_models), 1)
        self.assertEqual(kiro_models[0]["provider"], "kiro")
        self.assertEqual(registry.models("web")[0]["provider"], "web")
        self.assertEqual(registry.models("chatgpt")[0]["provider"], "chatgpt")


if __name__ == "__main__":
    unittest.main()