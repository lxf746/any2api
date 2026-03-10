from __future__ import annotations

from dataclasses import asdict

from gateway_types import ModelInfo, ProviderCapabilities
from .common import Provider, SimpleProvider, SnapshotGetter
from .cursor import CursorProvider
from .chatgpt import ChatGPTProvider
from .grok import GrokProvider
from .kiro import KiroProvider
from .orchids import OrchidsProvider
from .web import WebProvider


class ProviderRegistry:
    def __init__(self, default_provider: str):
        self.default_provider = default_provider
        self._providers: dict[str, Provider] = {}

    def register(self, provider: Provider) -> None:
        self._providers[provider.provider_id()] = provider

    def resolve(self, provider_id: str | None = None) -> Provider:
        key = provider_id or self.default_provider
        if key not in self._providers:
            raise KeyError(f"unknown provider: {key}")
        return self._providers[key]

    def models(self, provider_id: str | None = None) -> list[dict]:
        if provider_id:
            return [asdict(item) for item in self.resolve(provider_id).models()]
        all_models: list[dict] = []
        for key in sorted(self._providers):
            all_models.extend(asdict(item) for item in self._providers[key].models())
        return all_models

    def provider_ids(self) -> list[str]:
        return sorted(self._providers)


def default_registry(default_provider: str = "cursor", snapshot_getter: SnapshotGetter | None = None) -> ProviderRegistry:
    registry = ProviderRegistry(default_provider=default_provider)
    registry.register(CursorProvider())
    registry.register(KiroProvider(snapshot_getter))
    registry.register(GrokProvider(snapshot_getter))
    registry.register(OrchidsProvider(snapshot_getter))
    registry.register(WebProvider(snapshot_getter))
    registry.register(ChatGPTProvider(snapshot_getter))
    return registry