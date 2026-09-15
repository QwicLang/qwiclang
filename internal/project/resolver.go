package project

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/packages"
	"qwiclang/internal/parser"
	"qwiclang/internal/sema"
	"qwiclang/internal/stdlib"
	"qwiclang/internal/token"
)

type sourceFile struct {
	path   string
	module string
}

type resolver struct {
	projectRoot    string
	globalPackages string
	visited        map[string]bool
	files          []sema.SourceFile
	diagnostics    []diagnostic.Diagnostic
}

func CollectSourceFiles(rootPath string) ([]sema.SourceFile, []diagnostic.Diagnostic) {
	absoluteRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, []diagnostic.Diagnostic{diagnostic.Error(position(rootPath), fmt.Sprintf("resolve source path: %v", err))}
	}
	globalPackages, err := packages.Directory()
	if err != nil {
		return nil, []diagnostic.Diagnostic{diagnostic.Error(position(rootPath), err.Error())}
	}

	loader := resolver{
		projectRoot:    findProjectRoot(absoluteRoot),
		globalPackages: globalPackages,
		visited:        map[string]bool{},
	}
	loader.visit(sourceFile{path: absoluteRoot})
	return loader.files, loader.diagnostics
}

func findProjectRoot(sourcePath string) string {
	sourceDirectory := filepath.Dir(sourcePath)
	for directory := sourceDirectory; ; directory = filepath.Dir(directory) {
		if info, err := os.Stat(filepath.Join(directory, "qwic.toml")); err == nil && !info.IsDir() {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
	}

	workingDirectory, err := os.Getwd()
	if err == nil {
		if relative, relativeErr := filepath.Rel(workingDirectory, sourcePath); relativeErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return workingDirectory
		}
	}
	return sourceDirectory
}

func (loader *resolver) visit(file sourceFile) {
	cleanPath, err := filepath.Abs(file.path)
	if err != nil {
		loader.addError(file.path, fmt.Sprintf("resolve source path: %v", err))
		return
	}
	if loader.visited[cleanPath] {
		return
	}
	loader.visited[cleanPath] = true

	source, err := os.ReadFile(cleanPath)
	if err != nil {
		loader.addError(cleanPath, fmt.Sprintf("read source file: %v", err))
		return
	}
	loader.files = append(loader.files, sema.SourceFile{
		Filename: cleanPath,
		Source:   string(source),
		Module:   file.module,
	})

	program, parseDiagnostics := parser.Parse(cleanPath, string(source))
	if len(parseDiagnostics) > 0 {
		loader.diagnostics = append(loader.diagnostics, parseDiagnostics...)
		return
	}
	for _, importedModule := range importsFor(program) {
		if stdlib.HasPackage(importedModule) {
			continue
		}
		resolved, resolveErr := loader.resolveImport(cleanPath, importedModule)
		if resolveErr != nil {
			loader.addError(cleanPath, resolveErr.Error())
			continue
		}
		for _, importedFile := range resolved {
			loader.visit(importedFile)
		}
	}
}

func (loader *resolver) resolveImport(importerPath, name string) ([]sourceFile, error) {
	locations := []string{
		filepath.Join(loader.globalPackages, name),
		filepath.Join(loader.projectRoot, "packages", name),
		filepath.Join(loader.projectRoot, name),
	}
	for _, location := range locations {
		files, found, err := packageFiles(location, name)
		if err != nil {
			return nil, err
		}
		if found {
			return files, nil
		}
	}

	legacyPath := filepath.Join(filepath.Dir(importerPath), name+".qw")
	if info, err := os.Stat(legacyPath); err == nil && !info.IsDir() {
		return []sourceFile{{path: legacyPath}}, nil
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect module %q: %w", name, err)
	}

	return nil, fmt.Errorf("package %q was not found in %s or the project workspace", name, loader.globalPackages)
}

func packageFiles(root, module string) ([]sourceFile, bool, error) {
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("inspect package %q: %w", module, err)
	}
	if !info.IsDir() {
		return nil, false, fmt.Errorf("package %q path %s is not a directory", module, root)
	}

	sourceRoot := filepath.Join(root, "src")
	var files []sourceFile
	err = filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("package %q contains unsupported symbolic link %s", module, path)
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".qw") {
			files = append(files, sourceFile{path: path, module: module})
		}
		return nil
	})
	if err != nil {
		return nil, true, fmt.Errorf("load package %q: %w", module, err)
	}
	if len(files) == 0 {
		return nil, true, fmt.Errorf("package %q contains no .qw source files under %s", module, sourceRoot)
	}
	return files, true, nil
}

func importsFor(program *ast.Program) []string {
	var imports []string
	for _, declaration := range program.Declarations {
		if imported, ok := declaration.(*ast.ImportDeclaration); ok {
			imports = append(imports, imported.Name)
		}
	}
	return imports
}

func (loader *resolver) addError(filename, message string) {
	loader.diagnostics = append(loader.diagnostics, diagnostic.Error(position(filename), message))
}

func position(filename string) token.Position {
	return token.Position{Filename: filename, Line: 1, Column: 1}
}
