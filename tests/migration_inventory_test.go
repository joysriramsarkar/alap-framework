package tests

// TestCanonicalMigrationInventory — alap M4 convergence tracker
// (nilLang is the canonical implementation; RELATIONSHIP.md).
//
// Every alap example is compiled with the canonical nilLang frontend. A file
// the canonical toolchain accepts is one step closer to convergence; a file
// it rejects names the first divergence. When a convergence step lands, its
// row flips to wantOK=true (and the stale wantCode is deleted) — that is the
// whole migration, one row at a time.
//
// Scope separates what must converge from what is intentionally Alap:
//   - "core": the shared language surface — rows flip to wantOK as M4 lands.
//   - "ui": Alap's declarative UI DSL (components, build blocks, platform
//     views), which RELATIONSHIP.md keeps as alap's preserved value. These
//     rows pin the canonical *rejection* at the UI boundary and, via
//     TestUiExamplesRunOnAlap below, prove the Alap toolchain runs them.
//   - "stdlib": canonical language surface, but calling host builtins the
//     canonical stdlib does not have yet (P2 stdlib parity). Rows pin the
//     first unknown builtin; the Alap side is covered by the example's own
//     test suite (no extra runner — these suites own their fixtures).
//
// The table is exhaustive: adding an example without a row fails the count
// check below.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	nilfrontend "github.com/joysriramsarkar/nilLang/compiler/frontend"
)

func TestCanonicalMigrationInventory(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		file       string
		scope      string // "core" or "ui" (see above)
		wantOK     bool
		wantCode   string // first diagnostic code when wantOK is false
		alapMarker string // ui only: substring the Alap run output must contain
	}{
		// Row 1 (M4 done): canonical `fn` + `->` — migrates cleanly.
		{"examples/hello/main.nil", "core", true, "", ""},
		// Row 2 (M4 done): `fn` + `->` + `Int` + expression match arms.
		{"examples/calculator/main.nil", "core", true, "", ""},
		// Row 2 (M4 done): canonical `fn` + `->` + commas + named literals.
		{"examples/counter/main.nil", "core", true, "", ""},
		// Row 2 (M4 done): canonical `fn` + `->` + commas + named literals.
		{"examples/todo/main.nil", "core", true, "", ""},
		// Row 3 (M4 done): logic migrated (`fn`/`->`/`Int`/`Void`/`Array`/commas/named
		// literals); `component`/`build()` UI DSL stays Alap by design.
		{"examples/social/main.nil", "ui", false, "E0004", "=== Alap Social Feed ==="},
		// Row 3 (M4 done): `fn` migrated; `import`-from-`alap/web` + `@Component` +
		// `build()` UI surface stays Alap by design.
		{"examples/web-app/main.nil", "ui", false, "E0001", "Alap Web Client App Booted"},
		// Row 4 (M4 done): fully core-spelled; blocked on `db_*` host builtins
		// (P2 stdlib parity). Alap side covered by examples/pos (11 tests).
		{"examples/pos/src/main.nil", "stdlib", false, "E0102", ""},
		// Row 4 (M4 done): ArkUI declarative UI — Alap by design; parsed and
		// compiled by examples/pos, canonical rejects the component surface.
		{"examples/pos/src/pos_app.nil", "ui", false, "E0001", ""},
	}

	const wantFiles = 8
	var found []string
	err = filepath.Walk(filepath.Join(root, "examples"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".nil" {
			rel, _ := filepath.Rel(root, path)
			found = append(found, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != wantFiles {
		t.Fatalf("example inventory drifted: got %d .nil files %v, want %d table rows — add a row per file",
			len(found), found, wantFiles)
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			if tc.scope != "core" && tc.scope != "ui" && tc.scope != "stdlib" {
				t.Fatalf("row %q has scope %q, want core, ui or stdlib", tc.file, tc.scope)
			}
			src, err := os.ReadFile(filepath.Join(root, tc.file))
			if err != nil {
				t.Fatal(err)
			}
			res := nilfrontend.Check(string(src), nilfrontend.Options{Filename: tc.file})
			if tc.wantOK {
				if !res.OK {
					t.Fatalf("expected canonical accept, got %s", res.ErrSummary())
				}
				return
			}
			if res.OK {
				t.Fatalf("expected canonical rejection (%s) — convergence landed? flip wantOK to true", tc.wantCode)
			}
			got := ""
			if len(res.Diagnostics) > 0 {
				got = res.Diagnostics[0].Code
			} else if len(res.ParseErrors) > 0 {
				got = "parse"
			}
			if got != tc.wantCode {
				t.Fatalf("first diagnostic drifted: got %q, want %q — update the row (progress or regress?)", got, tc.wantCode)
			}
		})
	}
}

// TestUiExamplesRunOnAlap proves the ui-scoped rows run on the Alap
// toolchain: the canonical rejection above is a boundary, not breakage.
func TestUiExamplesRunOnAlap(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	ui := []struct {
		file   string
		marker string
	}{
		// Row 3 (M4 done): logic migrated to canonical spelling; UI DSL stays.
		{"examples/social/main.nil", "=== Alap Social Feed ==="},
		// Row 4 (M4 done): `fn` migrated; web UI surface stays.
		{"examples/web-app/main.nil", "Alap Web Client App Booted"},
		// Row 4 (M4 done): declarative UI with no entry point — compiling
		// clean (parse+codegen+run without error) is the proof; behavior is
		// covered by examples/pos.
		{"examples/pos/src/pos_app.nil", ""},
	}

	for _, tc := range ui {
		t.Run(tc.file, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(root, tc.file))
			if err != nil {
				t.Fatal(err)
			}
			out := compile(t, string(src))
			if tc.marker != "" && !strings.Contains(out, tc.marker) {
				t.Fatalf("alap run output missing %q, got:\n%s", tc.marker, out)
			}
		})
	}
}
