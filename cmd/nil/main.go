// nil — Alap project management CLI
// Usage:
//   nil init <name>            create a new NilLang project
//   nil run [file]             compile and run
//   nil build <platform>       build for platform: onuron|android|ios|linux
//   nil check                  type-check all .nil files
//   nil fmt                    format all .nil files
//   nil test                   run tests
//   nil clean                  clean build artifacts
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/formatter"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/lsp"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/pkg/manager"
	"github.com/joysriramsarkar/alap-framework/platform/android"
	"github.com/joysriramsarkar/alap-framework/platform/ios"
	"github.com/joysriramsarkar/alap-framework/platform/linux"
	"github.com/joysriramsarkar/alap-framework/platform/onuron"
	"github.com/joysriramsarkar/alap-framework/runtime/repl"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
)

const version = "0.1.0-alpha"

// ─── Templates for new projects ───────────────────────────────────────────────

const alapYaml = `# alap.yaml — Alap Project Manifest
name: {{.Name}}
version: 0.1.0
description: "An Alap application"

entry: src/main.nil

permissions:
  - storage.read
  - notifications

targets:
  onuron:
    arch: arm64
  android:
    minSdk: 26
    targetSdk: 35
  ios:
    minVersion: "16.0"
  linux:
    display: wayland

dependencies: {}
`

const mainNil = `// {{.Name}} — main entry point
function add(a: i32, b: i32): i32 {
    return a + b
}

function greet(name: string): void {
    print("Welcome to Alap, " + name + "!")
}

function main(): void {
    print("{{.Name}} is running on Onuron!")
    
    let x: i32 = 10
    let y: i32 = 20
    let sum: i32 = add(x, y)
    print("Sum: " + sum.toString())
    
    greet("Onuron User")
}

main()
`

const appNil = `// App.nil — Root UI component (Alap declarative UI)
@Entry
@Component
component App {
    @State title: string = "{{.Name}}"
    @State count: i32 = 0
    @State message: string = "Click the button!"

    build() {
        Column {
            Text(title)
                .fontSize(32)
                .fontWeight("bold")
                .color("#176BFF")
                .margin({ bottom: 16 })

            Text("Count: " + count.toString())
                .fontSize(24)
                .color("#333333")

            Text(message)
                .fontSize(16)
                .color("#666666")
                .margin({ top: 8, bottom: 16 })

            Row {
                Button("-") {
                    onClick => {
                        if count > 0 {
                            count -= 1
                        }
                        message = "Decremented to " + count.toString()
                    }
                }
                .padding({ horizontal: 24, vertical: 12 })
                .backgroundColor("#FF3B30")
                .color("#FFFFFF")
                .borderRadius(8)

                Button("+") {
                    onClick => {
                        count += 1
                        message = "Incremented to " + count.toString()
                    }
                }
                .padding({ horizontal: 24, vertical: 12 })
                .backgroundColor("#34C759")
                .color("#FFFFFF")
                .borderRadius(8)
            }
            .spacing(16)
            .margin({ top: 16 })
        }
        .width("100%")
        .height("100%")
        .justifyContent("center")
        .alignItems("center")
        .backgroundColor("#F5F5F7")
        .padding(24)
    }
}
`

const testNil = `// {{.Name}}_test.nil — tests
function testAdd(): void {
    let result: i32 = 10 + 20
    if result == 30 {
        print("testAdd: PASSED")
    } else {
        print("testAdd: FAILED")
    }
}

function testString(): void {
    let s: string = "Hello " + "Onuron"
    if s == "Hello Onuron" {
        print("testString: PASSED")
    } else {
        print("testString: FAILED")
    }
}

function main(): void {
    testAdd()
    testString()
    print("All tests passed!")
}

main()
`

const readme = `# {{.Name}}

An Alap Framework application written in NilLang.

## Getting Started

` + "```bash" + `
# Run the application
nil run

# Build for NilOS
nil build onuron

# Build for Android
nil build android

# Build for iOS
nil build ios

# Build for Linux
nil build linux

# Type-check
nil check

# Run tests
nil test
` + "```" + `

## Project Structure

` + "```" + `
{{.Name}}/
├── src/
│   ├── main.nil      # Entry point
│   ├── App.nil       # Root UI component
│   └── ...
├── tests/
│   └── main_test.nil
├── assets/
├── alap.yaml         # Project manifest
└── README.md
` + "```" + `
`

// ─── Internal Compiler Bridge ────────────────────────────────────────────────

func compileFile(filename string) (*codegen.Module, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read error: %v", err)
	}

	l := lexer.New(filename, string(src))
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		return nil, fmt.Errorf("lex error: %s", strings.Join(l.Errors(), "\n"))
	}

	p := parser.New(filename, tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse error: %s", strings.Join(p.Errors(), "\n"))
	}

	checker := types.New()
	checker.CheckProgram(prog)
	if len(checker.Errors()) > 0 {
		return nil, fmt.Errorf("type error: %s", strings.Join(checker.Errors(), "\n"))
	}

	modName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	gen := codegen.New(modName)
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		return nil, fmt.Errorf("codegen error: %s", strings.Join(gen.Errors(), "\n"))
	}

	return gen.Module(), nil
}

// ─── CLI Dispatcher ──────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		name := "my-app"
		if len(args) > 0 {
			name = args[0]
		}
		cmdInit(name)
	case "run":
		file := "src/main.nil"
		if len(args) > 0 {
			file = args[0]
		}
		cmdRun(file)
	case "repl":
		repl.Start(os.Stdin, os.Stdout)
	case "lsp":
		server := lsp.NewServer(os.Stdin, os.Stdout)
		if err := server.Serve(); err != nil {
			fmt.Fprintf(os.Stderr, "nil lsp: error: %v\n", err)
		}
	case "pm":
		pm := manager.New(".")
		if err := pm.Install(); err != nil {
			fmt.Printf("nil pm: error: %v\n", err)
		} else {
			fmt.Println("✓ Dependencies resolved and locked in alap.lock")
		}
	case "build":
		platform := "onuron"
		if len(args) > 0 {
			platform = args[0]
		}
		cmdBuild(platform)
	case "check":
		cmdCheck()
	case "test":
		cmdTest()
	case "fmt":
		cmdFmt()
	case "add":
		if len(args) == 0 {
			fmt.Println("Usage: nil add <package-name>")
			return
		}
		cmdAdd(args[0])
	case "clean":
		cmdClean()
	case "version":
		fmt.Printf("nil (Alap Framework) v%s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "nil: unknown command %q\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`nil — Alap Framework Project Tool v%s

Commands:
  nil init <name>         Create a new Alap project
  nil run [file]          Compile and run (default: src/main.nil)
  nil repl                Start interactive NilLang REPL
  nil lsp                 Start Language Server Protocol daemon
  nil pm [install]        Resolve and lock package dependencies
  nil build <platform>    Build for platform:
                            onuron  → Onuron native app (.nilax / .nabc)
                            android → Android APK / JNI scaffold
                            ios     → iOS app / Metal scaffold
                            linux   → Linux desktop bundle / Flatpak
  nil check               Type-check all .nil source files
  nil test                Run test files (*_test.nil)
  nil fmt                 Format all .nil files
  nil add <package>       Add package dependency to alap.yaml
  nil clean               Remove build artifacts
  nil version             Print version

Examples:
  nil init my-app
  nil run
  nil repl
  nil build onuron
  nil build android
  nil build linux
  nil pm install
  nil fmt
`, version)
}

// ─── Commands ─────────────────────────────────────────────────────────────────

type templateData struct{ Name string }

func cmdInit(name string) {
	dirs := []string{
		filepath.Join(name, "src"),
		filepath.Join(name, "tests"),
		filepath.Join(name, "assets"),
		filepath.Join(name, "native", "onuron"),
		filepath.Join(name, "native", "android"),
		filepath.Join(name, "native", "ios"),
		filepath.Join(name, "native", "linux"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "nil init: %v\n", err)
			os.Exit(1)
		}
	}

	data := templateData{Name: name}
	files := map[string]string{
		filepath.Join(name, "alap.yaml"):               alapYaml,
		filepath.Join(name, "src", "main.nil"):          mainNil,
		filepath.Join(name, "src", "App.nil"):           appNil,
		filepath.Join(name, "tests", "main_test.nil"): testNil,
		filepath.Join(name, "README.md"):               readme,
	}
	for path, tmplStr := range files {
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nil init: %v\n", err)
			os.Exit(1)
		}
		tmpl := template.Must(template.New("f").Parse(tmplStr))
		if err := tmpl.Execute(f, data); err != nil {
			fmt.Fprintf(os.Stderr, "nil init: template error: %v\n", err)
		}
		f.Close()
	}

	fmt.Printf("✓ Created Alap project: %s\n\n", name)
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  nil run\n\n")
	fmt.Println("Happy coding with NilLang! 🚀")
}

func cmdRun(file string) {
	mod, err := compileFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nil run: %v\n", err)
		os.Exit(1)
	}

	runner := vm.New(mod)
	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nil run: runtime error: %v\n", err)
		os.Exit(1)
	}
}

func cmdBuild(platform string) {
	fmt.Printf("nil: building for %s...\n", platform)
	switch platform {
	case "onuron", "nilos":
		buildOnuron()
	case "android":
		buildAndroid()
	case "ios":
		buildIOS()
	case "linux":
		buildLinux()
	default:
		fmt.Fprintf(os.Stderr, "nil build: unknown platform %q (use: onuron|android|ios|linux)\n", platform)
		os.Exit(1)
	}
}

func buildOnuron() {
	bytecode, err := compileEntryBytecode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nil build onuron: %v\n", err)
		os.Exit(1)
	}

	adapter := onuron.New()
	outDir := filepath.Join("build")
	if err := adapter.GenerateProject(outDir, bytecode); err != nil {
		fmt.Fprintf(os.Stderr, "nil build onuron: error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built for Onuron native app (.nilax) → build/onuron/\n")
	fmt.Printf("  • Manifest: build/onuron/app.alapmanifest\n")
	fmt.Printf("  • Bytecode bundled: %d bytes in bin/main.nabc\n\n", len(bytecode))
}

func buildAndroid() {
	bytecode, err := compileEntryBytecode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nil build android: %v\n", err)
		os.Exit(1)
	}

	adapter := android.New()
	outDir := filepath.Join("build", "android")
	if err := adapter.GenerateProject(outDir, bytecode); err != nil {
		fmt.Fprintf(os.Stderr, "nil build android: error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Android Gradle project generated → %s/\n", outDir)
	fmt.Printf("  • App module: %s/app\n", outDir)
	fmt.Printf("  • Bytecode bundled: %d bytes in assets/main.nabc\n", len(bytecode))
	fmt.Printf("  • Build command: cd %s && ./gradlew assembleDebug\n\n", outDir)
}

func buildIOS() {
	bytecode, err := compileEntryBytecode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nil build ios: %v\n", err)
		os.Exit(1)
	}

	adapter := ios.New()
	outDir := filepath.Join("build", "ios")
	if err := adapter.GenerateProject(outDir, bytecode); err != nil {
		fmt.Fprintf(os.Stderr, "nil build ios: error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ iOS Swift/Metal project generated → %s/\n", outDir)
	fmt.Printf("  • App source: %s/ios/%s\n", outDir, adapter.AppName)
	fmt.Printf("  • Bytecode bundled: %d bytes in assets\n\n", len(bytecode))
}

func buildLinux() {
	bytecode, err := compileEntryBytecode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nil build linux: %v\n", err)
		os.Exit(1)
	}

	adapter := linux.New()
	outDir := filepath.Join("build")
	if err := adapter.GenerateProject(outDir, bytecode); err != nil {
		fmt.Fprintf(os.Stderr, "nil build linux: error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Built for Linux desktop → build/linux/\n")
	fmt.Printf("  • AppDir bundle: build/linux/AppDir\n")
	fmt.Printf("  • Flatpak manifest: build/linux/org.alap.app.json\n")
	fmt.Printf("  • Bytecode bundled: %d bytes\n\n", len(bytecode))
}

func cmdCheck() {
	nils := findNilFiles("src")
	if len(nils) == 0 {
		nils = findNilFiles(".")
	}
	ok := true
	for _, f := range nils {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("  ✗ %s (read error: %v)\n", f, err)
			ok = false
			continue
		}
		l := lexer.New(f, string(src))
		tokens := l.Tokenize()
		if len(l.Errors()) > 0 {
			fmt.Printf("  ✗ %s (lex errors: %s)\n", f, strings.Join(l.Errors(), "; "))
			ok = false
			continue
		}
		p := parser.New(f, tokens)
		prog := p.Parse()
		if len(p.Errors()) > 0 {
			fmt.Printf("  ✗ %s (parse errors: %s)\n", f, strings.Join(p.Errors(), "; "))
			ok = false
			continue
		}
		checker := types.New()
		checker.CheckProgram(prog)
		if len(checker.Errors()) > 0 {
			fmt.Printf("  ✗ %s (type errors: %s)\n", f, strings.Join(checker.Errors(), "; "))
			ok = false
		} else {
			fmt.Printf("  ✓ %s\n", f)
		}
	}
	if ok {
		fmt.Println("\n✓ Project type-checks OK (0 fatal errors)")
	} else {
		fmt.Println("\n✗ Type check failed")
		os.Exit(1)
	}
}

func cmdTest() {
	tests := findNilFiles("tests")
	if len(tests) == 0 {
		tests = findNilFiles(".")
	}
	var testFiles []string
	for _, f := range tests {
		if strings.HasSuffix(f, "_test.nil") {
			testFiles = append(testFiles, f)
		}
	}

	if len(testFiles) == 0 {
		fmt.Println("nil test: no test files (*_test.nil) found in tests/")
		return
	}

	fmt.Printf("nil test: running %d test files...\n\n", len(testFiles))
	passed := 0
	failed := 0

	for _, f := range testFiles {
		start := time.Now()
		mod, err := compileFile(f)
		if err != nil {
			fmt.Printf("  ✗ FAIL %s (compilation error: %v)\n", f, err)
			failed++
			continue
		}

		runner := vm.New(mod)
		if err := runner.Run(); err != nil {
			fmt.Printf("  ✗ FAIL %s (%v)\n", f, err)
			failed++
		} else {
			elapsed := time.Since(start)
			fmt.Printf("  ✓ PASS %s (%v)\n", f, elapsed)
			passed++
		}
	}

	fmt.Printf("\n=========================================\n")
	fmt.Printf("Test Results: %d passed, %d failed, %d total\n", passed, failed, len(testFiles))
	fmt.Printf("=========================================\n")
	if failed > 0 {
		os.Exit(1)
	}
}

func cmdFmt() {
	nils := findNilFiles(".")
	if len(nils) == 0 {
		fmt.Println("nil fmt: no .nil files found")
		return
	}
	formattedCount := 0
	for _, f := range nils {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		formatted := formatter.Format(string(content))
		if formatted != string(content) {
			_ = os.WriteFile(f, []byte(formatted), 0644)
			fmt.Printf("  ✓ formatted %s\n", f)
			formattedCount++
		} else {
			fmt.Printf("  • %s (already formatted)\n", f)
		}
	}
	fmt.Printf("\n✓ %d files formatted\n", formattedCount)
}

func cmdAdd(pkg string) {
	manifestPath := "alap.yaml"
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		if _, err2 := os.Stat("nilx.yaml"); err2 == nil {
			manifestPath = "nilx.yaml"
		} else {
			fmt.Printf("nil add: no alap.yaml found in current directory\n")
			return
		}
	}
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("nil add: cannot read %s: %v\n", manifestPath, err)
		return
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inDep := false
	added := false

	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "dependencies:") {
			inDep = true
			newLines = append(newLines, l)
			newLines = append(newLines, fmt.Sprintf("  %s: \"^0.1.0\"", pkg))
			added = true
			continue
		}
		newLines = append(newLines, l)
	}

	if !inDep || !added {
		newLines = append(newLines, "dependencies:", fmt.Sprintf("  %s: \"^0.1.0\"", pkg))
	}

	_ = os.WriteFile(manifestPath, []byte(strings.Join(newLines, "\n")), 0644)
	fmt.Printf("✓ Added dependency %s to %s\n", pkg, manifestPath)
}

func cmdClean() {
	_ = os.RemoveAll("build")
	fmt.Println("✓ Cleaned build/")
}

// ─── Utilities ────────────────────────────────────────────────────────────────

func findNilFiles(dir string) []string {
	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".nil") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func compileEntryBytecode() ([]byte, error) {
	nils := findNilFiles("src")
	if len(nils) == 0 {
		nils = findNilFiles(".")
	}
	if len(nils) == 0 {
		return nil, fmt.Errorf("no .nil entry file found")
	}
	mod, err := compileFile(nils[0])
	if err != nil {
		return nil, err
	}
	return codegen.Serialize(mod), nil
}
