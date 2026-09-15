package packages

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Installer struct {
	Repository  string
	Destination string
}

type InstallResult struct {
	Path      string
	Installed bool
}

func DefaultInstaller() (Installer, error) {
	destination, err := Directory()
	if err != nil {
		return Installer{}, err
	}
	return Installer{Repository: Repository(), Destination: destination}, nil
}

func (installer Installer) Install(name string) (InstallResult, error) {
	if !validName.MatchString(name) {
		return InstallResult{}, fmt.Errorf("invalid package name %q: expected a Qwic identifier", name)
	}
	if strings.TrimSpace(installer.Repository) == "" {
		return InstallResult{}, fmt.Errorf("package repository is not configured")
	}
	if strings.TrimSpace(installer.Destination) == "" {
		return InstallResult{}, fmt.Errorf("package destination is not configured")
	}

	if err := os.MkdirAll(installer.Destination, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create package directory: %w", err)
	}
	target := filepath.Join(installer.Destination, name)
	if info, err := os.Stat(target); err == nil {
		if !info.IsDir() {
			return InstallResult{}, fmt.Errorf("package destination %s is not a directory", target)
		}
		if err := validatePackageSource(target); err != nil {
			return InstallResult{}, fmt.Errorf("installed package %q is invalid: %w", name, err)
		}
		return InstallResult{Path: target}, nil
	} else if !os.IsNotExist(err) {
		return InstallResult{}, fmt.Errorf("inspect package destination: %w", err)
	}

	repositoryPath, cleanup, err := installer.fetchRepository(name)
	if err != nil {
		return InstallResult{}, err
	}
	defer cleanup()

	source := filepath.Join(repositoryPath, name)
	if err := validatePackageSource(source); err != nil {
		return InstallResult{}, fmt.Errorf("package %q: %w", name, err)
	}

	staging, err := os.MkdirTemp(installer.Destination, "."+name+"-")
	if err != nil {
		return InstallResult{}, fmt.Errorf("create package staging directory: %w", err)
	}
	defer os.RemoveAll(staging)

	if err := copyDirectory(source, staging); err != nil {
		return InstallResult{}, fmt.Errorf("copy package %q: %w", name, err)
	}
	if err := os.Rename(staging, target); err != nil {
		return InstallResult{}, fmt.Errorf("install package %q: %w", name, err)
	}

	return InstallResult{Path: target, Installed: true}, nil
}

func (installer Installer) fetchRepository(name string) (string, func(), error) {
	if info, err := os.Stat(installer.Repository); err == nil && info.IsDir() {
		absolute, resolveErr := filepath.Abs(installer.Repository)
		if resolveErr != nil {
			return "", func() {}, fmt.Errorf("resolve package repository: %w", resolveErr)
		}
		return absolute, func() {}, nil
	}

	tempDir, err := os.MkdirTemp("", "qwic-packages-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("create package download directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }
	repositoryPath := filepath.Join(tempDir, "repository")
	clone := exec.Command("git", "clone", "--depth", "1", "--filter=blob:none", "--sparse", "--single-branch", installer.Repository, repositoryPath)
	output, err := clone.CombinedOutput()
	if err != nil {
		cleanup()
		return "", func() {}, commandError("download package repository", output, err)
	}

	sparseCheckout := exec.Command("git", "-C", repositoryPath, "sparse-checkout", "set", name)
	output, err = sparseCheckout.CombinedOutput()
	if err != nil {
		cleanup()
		return "", func() {}, commandError("select package from repository", output, err)
	}
	return repositoryPath, cleanup, nil
}

func commandError(action string, output []byte, err error) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		message = err.Error()
	}
	return fmt.Errorf("%s: %s", action, message)
}

func validatePackageSource(source string) error {
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return fmt.Errorf("not found in %s", source)
	}
	if err != nil {
		return fmt.Errorf("inspect source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("catalog entry is not a directory")
	}

	hasSource := false
	err = filepath.WalkDir(filepath.Join(source, "src"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not allowed: %s", path)
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".qw") {
			hasSource = true
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("inspect src directory: %w", err)
	}
	if !hasSource {
		return fmt.Errorf("src directory contains no .qw source files")
	}
	return nil
}

func copyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not allowed: %s", path)
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		input, err := os.Open(path)
		if err != nil {
			return err
		}

		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err != nil {
			_ = input.Close()
			return err
		}
		_, copyErr := io.Copy(output, input)
		inputCloseErr := input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if inputCloseErr != nil {
			return inputCloseErr
		}
		return closeErr
	})
}
