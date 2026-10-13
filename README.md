# Alap Framework & NilLang

<p align="center">
  <b>A modern cross-platform application framework and language for Onuron, Android, iOS, and Linux.</b>
</p>

<p align="center">
  <i>Status: early-alpha prototype · Go 1.25+ · AGPL-3.0 (see <code>LICENSE</code>)</i>
</p>

---

> **সংক্ষেপে (Bengali):** আলাপ একটি ক্রস-প্ল্যাটফর্ম ডিক্লেয়ারেটিভ UI ও অ্যাপ্লিকেশন
> ফ্রেমওয়ার্ক, যা NilLang নামক একটি ভাষা দিয়ে চলে (লেক্সার → পার্সার → টাইপ-চেক
> → NABC বাইটকোড → VM → UI ট্রি)। এটি অনুরণ ইকোসিস্টেমের একটি **আর্লি-আলফা
> প্রোটোটাইপ** — canonical NilLang বাস্তবায়ন `nilLang` রিপোজিটরিতে (Go)। বিস্তারিত
> সম্পর্ক: [`RELATIONSHIP.md`](./RELATIONSHIP.md)।

---

## Overview

**Alap** is a cross-platform application framework powered by the **NilLang**
language. It combines:

- **ArkTS / ArkUI-inspired declarative UI** — reactive component hierarchies.
- **TypeScript-style static typing** — `let x: i32 = 1`, generics, structs, enums.
- **Go-inspired concurrency** — tasks, actors, scheduler, channels (`Channel<T>`).
- **Multi-platform native adapters** — Onuron, Android (JNI/NDK), iOS
  (Metal/Swift), and Linux (Wayland/X11/AppImage/Flatpak).
- **A full web framework** — radix routing, middleware, SSR, SSE, WebSocket.
- **An enterprise data layer** — query builder, transactions, migrations.

The toolchain compiles `.nil` source through a conventional pipeline:

```text
source.nil ──lexer──▶ tokens ──parser──▶ AST ──type-check──▶ typed AST
           ──codegen──▶ NABC bytecode ──vm──▶ program output + UI tree
```

---

## Maturity & stewardship

| Aspect | Reality |
|--------|---------|
| **Maturity** | Early-alpha / experimental. Interfaces may change. |
| **Canonical language** | **`nilLang` (Go)** — this repo's compiler is *not* canonical. |
| **Relationship** | See [`RELATIONSHIP.md`](./RELATIONSHIP.md) and [`docs/ROADMAP.md`](./docs/ROADMAP.md). |
| **Known gaps** | No full type checker enforcement at runtime, limited error handling, limited standard library, its **own** compiler/VM today. |

Planned direction: converge on the canonical nilLang compiler/runtime step by
step while keeping Alap's unique value (UI engine, platform adapters, web
server, ORM, package manager).

---

## Repository layout

```text
alap-framework/
├── abi/                # Stable C ABI boundary + cgo bridge (nilabi.h)
├── auth/               # JWT, password hashing, session helpers
├── cmd/
│   ├── nil/            # Alap project CLI (init, run, build, test, fmt, repl, pm, …)
│   ├── nilc/           # NilLang bytecode compiler CLI
│   └── nills/          # Language Server Protocol daemon (LSP)
├── compiler/
│   ├── ast/            # Abstract Syntax Tree nodes
│   ├── codegen/        # NABC bytecode generator & serializer
│   ├── formatter/      # Canonical code formatter
│   ├── hir/            # High-level IR & constant folding optimizer
│   ├── lexer/          # Tokenizer
│   ├── lsp/            # Language server engine (diagnostics, completion, hover)
│   ├── parser/         # Recursive-descent parser
│   └── types/          # Static type checker
├── data/
│   ├── migration/      # Versioned schema migrations
│   └── orm/            # SQL query builder, connection pool, transactions
├── pkg/manager/        # NilPM package & lockfile manager (alap.yaml / alap.lock)
├── platform/
│   ├── android/        # Android Studio / Gradle project generator & JNI bridge
│   ├── ios/            # iOS Swift / Metal view hierarchy generator
│   ├── linux/          # Linux AppImage / Flatpak desktop packager
│   └── onuron/         # Onuron native `.nilax` bundle + Vulkan/Wayland adapter
├── runtime/
│   ├── actor/          # Actor concurrency & mailbox message passing
│   ├── gc/             # Allocation tracking & telemetry
│   ├── nilrt/          # End-to-end pipeline runner
│   ├── repl/           # Interactive REPL
│   ├── scheduler/      # Cooperative task scheduler
│   └── vm/             # Stack-based NABC bytecode virtual machine
├── stdlib/             # core, json, time, math, fs, net, crypto, log, async, database
├── ui/                 # Declarative UI engine (engine, layout, state, render, theme, widgets, animation)
├── web/                # HTTP server, radix router, middleware, SSR, SSE, WebSocket
├── examples/           # calculator, counter, hello, pos, social, todo, web-app
├── tests/              # End-to-end integration tests
└── docs/               # In-depth guides (architecture, CLI, language, platforms, web, data, testing)
```

---

## Requirements

- **Go 1.25+** (the module declares `go 1.25.0`; CI uses `1.25.x`).
- Optional: a C toolchain (`gcc`/`clang`) if you want the cgo `abi` package built.

---

## Quick start

### 1. Build the toolchain

```bash
go build -o bin/nil ./cmd/nil
go build -o bin/nilc ./cmd/nilc
go build -o bin/nills ./cmd/nills
```

### 2. Create and run a project

```bash
nil init my-app
cd my-app
nil run            # compile & run
nil repl           # interactive REPL
```

### 3. Build for a platform

```bash
nil build onuron   # Onuron native package
nil build android  # Android Studio Gradle project
nil build ios      # iOS Xcode bundle
nil build linux    # Linux AppImage / Flatpak bundle
```

### 4. Developer tools

```bash
nil check          # type-check all .nil files
nil fmt            # format source files
nil test           # run tests
nil pm install     # resolve & lock dependencies
```

Full command reference: [`docs/CLI.md`](./docs/CLI.md).

---

## Language at a glance

```nil
// Functions
function add(a: i32, b: i32): i32 {
    return a + b
}

// Variables, arrays, control flow
let nums: i32[] = [10, 20, 30]
let sum: i32 = 0
let i: i32 = 0
while i < nums.length {
    sum = sum + nums[i]
    i = i + 1
}
print("Sum: " + sum.toString())

// Declarative UI component
@Entry
@Component
component App {
    @State count: i32 = 0

    build() {
        Column {
            Text("Count: " + count.toString())
            Button("+") {
                onClick => { count += 1 }
            }
        }
    }
}
```

More: [`docs/LANGUAGE.md`](./docs/LANGUAGE.md).

---

## Examples

| Example | Path | Demonstrates |
|---------|------|--------------|
| Hello | `examples/hello/main.nil` | `print`, strings |
| Calculator | `examples/calculator/main.nil` | arithmetic, functions |
| Counter | `examples/counter/main.nil` | state, UI tree |
| Todo | `examples/todo/main.nil` | arrays, loops |
| Social | `examples/social/main.nil` | structs, records |
| POS | `examples/pos/` | multi-file app, ORM, tests |
| Web app | `examples/web-app/` | HTTP server + routes |

---

## Testing

```bash
go build ./...     # compile everything
go vet ./...       # static checks (clean)
go test ./...      # unit + integration tests
```

Every package is expected to build and vet cleanly; see
[`docs/TESTING.md`](./docs/TESTING.md) for the coverage map.

---

## Relationship to nilLang

`nilLang` (Go) is the **canonical** language implementation. This repository is a
separate early-alpha prototype with its own compiler and VM that must converge
on the canonical toolchain. The full analysis, comparison tables, and migration
plan live in [`RELATIONSHIP.md`](./RELATIONSHIP.md).

---

## Documentation

| Document | Contents |
|----------|----------|
| [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) | Components, data flow, and boundaries |
| [`docs/CLI.md`](./docs/CLI.md) | `nil`, `nilc`, `nills` command reference |
| [`docs/LANGUAGE.md`](./docs/LANGUAGE.md) | NilLang (Alap dialect) syntax reference |
| [`docs/PLATFORMS.md`](./docs/PLATFORMS.md) | Android, iOS, Linux, Onuron adapters |
| [`docs/WEB.md`](./docs/WEB.md) | Server, radix router, middleware, SSR/SSE/WebSocket |
| [`docs/DATA.md`](./docs/DATA.md) | ORM, transactions, migrations |
| [`docs/TESTING.md`](./docs/TESTING.md) | Build, vet, test, and coverage map |
| [`docs/CONTRIBUTING.md`](./docs/CONTRIBUTING.md) | Contribution workflow |
| [`docs/ROADMAP.md`](./docs/ROADMAP.md) | Convergence roadmap |
| [`RELATIONSHIP.md`](./RELATIONSHIP.md) | Alap ↔ nilLang relationship & migration plan |
| [`Alap_Framework_NilLang_Blueprint.md`](./Alap_Framework_NilLang_Blueprint.md) | Original long-form blueprint |
| [`Alap_NilLang_Complete_Documentation_BN_EN.md`](./Alap_NilLang_Complete_Documentation_BN_EN.md) | Bilingual complete documentation |

---

## License

AGPL-3.0 © 2026 Joy Sarkar / Onuron Project. See [`LICENSE`](./LICENSE).
