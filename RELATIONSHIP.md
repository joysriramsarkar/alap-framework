# Alap Framework — Relationship to NilLang

**Status:** Early-Alpha Prototype  
**Last Updated:** 2026-10-12

---

## 1. Executive Summary

The alap-framework is an **early-alpha prototype** of the Alap UI framework.
It shares conceptual DNA with nilLang's `pkg/alap/ui` but has a separate
compiler and VM. It is **NOT** the canonical implementation — nilLang is.

---

## 2. What is Alap Framework?

Alap Framework is a cross-platform application framework that provides:

- **Declarative UI** — ArkTS/ArkUI-inspired component hierarchies
- **Static Typing** — TypeScript-style type annotations
- **Multi-Platform Adapters** — Android (JNI/NDK), iOS (Metal/Swift), Linux (Wayland/X11/Flatpak)
- **Web Server** — HTTP, WebSocket, SSE, SSR, middleware
- **ORM** — Database migrations and query builder
- **Package Manager** — NilPM with alap.yaml / alap.lock

---

## 3. Relationship to NilLang

### 3.1 Shared Concepts

| Concept | nilLang | Alap Framework |
|---------|---------|----------------|
| Declarative UI | `pkg/alap/ui` | `ui/` package |
| Component model | `component` keyword | `ui/engine/` |
| State management | `state` keyword | `ui/state/` |
| Reactive signals | `signal()`, `sset()`, `sget()` | `ui/state/reconciler.go` |
| Platform adapters | `pkg/mobile/` | `platform/` |
| Web server | `pkg/alap/server` | `web/server/` |
| ORM | `pkg/alap/data` | `data/orm/` |

### 3.2 Key Differences

| Aspect | nilLang (Canonical) | Alap Framework |
|--------|---------------------|----------------|
| **Compiler** | `compiler/` — mature, 3 engines | `compiler/` — early-alpha |
| **VM** | `compiler/vm/` — stack-based bytecode | `runtime/vm/` — NABC bytecode |
| **Type System** | Full type checker with diagnostics | Basic type annotations |
| **Error Handling** | `Result<T,E>`, `?T`, categories | Not implemented |
| **Concurrency** | `async`/`await`, `task`, channels | Actor model |
| **Capabilities** | `requires`, `pure`, `unsafe` | Not implemented |
| **Standard Library** | Full stdlib | Basic stdlib |
| **Maturity** | Production | Early-alpha |

---

## 4. Maturity Comparison

### nilLang (Go) — Production

- ✅ 3 execution engines (evaluator, VM, AOT/LLVM)
- ✅ Full type checker with diagnostics
- ✅ Complete standard library
- ✅ Cross-engine conformance tests
- ✅ WASM target
- ✅ Mobile runtime (Android, iOS, HarmonyOS)
- ✅ Package manager (nilpkg)
- ✅ LSP server

### Alap Framework — Early-Alpha

- ✅ Basic compiler (lexer, parser, AST)
- ✅ Basic VM (NABC bytecode)
- ✅ UI engine (layout, render, state)
- ✅ Platform adapters (Android, iOS, Linux, Onuron)
- ✅ Web server (HTTP, WebSocket, SSE, SSR)
- ✅ ORM and migrations
- ✅ Package manager (NilPM)
- ❌ No type checker
- ❌ No error handling
- ❌ No concurrency
- ❌ No capabilities
- ❌ Limited standard library

---

## 5. The Path Forward

### 5.1 Recommended Direction

The alap-framework should **import nilLang packages** rather than maintain a
separate compiler and VM. This would:

1. Eliminate duplication of effort
2. Give alap-framework access to nilLang's mature type system
3. Provide a single canonical implementation
4. Allow alap-framework to focus on its unique value: platform adapters

### 5.2 Migration Steps

1. **Phase 1** — alap-framework imports nilLang as a library
   - Replace `compiler/` with nilLang's compiler
   - Replace `runtime/vm/` with nilLang's VM
   - Keep `ui/`, `platform/`, `web/`, `data/` packages

2. **Phase 2** — Merge unique alap-framework features into nilLang
   - Platform adapters → `pkg/mobile/`
   - Web server → `pkg/alap/server/`
   - ORM → `pkg/alap/data/`
   - Package manager → `pkg/nilpkg/`

3. **Phase 3** — Deprecate alap-framework's separate compiler/VM
   - All compilation through nilLang
   - alap-framework becomes a UI framework package

### 5.3 Alternative: Merge

Alternatively, alap-framework could be **merged into nilLang** as a set of
packages:

```text
nilLang/
├── pkg/
│   ├── alap/          # Already exists — UI framework
│   ├── mobile/        # Already exists — platform adapters
│   ├── nilpkg/        # Already exists — package manager
│   └── ...
└── ...
```

---

## 6. What alap-framework Does Well

These features are worth preserving and should be migrated to nilLang:

1. **Platform Adapters** — Android/iOS/Linux generators are mature and useful
2. **Web Server** — HTTP, WebSocket, SSE, SSR, middleware
3. **ORM** — Database migrations and query builder
4. **Package Manager** — NilPM with alap.yaml / alap.lock
5. **UI Engine** — Layout, render, state, animation, widgets

---

## 7. Conclusion

- **nilLang is the canonical implementation** of the NilLang language
- **alap-framework is a separate early-alpha prototype** with its own compiler/VM
- **alap-framework's value** is in its platform adapters, web server, and ORM
- **Future direction**: alap-framework should import nilLang packages or be merged

---

## 8. References

- [nilLang Repository](../nilLang/)
- [nilLang Language Specification](../nilLang/docs/spec/LANGUAGE_SPEC.md)
- [nilLang Canonical Spec](../nilLang/docs/spec/LANGUAGE_SPEC_CANONICAL.md)
- [OnuronOS Project](../onuronOS/)
- [Ecosystem Maturity Map](../onuron_project/docs/ECOSYSTEM_MATURITY.md)
