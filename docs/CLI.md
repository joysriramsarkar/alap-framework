# Alap Framework — CLI Reference

The toolchain ships three binaries:

| Binary | Purpose |
|--------|---------|
| `nil` | Project management CLI (init, run, build, test, fmt, repl, pm…). |
| `nilc` | Standalone NilLang compiler (`source.nil` → `.nabc`). |
| `nills` | Language Server Protocol daemon for editors. |

Build them all:

```bash
go build -o bin/nil   ./cmd/nil
go build -o bin/nilc  ./cmd/nilc
go build -o bin/nills ./cmd/nills
```

---

## `nil` — project CLI

```text
nil init <name>          create a new NilLang project
nil run [file]           compile and run
nil build <platform>     build for platform: onuron|android|ios|linux
nil check                type-check all .nil files
nil fmt                  format all .nil files
nil test                 run tests
nil repl                 interactive REPL
nil lsp                  run the language server
nil pm <subcommand>      package manager (install, add, remove, list)
nil add <pkg>            add a dependency
nil clean                clean build artifacts
nil version              print version
nil help                 print usage
```

### `nil init <name>`

Creates a project scaffold:

```text
my-app/
├── alap.yaml           # project manifest
└── src/
    ├── main.nil        # program entry point
    └── App.nil         # declarative UI root component
```

The generated `alap.yaml` declares the entry file, permissions, and per-target
settings:

```yaml
name: my-app
version: 0.1.0
entry: src/main.nil

permissions:
  - storage.read
  - notifications

targets:
  onuron:  { arch: arm64 }
  android: { minSdk: 26, targetSdk: 35 }
  ios:     { minVersion: "16.0" }
  linux:   { display: wayland }

dependencies: {}
```

### `nil run [file]`

Compiles and executes a `.nil` file (defaults to the manifest entry). Prints
program output and, if UI nodes were built, reports the UI tree.

### `nil build <platform>`

Compiles the project and hands the NABC bytecode to the matching platform
adapter's `GenerateProject`:

| Target | Adapter | Output |
|--------|---------|--------|
| `onuron` (alias `nilos`) | `platform/onuron` | Onuron `.nilax` bundle + native runtime |
| `android` | `platform/android` | Android Studio / Gradle project + JNI bridge |
| `ios` | `platform/ios` | iOS Swift / Metal bundle |
| `linux` | `platform/linux` | AppImage / Flatpak desktop bundle |

### `nil check`

Runs the type checker across the project's `.nil` files and prints diagnostics.

### `nil fmt`

Runs the canonical formatter (`compiler/formatter`) over `.nil` files.

### `nil test`

Compiles and runs the project's test entry points.

### `nil repl`

Starts the interactive REPL (`runtime/repl`).

### `nil lsp`

Starts the language server (`compiler/lsp`).

### `nil pm` / `nil add`

Manage dependencies through `pkg/manager` (NilPM): `install`, `add`, `remove`,
`list`, with an `Audit` check.

---

## `nilc` — standalone compiler

```text
nilc [flags] [input.nil] [output.nabc]

  -in <file>     Input .nil source file
  -out <file>    Output .nabc bytecode file
  -run           Compile and run immediately
  -check         Type-check only
  -dump          Print AST and bytecode disassembly
```

Examples:

```bash
nilc -check src/main.nil
nilc src/main.nil out/main.nabc
nilc -run src/main.nil
nilc -dump src/main.nil
```

The positional form `nilc in.nil out.nabc` is also accepted.

---

## `nills` — language server

Runs an LSP server over stdio. It publishes diagnostics and answers requests:

- `textDocument/completion`
- `textDocument/hover`
- `textDocument/definition`
- `textDocument/references`
- `textDocument/rename`
- `textDocument/documentSymbol`

```bash
nills --version
```

Point your editor's LSP client at `nills`.
