//go:build !windows && !darwin
// +build !windows,!darwin

package cmd

import (
	"context"
	"fmt"
	"ollama-desktop/internal/ollama/api"
)

func StartApp(ctx context.Context, client *api.Client) error {
	return fmt.Errorf("could not connect to ollama server, run 'ollama serve' to start it")
}
