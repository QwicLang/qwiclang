package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryUsesConfiguredAbsolutePath(t *testing.T) {
	relative := filepath.Join("testdata", "shared-packages")
	t.Setenv("QWIC_PACKAGES_DIR", relative)

	directory, err := Directory()
	if err != nil {
		t.Fatalf("resolve package directory: %v", err)
	}
	if !filepath.IsAbs(directory) {
		t.Fatalf("directory = %q, want an absolute path", directory)
	}
	want, err := filepath.Abs(relative)
	if err != nil {
		t.Fatalf("resolve expected directory: %v", err)
	}
	if directory != want {
		t.Fatalf("directory = %q, want %q", directory, want)
	}
}

func TestDirectoryUsesNativeUserHomePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("QWIC_PACKAGES_DIR", "")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	directory, err := Directory()
	if err != nil {
		t.Fatalf("resolve package directory: %v", err)
	}
	want := filepath.Join(home, "qwic", "packages")
	if directory != want {
		t.Fatalf("directory = %q, want %q", directory, want)
	}
}

func TestInstallerCreatesNestedSharedPackageDirectory(t *testing.T) {
	catalog := t.TempDir()
	writeTestFile(t, filepath.Join(catalog, "sample", "src", "sample.qw"), "public func value(): int { return 1 }\n")
	destination := filepath.Join(t.TempDir(), "qwic", "packages")

	result, err := (Installer{Repository: catalog, Destination: destination}).Install("sample")
	if err != nil {
		t.Fatalf("install package: %v", err)
	}
	info, err := os.Stat(result.Path)
	if err != nil {
		t.Fatalf("inspect installed package: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("installed path %q is not a directory", result.Path)
	}
}
