package app

import (
	"context"
	"net/http"
	"net/url"
	ollama2 "ollama-desktop/internal/ollama/ollama"
	"testing"
)

func TestOllama_ModelInfoOnline(t *testing.T) {
	base, _ := url.Parse("https://ollama.com")

	client := &ollama2.Client{
		Base: base,
		Http: http.DefaultClient,
	}
	resp, err := client.ModelInfo(context.Background(), "llama3.1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("model", resp.Model)
	t.Log("tags")
	for _, tag := range resp.Tags {
		t.Log(tag)
	}
	t.Log("metas")
	for _, meta := range resp.Metas {
		t.Log(meta)
	}
	t.Log("readme", resp.Readme)
}
