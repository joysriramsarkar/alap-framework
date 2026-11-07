# Alap Framework — CLI Reference

> **Naming:** these binaries are `alap`/`alapc`/`alap-lsp`. The name `nil` belongs
> to the canonical NilLang toolchain (`nilLang` repo); Alap renamed its binaries
> so both toolchains can coexist on one `PATH`.

The toolchain ships three binaries:

| Binary | Purpose |
|--------|---------|
| `alap` | Project management CLI (init, run, build, test, fmt, repl, pm…). |
| `alapc` | Standalone NilLang compiler (`source.nil` → `.nabc`). |
| `alap-lsp` | Language Server Protocol daemon for editors. |

Build them all:

```bash
go build -o bin/alap   ./cmd/nil
go build -o bin/alapc  ./cmd/nilc
go build -o bin/alap-lsp ./cmd/nills
```

---

## `alap` — project CLI

```text
alap init <name>          create a new Alap project
alap run [file]           compile and run
alap build <platform>     build for platform: onuron|android|ios|linux|web
alap check                type-check all .nil files
alap fmt                  format all .nil files
alap test                 run tests
alap repl                 interactive REPL
alap lsp                  run the language server
alap pm <subcommand>      package manager (install, add, remove, list)
alap add <pkg>            add a dependency
alap clean                clean build artifacts
alap version              print version
alap help                 print usage
```

### `alap init <name>`

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

### `alap run [file]`

Compiles and executes a `.nil` file (defaults to the manifest entry). Prints
program output and, if UI nodes were built, reports the UI tree.

### `alap build <platform>`

Compiles the project and hands the NABC bytecode to the matching platform
adapter's `GenerateProject`:

| Target | Adapter | Output |
|--------|---------|--------|
| `onuron` (alias `nilos`) | `platform/onuron` | Onuron `.nilax` bundle + native runtime |
| `android` | `platform/android` | Android Studio / Gradle project + JNI bridge |
| `ios` | `platform/ios` | iOS Swift / Metal bundle |
| `linux` | `platform/linux` | AppImage / Flatpak desktop bundle |
| `web` | `platform/web` | Web bundle: `build/web/index.html` + `main.nabc` (NABC bytecode, not WebAssembly) |

### `alap check`

Runs the type checker across the project's `.nil` files and prints
diagnostics. Exits non-zero if any file has a type error; `alap run`,
`alap build`, and `alapc` likewise abort on type errors.

### `alap fmt`

Runs the canonical formatter (`compiler/formatter`) over `.nil` files.

### `alap test`

Compiles and runs the project's test entry points.

### `alap repl`

Starts the interactive REPL (`runtime/repl`).

### `alap lsp`

Starts the language server (`compiler/lsp`).

### `alap pm` / `alap add`

Manage dependencies through `pkg/manager` (NilPM). Only
`alap pm install` is wired up (other subcommands are rejected);
`alap add <pkg>` adds a dependency. Lockfile entries label integrity as
`unresolved`: no artifact is downloaded or hashed, so the lock provides no
integrity guarantee yet.

---

## `alapc` — standalone compiler

```text
alapc [flags] [input.nil] [output.nabc]

  -in <file>     Input .nil source file
  -out <file>    Output .nabc bytecode file
  -run           Compile and run immediately
  -check         Type-check only
  -dump          Print AST and bytecode disassembly
```

Examples:

```bash
alapc -check src/main.nil
alapc src/main.nil out/main.nabc
alapc -run src/main.nil
alapc -dump src/main.nil
```

The positional form `alapc in.nil out.nabc` is also accepted.

---

## `alap-lsp` — language server

Runs an LSP server over stdio. It publishes diagnostics and answers requests:

- `textDocument/completion`
- `textDocument/hover`
- `textDocument/definition`
- `textDocument/references`
- `textDocument/rename`
- `textDocument/documentSymbol`

```bash
alap-lsp --version
```

Point your editor's LSP client at `alap-lsp`.
