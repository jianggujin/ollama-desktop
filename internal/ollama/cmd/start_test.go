package cmd

import (
	"context"
	"ollama-desktop/internal/ollama/api"
	"testing"
)

func TestStartApp(t *testing.T) {
	err := StartApp(context.Background(), api.ClientFromConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Success...")
}
