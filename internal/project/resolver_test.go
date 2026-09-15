package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qwiclang/internal/sema"
)

func TestCollectSourceFilesLoadsMultiFileGlobalPackage(t *testing.T) {
	globalPackages := t.TempDir()
	t.Setenv("QWIC_PACKAGES_DIR", globalPackages)
	writeSource(t, filepath.Join(globalPackages, "greeting", "src", "api.qw"), `import strings

public func message(): string {
    return helper(strings.trim("  hello  "))
}
`)
	writeSource(t, filepath.Join(globalPackages, "greeting", "src", "helper.qw"), `func helper(value: string): string {
    return value
}
`)

	projectRoot := t.TempDir()
	mainPath := filepath.Join(projectRoot, "main.qw")
	writeSource(t, mainPath, `import greeting

public func main() {
    print(greeting.message())
}
`)

	files, diagnostics := CollectSourceFiles(mainPath)
	if len(diagnostics) > 0 {
		t.Fatalf("collect diagnostics: %v", diagnostics)
	}
	if len(files) != 3 {
		t.Fatalf("source file count = %d, want 3", len(files))
	}
	_, checkDiagnostics := sema.CheckFiles(files)
	if len(checkDiagnostics) > 0 {
		t.Fatalf("semantic diagnostics: %v", checkDiagnostics)
	}
}

func TestCollectSourceFilesPrefersGlobalPackage(t *testing.T) {
	globalPackages := t.TempDir()
	t.Setenv("QWIC_PACKAGES_DIR", globalPackages)
	writeSource(t, filepath.Join(globalPackages, "shared", "src", "shared.qw"), `public func value(): int {
    return 42
}
`)

	projectRoot := t.TempDir()
	writeSource(t, filepath.Join(projectRoot, "shared.qw"), `module shared

public func value(): string {
    return "local"
}
`)
	mainPath := filepath.Join(projectRoot, "main.qw")
	writeSource(t, mainPath, `import shared

public func main() {
    const value: int = shared.value()
    print(value)
}
`)

	files, diagnostics := CollectSourceFiles(mainPath)
	if len(diagnostics) > 0 {
		t.Fatalf("collect diagnostics: %v", diagnostics)
	}
	if len(files) != 2 {
		t.Fatalf("source file count = %d, want 2", len(files))
	}
	if !strings.Contains(files[1].Filename, filepath.Join("shared", "src", "shared.qw")) {
		t.Fatalf("resolved package path = %q", files[1].Filename)
	}
	_, checkDiagnostics := sema.CheckFiles(files)
	if len(checkDiagnostics) > 0 {
		t.Fatalf("semantic diagnostics: %v", checkDiagnostics)
	}
}

func TestCollectSourceFilesLoadsProjectPackageDirectory(t *testing.T) {
	t.Setenv("QWIC_PACKAGES_DIR", t.TempDir())
	projectRoot := t.TempDir()
	writeSource(t, filepath.Join(projectRoot, "packages", "custom", "src", "custom.qw"), `public func value(): string {
    return "project"
}
`)
	mainPath := filepath.Join(projectRoot, "main.qw")
	writeSource(t, mainPath, `import custom

public func main() {
    print(custom.value())
}
`)

	files, diagnostics := CollectSourceFiles(mainPath)
	if len(diagnostics) > 0 {
		t.Fatalf("collect diagnostics: %v", diagnostics)
	}
	_, checkDiagnostics := sema.CheckFiles(files)
	if len(checkDiagnostics) > 0 {
		t.Fatalf("semantic diagnostics: %v", checkDiagnostics)
	}
}

func TestCollectSourceFilesFindsWorkspacePackageAboveSourceDirectory(t *testing.T) {
	t.Setenv("QWIC_PACKAGES_DIR", t.TempDir())
	projectRoot := t.TempDir()
	writeSource(t, filepath.Join(projectRoot, "qwic.toml"), "[package]\nname = \"app\"\n")
	writeSource(t, filepath.Join(projectRoot, "packages", "custom", "src", "custom.qw"), `public func value(): string {
    return "workspace"
}
`)
	mainPath := filepath.Join(projectRoot, "src", "main.qw")
	writeSource(t, mainPath, `import custom

public func main() {
    print(custom.value())
}
`)

	files, diagnostics := CollectSourceFiles(mainPath)
	if len(diagnostics) > 0 {
		t.Fatalf("collect diagnostics: %v", diagnostics)
	}
	_, checkDiagnostics := sema.CheckFiles(files)
	if len(checkDiagnostics) > 0 {
		t.Fatalf("semantic diagnostics: %v", checkDiagnostics)
	}
}

func TestCollectSourceFilesReportsMissingPackageLocations(t *testing.T) {
	globalPackages := t.TempDir()
	t.Setenv("QWIC_PACKAGES_DIR", globalPackages)
	mainPath := filepath.Join(t.TempDir(), "main.qw")
	writeSource(t, mainPath, `import missing

public func main() {
}
`)

	_, diagnostics := CollectSourceFiles(mainPath)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1", len(diagnostics))
	}
	if !strings.Contains(diagnostics[0].Message, `package "missing" was not found`) || !strings.Contains(diagnostics[0].Message, globalPackages) {
		t.Fatalf("diagnostic = %q", diagnostics[0].Message)
	}
}

func writeSource(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create source directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
}
