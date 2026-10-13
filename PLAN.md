# alap-framework — প্ল্যান (UI + অ্যাডাপ্টার)

> ভূমিকা: Declarative UI, platform adapters, web, ORM, NilPM। **নিজস্ব compiler/VM = early-alpha, non-canonical** (`RELATIONSHIP.md`)।
> মূল্য: `ui/`, `platform/`, `web/`, `data/`, NilPM — এগুলো রক্ষা করে nilLang-এ একীভূত হওয়া (M4)।

## বর্তমান অবস্থা
- Pipeline সম্পূর্ণ: lexer→parser→types (non-fatal)→HIR→NABC→VM→UI tree। `go test ./...` সবুজ (সার্ভে অনুযায়ী)।
- UI tree/layout/state/render আছে; reconciler diff, layout wrap/align, animation কভারেজ অসম্পূর্ণ।
- Platform = generator-only (Gradle/JNI, Swift/Metal, AppDir, Onuron scaffold); device/store validation বাকি।
- `examples/pos` (mini-pos 1.0.0) একমাত্র টেস্টেড উদাহরণ; `build/` + `pos.db` কমিটেড (hygiene issue); অন্য উদাহরণে টেস্ট নেই।

## M1 — কভারেজ (প্রথম)
- `runtime/repl`, `ui/animation`, `stdlib/*`, `cmd/*`-এ scripted-stdin + unit test। কমিটেড বাইনারি (`pos.db`, `examples/pos/pos.db`, `build/*/main.nabc`) `.gitignore`-এ সরানো।

## M3 — UI ইঞ্জিন
- `ui/state/reconciler.go`: keyed list diff + patch; `ui/layout/layout.go`: constraints/wrap/align + JSON golden (পিক্সেল নয়)।
- `examples/pos/src/pos_app.nil` → `ui/engine/tree.go` দিয়ে end-to-end POS রেফারেন্স।
- বাংলা টেক্সট: shaping/input/i18n টর্চার-স্যুট (পরে); a11y semantics শুরু থেকে।

## M4 — Convergence (M3-এর আগে শুরু)
- Feature-flag দিয়ে canonical nilLang frontend; `compiler/types` enforcement; `Result/?`/capability গ্রহণ; duplicate VM deprecate। NABC বনাম canonical bytecode = benchmark দিয়ে সিদ্ধান্ত।

## M5 — প্ল্যাটফর্ম যাচাই
- `platform/onuron|android|ios|linux`: real toolchain বিল্ড (Android SDK/Gradle, Xcode, AppImage), signed `.nilax`, Linux headless smoke; CI gate। Android ABI prefix এক করা (`nilx_` বনাম `alap_` mismatch যাচাই)।

## M6 — web/data
- WS-upgrade/TLS, real SQLite/Postgres driver, rate-limiter load/soak।

## DoD
- `go test ./...` + `go vet` + `golangci-lint`; উদাহরণ-অ্যাপ headless render golden; `docs/*.md` ক্রস-লিংক।

## এই সেশনে
- [x] 13 Oct: `go test ./...` সবুজ (Phase 0)।
- [x] 13 Oct: M4 start — dependency + inventory (commit `6aa0dc5`)।
- [x] 13 Oct: M4 row 1 (hello) — alap parser-এ `fn` + `->` (backward compat), hello migrate, দুই toolchain সবুজ, inventory flip — নিচে।
