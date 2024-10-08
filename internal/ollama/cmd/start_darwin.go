//go:build darwin
// +build darwin

package cmd

import (
	"context"
	"fmt"
	"ollama-desktop/internal/ollama/api"
	"os"
	"os/exec"
	"strings"
)

func StartApp(ctx context.Context, client *api.Client) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	link, err := os.Readlink(exe)
	if err != nil {
		return err
	}
	if !strings.Contains(link, "Ollama.app") {
		return fmt.Errorf("could not find ollama app")
	}
	path := strings.Split(link, "Ollama.app")
	if err := exec.Command("/usr/bin/open", "-a", path[0]+"Ollama.app").Run(); err != nil {
		return err
	}
	return waitForServer(ctx, client)
}
