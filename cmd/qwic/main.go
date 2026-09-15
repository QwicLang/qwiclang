package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"qwiclang"
	"qwiclang/internal/codegen"
	"qwiclang/internal/diagnostic"
	qwicfmt "qwiclang/internal/format"
	"qwiclang/internal/ir"
	"qwiclang/internal/lsp"
	"qwiclang/internal/packages"
	"qwiclang/internal/project"
	"qwiclang/internal/sema"
)

const Version = "v0.1.0-alpha.1"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return 0
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Printf("qwic version %s\n", Version)
		return 0
	}

	switch args[0] {
	case "build":
		return runBuild(args[1:])
	case "run":
		return runRun(args[1:])
	case "check":
		return runCheck(args[1:])
	case "fmt":
		return runFmt(args[1:])
	case "clean":
		return runClean(args[1:])
	case "install":
		return runInstall(args[1:])
	case "lsp":
		return runLSP(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printHelp()
		return 2
	}
}

func runCheck(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: qwic check <source.qw>")
		return 2
	}

	files, diagnostics := project.CollectSourceFiles(args[0])
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}
	result, diagnostics := sema.CheckFiles(files)
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}
	if _, diagnostics := ir.Build(result); len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}
	return 0
}

func runFmt(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: qwic fmt <source.qw>")
		return 2
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read source file: %v\n", err)
		return 1
	}
	formatted, diagnostics := qwicfmt.Source(args[0], string(source))
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}
	if err := os.WriteFile(args[0], []byte(formatted), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write formatted source: %v\n", err)
		return 1
	}
	return 0
}

func runClean(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: qwic clean [source.qw]")
		return 2
	}

	target := ".qwic-cache"
	if len(args) == 1 {
		target = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
		if target == "" {
			target = "a.out"
		}
	}

	if err := os.RemoveAll(target); err != nil {
		fmt.Fprintf(os.Stderr, "clean %s: %v\n", target, err)
		return 1
	}
	return 0
}

func runInstall(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: qwic install <package>")
		return 2
	}

	installer, err := packages.DefaultInstaller()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configure package installer: %v\n", err)
		return 1
	}
	result, err := installer.Install(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "install package: %v\n", err)
		return 1
	}
	if result.Installed {
		fmt.Printf("installed %s at %s\n", args[0], result.Path)
	} else {
		fmt.Printf("package %s is already installed at %s\n", args[0], result.Path)
	}
	return 0
}

func runBuild(args []string) int {
	sourcePath, outputPath, ok := parseBuildArgs(args)
	if !ok {
		return 2
	}

	return buildSource(sourcePath, outputPath)
}

func runRun(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: qwic run <source.qw>")
		return 2
	}

	tempDir, err := os.MkdirTemp("", "qwic-run-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temporary run directory: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, executableName(strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))))
	if exitCode := buildSource(args[0], outputPath); exitCode != 0 {
		return exitCode
	}

	exitCode, output, err := codegen.RunExecutable(outputPath)
	if output != "" {
		fmt.Print(output)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "run executable: %v\n", err)
		return 1
	}
	return exitCode
}

func parseBuildArgs(args []string) (string, string, bool) {
	if len(args) != 1 && len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: qwic build <source.qw> [-o output]")
		return "", "", false
	}

	sourcePath := args[0]
	outputPath := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	if outputPath == "" {
		outputPath = "a.out"
	}
	outputPath = executableName(outputPath)

	if len(args) == 3 {
		if args[1] != "-o" {
			fmt.Fprintln(os.Stderr, "usage: qwic build <source.qw> [-o output]")
			return "", "", false
		}
		outputPath = args[2]
	}

	return sourcePath, outputPath, true
}

func buildSource(sourcePath, outputPath string) int {
	buildDir, err := os.MkdirTemp("", "qwic-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temporary build directory: %v\n", err)
		return 1
	}
	defer os.RemoveAll(buildDir)

	files, diagnostics := project.CollectSourceFiles(sourcePath)
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}

	result, diagnostics := sema.CheckFiles(files)
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}

	module, diagnostics := ir.Build(result)
	if len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}

	runtimePath, err := qwiclang.WriteRuntime(buildDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare bundled runtime: %v\n", err)
		return 1
	}

	if diagnostics := codegen.BuildExecutable(module, codegen.Options{OutputPath: outputPath, WorkDir: buildDir, RuntimePath: runtimePath}); len(diagnostics) > 0 {
		printDiagnostics(diagnostics)
		return 1
	}

	return 0
}

func printDiagnostics(diagnostics []diagnostic.Diagnostic) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintln(os.Stderr, diagnostic.Error())
	}
}

func executableName(name string) string {
	if name == "" {
		return "a.out"
	}
	return name
}

func runLSP(args []string) int {
	server := lsp.NewServer(Version)
	if err := server.RunStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "lsp error: %v\n", err)
		return 1
	}
	return 0
}

func printHelp() {
	fmt.Println(`qwic - QwicLang compiler

Usage:
  qwic build <source.qw> [-o output]
  qwic run <source.qw>
  qwic check <source.qw>
  qwic fmt <source.qw>
  qwic clean [source.qw]
  qwic install <package>
  qwic lsp
  qwic --version
  qwic --help`)
}
