package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBuildArgsDefaultsOutputFromSourceName(t *testing.T) {
	source, output, ok := parseBuildArgs([]string{"examples/hello.qw"})
	if !ok {
		t.Fatal("expected build args to parse")
	}
	if source != "examples/hello.qw" {
		t.Fatalf("source = %q, want examples/hello.qw", source)
	}
	if output != "hello" {
		t.Fatalf("output = %q, want hello", output)
	}
}

func TestParseBuildArgsAcceptsOutputPath(t *testing.T) {
	_, output, ok := parseBuildArgs([]string{"examples/hello.qw", "-o", "/tmp/hello"})
	if !ok {
		t.Fatal("expected build args to parse")
	}
	if output != "/tmp/hello" {
		t.Fatalf("output = %q, want /tmp/hello", output)
	}
}

func TestRunCheckReturnsZeroForValidProgram(t *testing.T) {
	path := writeTempSource(t, `public func main() {
    print("ok")
}
`)
	if exitCode := run([]string{"check", path}); exitCode != 0 {
		t.Fatalf("check exit code = %d, want 0", exitCode)
	}
}

func TestRunCheckReturnsOneForInvalidProgram(t *testing.T) {
	path := writeTempSource(t, `public func main() {
    missing()
}
`)
	if exitCode := run([]string{"check", path}); exitCode != 1 {
		t.Fatalf("check exit code = %d, want 1", exitCode)
	}
}

func TestRunFmtFormatsSourceInPlace(t *testing.T) {
	path := writeTempSource(t, `public func main(){print("ok")}`)
	if exitCode := run([]string{"fmt", path}); exitCode != 0 {
		t.Fatalf("fmt exit code = %d, want 0", exitCode)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read formatted source: %v", err)
	}
	want := "public func main() {\n    print(\"ok\")\n}\n"
	if string(data) != want {
		t.Fatalf("formatted source = %q, want %q", string(data), want)
	}
}

func TestRunBuildUsesBundledRuntimeOutsideRepoRoot(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.qw")
	outputPath := filepath.Join(tempDir, "main")
	if err := os.WriteFile(sourcePath, []byte(`public func main() {
    print("ok")
}
`), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	defer os.Chdir(previous)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}

	if exitCode := run([]string{"build", sourcePath, "-o", outputPath}); exitCode != 0 {
		t.Fatalf("build exit code = %d, want 0", exitCode)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected executable output: %v", err)
	}
}

func TestRunCheckResolvesStandardPackageWithoutLocalFile(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.qw")
	if err := os.WriteFile(sourcePath, []byte(`import strings

public func main() {
    print(strings.trim("  ok  "))
}
`), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if exitCode := run([]string{"check", sourcePath}); exitCode != 0 {
		t.Fatalf("check exit code = %d, want 0", exitCode)
	}
}

func TestRunCleanRemovesDefaultOutput(t *testing.T) {
	tempDir := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	defer os.Chdir(previous)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	if err := os.WriteFile("hello", []byte("binary"), 0o755); err != nil {
		t.Fatalf("write output: %v", err)
	}

	if exitCode := run([]string{"clean", "hello.qw"}); exitCode != 0 {
		t.Fatalf("clean exit code = %d, want 0", exitCode)
	}
	if _, err := os.Stat("hello"); !os.IsNotExist(err) {
		t.Fatalf("expected default output to be removed, stat err = %v", err)
	}
}

func writeTempSource(t *testing.T, source string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "main.qw")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("write temp source: %v", err)
	}
	return path
}
