package pathutil

import (
	"fmt"
	"os/user"
	"strings"
)

const (
	homeVar = "$HOME"
)

func ReplaceHome(s string) (string, error) {
	if !strings.Contains(s, homeVar) {
		return s, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return strings.ReplaceAll(s, homeVar, u.HomeDir), nil
}
