package packages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerCopiesPackageFromCatalog(t *testing.T) {
	catalog := t.TempDir()
	writeTestFile(t, filepath.Join(catalog, "sample", "src", "sample.qw"), `public func message(): string {
    return "installed"
}
`)
	writeTestFile(t, filepath.Join(catalog, "sample", "README.md"), "# Sample\n")

	destination := t.TempDir()
	installer := Installer{Repository: catalog, Destination: destination}
	result, err := installer.Install("sample")
	if err != nil {
		t.Fatalf("install package: %v", err)
	}
	if !result.Installed {
		t.Fatal("expected a new installation")
	}
	data, err := os.ReadFile(filepath.Join(result.Path, "src", "sample.qw"))
	if err != nil {
		t.Fatalf("read installed source: %v", err)
	}
	if !strings.Contains(string(data), `return "installed"`) {
		t.Fatalf("installed source = %q", data)
	}

	second, err := installer.Install("sample")
	if err != nil {
		t.Fatalf("install existing package: %v", err)
	}
	if second.Installed {
		t.Fatal("expected existing package installation to be reused")
	}
}

func TestInstallerRejectsUnsafePackageName(t *testing.T) {
	installer := Installer{Repository: t.TempDir(), Destination: t.TempDir()}
	for _, name := range []string{"../sample", "with/slash", "with-dash", ""} {
		if _, err := installer.Install(name); err == nil {
			t.Fatalf("expected package name %q to be rejected", name)
		}
	}
}

func TestInstallerRequiresQwicSources(t *testing.T) {
	catalog := t.TempDir()
	writeTestFile(t, filepath.Join(catalog, "empty", "README.md"), "# Empty\n")
	installer := Installer{Repository: catalog, Destination: t.TempDir()}

	_, err := installer.Install("empty")
	if err == nil || !strings.Contains(err.Error(), "src directory") {
		t.Fatalf("expected missing source diagnostic, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
