package qwiclang

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

//go:embed runtime/qwic_runtime.c runtime/qwic_runtime.h
var runtimeFiles embed.FS

// WriteRuntime writes the bundled Qwic runtime sources into dir and returns the
// include/source directory path expected by the C backend.
func WriteRuntime(dir string) (string, error) {
	runtimeDir := filepath.Join(dir, "runtime")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return "", fmt.Errorf("create runtime directory: %w", err)
	}

	for _, name := range []string{"qwic_runtime.c", "qwic_runtime.h"} {
		data, err := runtimeFiles.ReadFile(path.Join("runtime", name))
		if err != nil {
			return "", fmt.Errorf("read bundled runtime %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(runtimeDir, name), data, 0o644); err != nil {
			return "", fmt.Errorf("write bundled runtime %s: %w", name, err)
		}
	}

	return runtimeDir, nil
}
