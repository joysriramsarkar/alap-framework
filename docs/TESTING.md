# Alap Framework — Testing

## Commands

```bash
go build ./...          # compile every package
go vet ./...            # static analysis (must be clean)
go test ./...           # unit + integration tests
go test -race ./...     # race detector (needs a C toolchain)
go test -cover ./...    # coverage summary
```

CI (`.github/workflows/ci.yml`) runs, on Linux/macOS/Windows:

1. `go mod verify`
2. `go vet ./...`
3. `go test -v ./...`
4. a build of `bin/nil`, `bin/nilc`, `bin/nills`

`go vet` is a required gate: a new warning fails CI.

---

## Test layout

| Area | Test files |
|------|-----------|
| Compiler | `compiler/{lexer,parser,types,codegen,hir,formatter,lsp}/*_test.go` |
| Runtime | `runtime/{vm,actor,gc,scheduler}/*_test.go`, `runtime/nilrt/pipeline_test.go` |
| UI | `ui/{engine,layout,render,state,theme,widgets}/*_test.go` |
| Web | `web/{router,middleware,server,ssr,sse,websocket}/*_test.go` |
| Data | `data/{orm,migration}/*_test.go`, `stdlib/nil/nil_test.go` |
| Platforms | `platform/{android,ios,linux,onuron,web}/*_test.go` |
| Package manager | `pkg/manager/manager_test.go` |
| ABI | `abi/nilabi_test.go` |
| End-to-end | `tests/integration_test.go`, `examples/pos/*_test.go` |

---

## Coverage map

Every package below is exercised by at least one test.

| Package | Covered |
|---------|:-------:|
| `abi` | ✅ |
| `auth` | ✅ |
| `compiler/lexer` | ✅ |
| `compiler/parser` | ✅ |
| `compiler/types` | ✅ |
| `compiler/codegen` | ✅ |
| `compiler/hir` | ✅ |
| `compiler/formatter` | ✅ |
| `compiler/lsp` | ✅ |
| `runtime/vm` | ✅ |
| `runtime/actor` | ✅ |
| `runtime/gc` | ✅ |
| `runtime/scheduler` | ✅ |
| `runtime/nilrt` | ✅ |
| `ui/engine` | ✅ |
| `ui/layout` | ✅ |
| `ui/render` | ✅ |
| `ui/state` | ✅ |
| `ui/theme` | ✅ |
| `ui/widgets` | ✅ |
| `web/router` | ✅ |
| `web/middleware` | ✅ |
| `web/server` | ✅ |
| `web/ssr` | ✅ |
| `web/sse` | ✅ |
| `web/websocket` | ✅ |
| `data/orm` | ✅ |
| `data/migration` | ✅ |
| `stdlib/json` | ✅ |
| `stdlib/math` | ✅ |
| `stdlib/nil` | ✅ |
| `pkg/manager` | ✅ |
| `platform/android` | ✅ |
| `platform/ios` | ✅ |
| `platform/linux` | ✅ |
| `platform/onuron` | ✅ |
| `platform/web` | ✅ |

### Known gaps (no dedicated tests yet)

| Package | Why |
|---------|-----|
| `cmd/{nil,nilc,nills}` | CLI entry points; exercised indirectly via `tests/`. |
| `compiler/ast` | Pure data types (nodes), no behaviour. |
| `runtime/repl` | Interactive loop; needs a scripted stdin harness. |
| `stdlib/{async,core,crypto,database,fs,log,net,time}` | Library bindings; exercised through the VM. |
| `ui/animation` | Easing/interpolation helpers; no tests yet. |
| `platform/onuron/{nilbus,nilhal,nilui}` | Host-hardware shims; only meaningful on a real device. |
| `examples/{calculator,counter,hello,web-app}` | Sample source; `examples/pos` has tests. |

These gaps are tracked in [`ROADMAP.md`](./ROADMAP.md).

---

## Writing a test

Conventional Go tests, one package at a time. The end-to-end pattern from
`tests/integration_test.go` compiles and runs a snippet through the whole
pipeline:

```go
l := lexer.New("test.nil", src)
tokens := l.Tokenize()

p := parser.New("test.nil", tokens)
prog := p.Parse()

gen := codegen.New("test")
gen.GenerateProgram(prog)
blob := codegen.Serialize(gen.Module())

// execute blob with runtime/vm (or runtime/nilrt for the full pipeline)
```

Keep tests deterministic. The only concurrency-sensitive tests intentionally
block a worker (see `runtime/scheduler/scheduler_test.go`) or a mailbox
(`runtime/actor/actor_test.go`) to prove timeout / error paths.
