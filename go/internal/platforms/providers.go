package platforms

import (
	"any2api-go/internal/core"
	chatgptprovider "any2api-go/internal/platforms/chatgpt"
	cursorprovider "any2api-go/internal/platforms/cursor"
	grokprovider "any2api-go/internal/platforms/grok"
	kiroprovider "any2api-go/internal/platforms/kiro"
	orchidsprovider "any2api-go/internal/platforms/orchids"
	webprovider "any2api-go/internal/platforms/web"
)

func NewCursorProvider() core.Provider {
	return cursorprovider.NewProvider()
}

func NewCursorProviderWithConfig(cfg core.CursorConfig) core.Provider {
	return cursorprovider.NewProviderWithConfig(cfg)
}

func NewKiroProvider() core.Provider {
	return kiroprovider.NewProvider()
}

func NewKiroProviderWithConfig(cfg core.KiroConfig) core.Provider {
	return kiroprovider.NewProviderWithConfig(cfg)
}

func NewGrokProvider() core.Provider {
	return grokprovider.NewProvider()
}

func NewGrokProviderWithConfig(cfg core.GrokConfig) core.Provider {
	return grokprovider.NewProviderWithConfig(cfg)
}

func NewOrchidsProvider() core.Provider {
	return orchidsprovider.NewProvider()
}

func NewOrchidsProviderWithConfig(cfg core.OrchidsConfig) core.Provider {
	return orchidsprovider.NewProviderWithConfig(cfg)
}

func NewWebProvider() core.Provider {
	return webprovider.NewProvider()
}

func NewWebProviderWithConfig(cfg core.WebConfig) core.Provider {
	return webprovider.NewProviderWithConfig(cfg)
}

func NewChatGPTProvider() core.Provider {
	return chatgptprovider.NewProvider()
}

func NewChatGPTProviderWithConfig(cfg core.ChatGPTConfig) core.Provider {
	return chatgptprovider.NewProviderWithConfig(cfg)
}
