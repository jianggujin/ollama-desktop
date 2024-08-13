//go:build !windows && !darwin
// +build !windows,!darwin

package cmd

import (
	"context"
	"ollama-desktop/internal/util"
	"strings"
)

func CheckInstalled(ctx context.Context) (bool, error) {
	response, err := util.GetInvoker().CommandWithContext(ctx, "command -v ollama >/dev/null 2>&1 || { echo >&2 'not'; }")
	if err != nil {
		return false, err
	}
	return !strings.Contains(string(response), "not"), nil
}
