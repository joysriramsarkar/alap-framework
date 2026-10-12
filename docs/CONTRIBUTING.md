# Contributing to Alap Framework

Thanks for your interest in Alap! This repository is an **early-alpha
prototype** of the Alap UI framework and the NilLang dialect it runs.

Before contributing, please read:

- [`../RELATIONSHIP.md`](../RELATIONSHIP.md) — how Alap relates to the canonical
  `nilLang`, and why it is converging on it.
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — module boundaries.
- [`ROADMAP.md`](./ROADMAP.md) — where the project is heading.

---

## Ground rules

1. **Keep the boundaries.** Compiler stages stay in `compiler/*`, runtime in
   `runtime/*`, UI in `ui/*`, web in `web/*`, data in `data/*`, platforms in
   `platform/*`. Do not grow a second compiler/VM competing with canonical
   NilLang — move toward using it.
2. **No fabricated behaviour.** Do not claim a feature works unless a test
   proves it. Mark unfinished work as prototype in docs.
3. **Honest maturity labels.** Reference/experimental features must be labelled
   as such.
4. **Every package builds and vets cleanly:** `go build ./...` and
   `go vet ./...` must pass.
5. **Tests accompany behaviour.** New behaviour needs a test; bug fixes should
   add a regression test.

---

## Development workflow

```bash
# 1. build
go build ./...

# 2. static checks
go vet ./...

# 3. tests
go test ./...

# 4. format (gofmt is enforced)
gofmt -l .          # should print nothing
```

Commit messages follow `type(scope): summary`, e.g.
`fix(abi): avoid uintptr→unsafe.Pointer in cgo bridge`.

---

## Adding a compiler stage change

1. Update the AST (`compiler/ast`) if needed.
2. Update the parser (`compiler/parser`) and add a parse test.
3. Update the type checker (`compiler/types`) if the construct is typed.
4. Update codegen (`compiler/codegen`) and add a serialize/round-trip test.
5. Update the VM (`runtime/vm`) if a new opcode is needed.
6. Update [`LANGUAGE.md`](./LANGUAGE.md).

Every new opcode must be handled by both `Serialize` and `Deserialize`.

---

## Documentation

If you change user-visible behaviour, update the matching guide in `docs/`:

| Change | Update |
|--------|--------|
| CLI command/flag | [`CLI.md`](./CLI.md) |
| Language syntax | [`LANGUAGE.md`](./LANGUAGE.md) |
| Platform adapter | [`PLATFORMS.md`](./PLATFORMS.md) |
| Web stack | [`WEB.md`](./WEB.md) |
| Data layer | [`DATA.md`](./DATA.md) |
| Tests/CI | [`TESTING.md`](./TESTING.md) |

---

## License

By contributing you agree your contributions are licensed under the repository's
MIT license (see [`../LICENSE`](../LICENSE)). Note the ecosystem licensing
policy: `alap-framework` keeps its MIT carve-out, documented in
`onuron_project/docs/LICENSING.md`.
