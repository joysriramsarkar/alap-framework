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
// The table is exhaustive: adding an example without a row fails the count
// check below.

import (
	"os"
	"path/filepath"
	"testing"

	nilfrontend "github.com/joysriramsarkar/nilLang/compiler/frontend"
)

func TestCanonicalMigrationInventory(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		file     string
		wantOK   bool
		wantCode string // first diagnostic code when wantOK is false
	}{
		// Row 1 (M4 done): canonical `fn` + `->` — migrates cleanly.
		{"examples/hello/main.nil", true, ""},
		// Row 2 (M4 done): `fn` + `->` + `Int` + expression match arms.
		{"examples/calculator/main.nil", true, ""},
		// Row 2 (M4 done): canonical `fn` + `->` + commas + named literals.
		{"examples/counter/main.nil", true, ""},
		// Row 2 (M4 done): canonical `fn` + `->` + commas + named literals.
		{"examples/todo/main.nil", true, ""},
		// TODO(M4): component-call syntax differs from canonical `component`.
		{"examples/social/main.nil", false, "E0004"},
		// TODO(M4): web-app surface is outside the canonical grammar.
		{"examples/web-app/main.nil", false, "E0001"},
		// TODO(M4): same `function` spelling (reference POS app).
		{"examples/pos/src/main.nil", false, "E0004"},
		// TODO(M4): ArkUI `@Component`/`build()` vs canonical keywords.
		{"examples/pos/src/pos_app.nil", false, "E0001"},
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
