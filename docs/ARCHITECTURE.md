# Alap Framework — Architecture

This document describes how the Alap framework is put together, module by
module, and where each responsibility lives.

> **Scope note:** this repository is an early-alpha prototype. The canonical
> NilLang implementation is [`nilLang`](../nilLang/). See
> [`../RELATIONSHIP.md`](../RELATIONSHIP.md) for the convergence plan.

---

## 1. The compilation pipeline

Alap compiles `.nil` source through a conventional, well-separated pipeline:

```text
source.nil
  │
  ▼  compiler/lexer      → []Token
  ▼  compiler/parser     → *ast.Program
  ▼  compiler/types      → diagnostics (fatal: the pipeline aborts on any type error)
  ▼  compiler/codegen    → *codegen.Module  (NABC bytecode)
  ▼  codegen.Serialize   → []byte           (portable .nabc)
  ▼  runtime/vm          → program output + *engine.UITree
```

Each stage is independently testable and occupies its own package.

| Package | Responsibility | Entry point |
|---------|----------------|-------------|
| `compiler/lexer` | Turn source text into tokens | `lexer.New(file, src).Tokenize()` |
| `compiler/parser` | Build the AST | `parser.New(file, tokens).Parse()` |
| `compiler/types` | Static type checking | `types.New().CheckProgram(prog)` |
| `compiler/codegen` | Emit NABC bytecode | `codegen.New(name).GenerateProgram(prog)` |
| `runtime/vm` | Execute NABC | `vm.New(module).Run()` |

There is no separate pipeline-runner package: `cmd/nilc` (`compile`) and
`cmd/nil` (`compileFile`) wire the stages together directly, and the
integration tests in `tests/` do the same. Any stage error — including a
type error — aborts compilation before codegen/execution.

---

## 2. NABC bytecode

`compiler/codegen` defines the **N**ilLang **A**pplication **B**yte**C**ode
format:

- `Opcode` — a `byte` enum (`OP_ADD`, `OP_JUMP`, `OP_UI_CREATE`, …).
- `Instruction` — `{ Op Opcode; Operand int32 }`.
- `Constant` — a tagged constant (`ConstInt`, `ConstFloat`, `ConstString`,
  `ConstBool`, `ConstNull`).
- `Function` — name, param/local counts, `[]Instruction`, constant pool.
- `Module` — name, `[]*Function`, globals, `MainFunc`.

`Serialize(mod)` produces a compact little-endian buffer prefixed with the
magic `NABC`; `Deserialize([]byte)` validates the magic and rebuilds the
module. This buffer is the artifact that crosses the platform adapters.

```go
mod := gen.Module()
blob := codegen.Serialize(mod)      // portable bytes
restored, err := codegen.Deserialize(blob)
```

---

## 3. The runtime

### Virtual machine (`runtime/vm`)

A stack-based interpreter over `Module`:

- **Values** are a tagged union `Value` (`ValInt`, `ValFloat`, `ValString`,
  `ValBool`, `ValArray`, `ValObject`, `ValFunc`, `ValChannel`, …) with
  constructors (`IntVal`, `StrVal`, `ArrayVal`, `ObjVal`, …) and operations
  (`String()`, `Truthy()`, `Equal()`).
- **Frames** hold locals, an instruction pointer, and a stack base.
- **Builtins** are registered up front and bridge to `stdlib/*`
  (`crypto.Sha256`, `math.Sqrt`, file IO, JSON, time, net, …) plus the UI
  instructions that build an `engine.UITree`.
- **Native functions** can be injected with `RegisterNative(name, fn)`.

### Concurrency primitives

| Package | Model |
|---------|-------|
| `runtime/actor` | Isolated actors with buffered mailboxes; `Send` (fire-and-forget) and `Ask` (reply awaited). |
| `runtime/scheduler` | Worker pool that multiplexes `Task`s; `Spawn` + `Await`/`AwaitTimeout`. |
| `runtime/gc` | Allocation tracking and telemetry (`TrackAlloc`, `TrackFree`, `Collect`, `GetStats`). |
| `vm.Channel` | Go-channel-backed `Channel<T>` (`NewChannel`, `Send`, `Recv`). |

### REPL (`runtime/repl`)

`repl.Start(in, out)` provides an interactive loop that compiles and runs
snippets, plus `:` commands.

---

## 4. The UI engine

`ui/*` implements the declarative, reactive layer:

| Package | Responsibility |
|---------|----------------|
| `ui/engine` | The tree model. `UITree` (`BeginNode`/`EndNode`/`SetProp`/`SetEvent`) and `WidgetNode` (props, children, bounds, events), plus hit-testing, touch dispatch, and JSON export. |
| `ui/layout` | Flex/leaf layout. `ComputeLayout(root, ctx)` assigns `Bounds`. |
| `ui/state` | Reactive primitives. `Signal` (`Get`/`Set`/`Subscribe`), `Store`, and a `Reconciler` that diffs old/new trees into mutations. |
| `ui/render` | Software rasteriser (`Canvas`, `DrawRect`, `DrawRoundedRect`, gradients, hex colour parsing). |
| `ui/theme` | `LightTheme()` / `DarkTheme()` colour palettes and tokens. |
| `ui/widgets` | Builder helpers: `Column`, `Row`, `Text`, `Button`, `TextInput`, `Image`, `Card`, `Divider`, `Spacer`, `Switch`. |
| `ui/animation` | Easing functions (`Linear`, `EaseInQuad`, `EaseOutQuad`, `EaseInOutCubic`, `Spring`) and `Interpolate`. |

The VM's UI instructions build the tree; the host renders it. For a platform
adapter this means: run bytecode → get `*engine.UITree` → render via the
platform (`nilui` for Onuron, JSON for Android/iOS native hosts).

---

## 5. Native embedding

There is no top-level abi (or auth) package in this repository. Native
embedding, where it exists, lives inside the generated projects: e.g. the
Android adapter (`platform/android`) emits its own JNI glue and a
`nilabi.h` header into the generated project's `cpp/` directory. That
generated header is a scaffold, not a stable, versioned C ABI package.

---

## 6. Platform adapters

`platform/*` generate host projects or drive a host runtime. See
[`PLATFORMS.md`](./PLATFORMS.md) for details. Every single-target adapter
exposes `New()` and `GenerateProject(outputDir, bytecode)`.

---

## 7. Web and data

- `web/*` — HTTP server, radix router, middleware chain, SSR, SSE, WebSocket.
  See [`WEB.md`](./WEB.md).
- `data/*` and `stdlib/database` — prototype query builder (identifier /
  operator allowlist, parameterised values), an in-memory prototype
  transaction boundary with no real DB driver, and migrations. See
  [`DATA.md`](./DATA.md).

---

## 8. Package manager

`pkg/manager` implements **NilPM** against `alap.yaml` and `alap.lock`:
`Install`, `Add`, `Remove`, `List`, `Audit`, and `PackageArchive`.

---

## 9. Boundaries & non-goals

- Alap does **not** claim to be the canonical NilLang implementation.
- The type checker blocks execution: `alap check` exits non-zero and the
  `alap run` / build / `alapc` pipelines abort when any type error exists.
- Error handling, capabilities, and the full standard library are **incomplete**
  (tracked in [`ROADMAP.md`](./ROADMAP.md)).
- Hardware backends (GPU/audio/camera) are represented but not production-ready.
