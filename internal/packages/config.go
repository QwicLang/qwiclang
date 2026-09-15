package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultRepository = "https://github.com/QwicLang/packages.git"

func Directory() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("QWIC_PACKAGES_DIR")); configured != "" {
		absolute, err := filepath.Abs(configured)
		if err != nil {
			return "", fmt.Errorf("resolve QWIC_PACKAGES_DIR: %w", err)
		}
		return absolute, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, "qwic", "packages"), nil
}

func Repository() string {
	if configured := strings.TrimSpace(os.Getenv("QWIC_PACKAGES_REPOSITORY")); configured != "" {
		return configured
	}
	return DefaultRepository
}
