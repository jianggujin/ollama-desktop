package app

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"ollama-desktop/internal/config"
	olm "ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	ollama2 "ollama-desktop/internal/ollama/ollama"
	"testing"
)

func TestOllama_Envs(t *testing.T) {
	envs := ollama.Envs()
	for _, env := range envs {
		t.Log(env)
	}
}

func TestOllama_Version(t *testing.T) {
	t.Log(newApiClient().Version(context.Background()))
}

func TestOllama_Show(t *testing.T) {
	resp, err := newApiClient().Show(context.Background(), &olm.ShowRequest{
		Model: "qwen2:0.5b",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("License", resp.License)
	t.Log("Modelfile", resp.Modelfile)
	t.Log("Parameters", resp.Parameters)
	t.Log("Template", resp.Template)
	t.Log("System", resp.System)
	t.Log("Details", resp.Details)
	t.Log("Messages", resp.Messages)
	t.Log("ModelInfo", resp.ModelInfo)
	t.Log("ProjectorInfo", resp.ProjectorInfo)
	t.Log("ModifiedAt", resp.ModifiedAt)
}

func TestOllama_ModelInfoOnline(t *testing.T) {
	resp, err := newOllamaClient().ModelInfo(context.Background(), "llama3.1")
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

func newApiClient() *api.Client {
	ollamaHost := config.Config.Ollama.Host
	return &api.Client{
		Base: &url.URL{
			Scheme: ollamaHost.Scheme,
			Host:   net.JoinHostPort(ollamaHost.Host, ollamaHost.Port),
		},
		Http: http.DefaultClient,
	}
}

func newOllamaClient() *ollama2.Client {
	base, _ := url.Parse("https://ollama.com")

	return &ollama2.Client{
		Base: base,
		Http: http.DefaultClient,
	}
}
