package platforms

import (
	"testing"

	"any2api-go/internal/core"
)

func TestProviderExecutionInterfaces(t *testing.T) {
	tests := []struct {
		name              string
		provider          core.Provider
		openAIChat        bool
		openAIStream      bool
		anthropicMessages bool
		anthropicStream   bool
	}{
		{name: "cursor", provider: NewCursorProvider(), openAIChat: true, openAIStream: true, anthropicMessages: true, anthropicStream: true},
		{name: "kiro", provider: NewKiroProvider(), openAIChat: true, openAIStream: true, anthropicMessages: true, anthropicStream: true},
		{name: "grok", provider: NewGrokProvider(), openAIChat: true, openAIStream: true},
		{name: "orchids", provider: NewOrchidsProvider(), openAIChat: true, openAIStream: true, anthropicMessages: true, anthropicStream: true},
		{name: "web", provider: NewWebProvider(), openAIChat: true, openAIStream: true},
		{name: "chatgpt", provider: NewChatGPTProvider(), openAIChat: true, openAIStream: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotOpenAIChat := tt.provider.(core.OpenAIChatProvider)
			if gotOpenAIChat != tt.openAIChat {
				t.Fatalf("OpenAIChatProvider mismatch: got %v want %v", gotOpenAIChat, tt.openAIChat)
			}
			_, gotOpenAIStream := tt.provider.(core.OpenAIStreamProvider)
			if gotOpenAIStream != tt.openAIStream {
				t.Fatalf("OpenAIStreamProvider mismatch: got %v want %v", gotOpenAIStream, tt.openAIStream)
			}
			_, gotAnthropicMessages := tt.provider.(core.AnthropicMessagesProvider)
			if gotAnthropicMessages != tt.anthropicMessages {
				t.Fatalf("AnthropicMessagesProvider mismatch: got %v want %v", gotAnthropicMessages, tt.anthropicMessages)
			}
			_, gotAnthropicStream := tt.provider.(core.AnthropicStreamProvider)
			if gotAnthropicStream != tt.anthropicStream {
				t.Fatalf("AnthropicStreamProvider mismatch: got %v want %v", gotAnthropicStream, tt.anthropicStream)
			}
		})
	}
}
