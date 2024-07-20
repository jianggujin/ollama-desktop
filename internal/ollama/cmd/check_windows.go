//go:build windows
// +build windows

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckInstalled(ctx context.Context) (bool, error) {
	// log.Printf("XXX Attempting to find and start ollama app")
	AppName := "ollama app.exe"
	exe, err := os.Executable()
	if err != nil {
		return false, err
	}
	appExe := filepath.Join(filepath.Dir(exe), AppName)
	_, err = os.Stat(appExe)
	if errors.Is(err, os.ErrNotExist) {
		// Try the standard install location
		localAppData := os.Getenv("LOCALAPPDATA")
		appExe = filepath.Join(localAppData, "Ollama", AppName)
		_, err := os.Stat(appExe)
		if errors.Is(err, os.ErrNotExist) {
			// Finally look in the path
			appExe, err = exec.LookPath(AppName)
			if err != nil {
				return false, fmt.Errorf("could not locate ollama app")
			}
		}
	}

	return true, nil
}
