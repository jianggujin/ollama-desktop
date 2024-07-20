//go:build darwin
// +build darwin

package cmd

import (
	"context"
	"strings"
)

func CheckInstalled(ctx context.Context) (bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	link, err := os.Readlink(exe)
	if err != nil {
		return err
	}
	if !strings.Contains(link, "Ollama.app") {
		return false, fmt.Errorf("could not find ollama app")
	}
	return true, nil
}
