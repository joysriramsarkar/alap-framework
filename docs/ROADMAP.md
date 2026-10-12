# Alap Framework — Roadmap

**Status:** early-alpha prototype.
**Companion:** `onuron_project/docs/ROADMAP.md` (ecosystem-wide phases) and
[`../RELATIONSHIP.md`](../RELATIONSHIP.md) (Alap ↔ nilLang).

This roadmap is deliberately **verifiable**: each milestone is "done" only when a
test or CI gate proves it. A directory or a stub is not a milestone.

---

## Guiding direction

Alap's unique value is its **declarative UI engine, platform adapters, web
server, ORM, and package manager**. The canonical NilLang compiler/runtime lives
in the `nilLang` repository. Alap should therefore:

- keep and harden its UI/data/platform layers, and
- **stop growing a competing compiler/VM**, converging instead on the canonical
  NilLang toolchain.

---

## Milestone 1 — Release hygiene (in progress)

- [x] `go build ./...`, `go vet ./...`, `go test ./...` all pass.
- [x] `go vet` enforced in CI as a gate.
- [x] Release workflow uses the same Go version as CI (`1.25.x`).
- [x] `abi` cgo bridge is `unsafe.Pointer`-clean.
- [x] Core packages have unit tests (compiler stages, runtime, UI, web, data).
- [ ] Raise coverage on the known-gap packages (`runtime/repl`, `ui/animation`,
      `stdlib/*`).

## Milestone 2 — Documentation & honesty

- [x] Detailed `README.md` with architecture, CLI, language, testing.
- [x] `docs/` guides (architecture, CLI, language, platforms, web, data, testing,
      contributing, roadmap).
- [x] Explicit prototype labels; no fabricated capabilities.
- [ ] Cross-link every guide from the docs site (`onuron_project`).

## Milestone 3 — UI engine depth

- [ ] Full diff/patch reconciliation from `ui/state` with tests.
- [ ] Layout completeness (constraints, wrapping, alignment) with visual tests.
- [ ] Widget set parity with the canonical nilLang `pkg/alap/ui` widget list.
- [ ] Reference app (POS) rendered by the engine end-to-end.

## Milestone 4 — Convergence on canonical NilLang

- [ ] Replace the in-repo lexer/parser with the canonical nilLang front end
      (feature-tagged), keeping Alap's UI/HIR extensions.
- [ ] Align the standard library surface with canonical nilLang.
- [ ] Adopt the canonical error handling and capability model.
- [ ] Deprecate the duplicate VM in favour of the canonical runtime.

## Milestone 5 — Platform readiness

- [ ] Android/iOS adapters validated on real toolchains (build, not flash).
- [ ] Onuron adapter against the real `nilhal`/`nilui`/`nilbus`.
- [ ] Linux packaging smoke tests in CI.
- [ ] Signed `.nilax` bundle output.

## Milestone 6 — Web & data hardening

- [ ] WebSocket upgrade handling and TLS termination.
- [ ] Database drivers behind `data/orm` (real SQLite/Postgres).
- [ ] Load/soak tests for the router and rate limiter.

---

## Non-goals (for now)

- Being the canonical NilLang implementation.
- Shipping to production app stores.
- Flashing physical devices.
- Adding syntax without a spec change in canonical NilLang.
