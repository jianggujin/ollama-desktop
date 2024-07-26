//go:build windows
// +build windows

package cmd

import (
	"context"
	"errors"
	"fmt"
	"ollama-desktop/internal/ollama/api"
	"os"
	"os/exec"
	"path/filepath"
)

func StartApp(ctx context.Context, client *api.Client) error {
	// log.Printf("XXX Attempting to find and start ollama app")
	AppName := "ollama app.exe"
	exe, err := os.Executable()
	if err != nil {
		return err
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
				return fmt.Errorf("could not locate ollama app")
			}
		}
	}
	// log.Printf("XXX attempting to start app %s", appExe)

	// cmdPath := "c:\\Windows\\system32\\cmd.exe"
	// cmd := exec.Command(cmdPath, "/c", appExe)
	cmd := exec.Command(appExe)

	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	CreationFlags: 0x08000000,
	//	HideWindow:    true,
	//}

	//cmd.Stdin = strings.NewReader("")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start ollama app %w", err)
	}

	//if cmd.Process != nil {
	//	defer cmd.Process.Release() //nolint:errcheck
	//}
	return waitForServer(ctx, client)
}
