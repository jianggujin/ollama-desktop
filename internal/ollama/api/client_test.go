package api

import (
	"context"
	"ollama-desktop/internal/ollama"
	"testing"
)

func TestClient_Chat(t *testing.T) {
	client := ClientFromConfig()
	client.Chat(context.Background(), &ollama.ChatRequest{
		Model: "qwen2:0.5b",
		Messages: []ollama.Message{
			{
				Role:    "user",
				Content: "介绍一下你自己",
			},
		},
	}, func(response ollama.ChatResponse) error {
		t.Log(response)
		return nil
	})
}
