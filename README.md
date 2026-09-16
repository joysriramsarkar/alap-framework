# Alap Framework & NilLang

<p align="center">
  <b>A modern cross-platform application framework and language for Onuron, Android, iOS, and Linux.</b>
</p>

---

## Overview

**Alap** is a next-generation application framework powered by **NilLang**—a language combining:
- **ArkTS / ArkUI-inspired Declarative UI** with reactive component hierarchies.
- **TypeScript-style static typing** and clear grammar.
- **Go-inspired Concurrency** with tasks, actors, and channels.
- **Multi-Platform Native Adapters** supporting Onuron (Onuron OS), Android (JNI/NDK), iOS (Metal/Swift), and Linux (Wayland/X11/Flatpak).

---

## Project Architecture

```text
alap-framework/
├── abi/            # Stable C ABI boundary & Cgo bridge (nilabi.h)
├── cmd/
│   ├── nil/        # Alap project CLI (init, run, build, test, fmt, repl, pm)
│   ├── nilc/       # NilLang bytecode compiler CLI
│   └── nills/      # Language Server Protocol daemon (LSP 3.17)
├── compiler/
│   ├── ast/        # Abstract Syntax Tree nodes
│   ├── codegen/    # NABC bytecode generator & serializer
│   ├── formatter/  # Canonical code formatter
│   ├── hir/        # High-level IR & constant folding optimizer
│   ├── lexer/      # Tokenizer
│   ├── lsp/        # Language server engine (diagnostics, completion, hover)
│   ├── parser/     # Recursive-descent parser
│   └── types/      # Static type checker
├── pkg/
│   └── manager/    # NilPM package & lockfile manager (alap.yaml / alap.lock)
├── platform/
│   ├── android/    # Android Studio / Gradle project generator & JNI bridge
│   ├── ios/        # iOS Swift / Metal layer view hierarchy generator
│   ├── linux/      # Linux AppImage / Flatpak desktop packager
│   └── onuron/     # Onuron native .nilax bundle & Vulkan / Wayland adapter
├── runtime/
│   ├── actor/      # Actor concurrency & mailbox message passing
│   ├── repl/       # Interactive REPL
│   └── vm/         # Stack-based NABC bytecode Virtual Machine
├── stdlib/         # Standard Library (core, json, time, math, fs, net, crypto, log, async)
├── ui/             # Declarative UI Engine (layout, state, theme, animation, widgets)
└── examples/       # Example apps (calculator, counter, hello, social, todo)
```

---

## Quick Start

### 1. Build the Toolchain
```bash
go build -o bin/nil ./cmd/nil
go build -o bin/nilc ./cmd/nilc
go build -o bin/nills ./cmd/nills
```

### 2. Create and Run a Project
```bash
# Initialize a new project
nil init my-app
cd my-app

# Run the app
nil run

# Interactive REPL
nil repl
```

### 3. Build for Multiple Platforms
```bash
# Onuron native package (.nilax)
nil build onuron

# Android Studio Gradle Project
nil build android

# iOS Xcode Bundle
nil build ios

# Linux Desktop Bundle (AppImage / Flatpak)
nil build linux
```

### 4. Developer Tools
```bash
# Format source files
nil fmt

# Type-check
nil check

# `nil check` exits non-zero when lexical, parse, or type errors are found.

# Run tests
nil test

# Resolve & lock dependencies
nil pm install
```

---

## License

MIT License © 2026 Joy Sarkar / Onuron Project
