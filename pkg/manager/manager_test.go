package manager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemverMatching(t *testing.T) {
	v1, err := ParseVersion("1.2.3")
	if err != nil || v1.Major != 1 || v1.Minor != 2 || v1.Patch != 3 {
		t.Fatalf("failed parsing 1.2.3: %v", err)
	}

	if !v1.MatchesConstraint("^1.2.0") {
		t.Errorf("expected 1.2.3 to match ^1.2.0")
	}
	if !v1.MatchesConstraint("~1.2.0") {
		t.Errorf("expected 1.2.3 to match ~1.2.0")
	}
	if v1.MatchesConstraint("^2.0.0") {
		t.Errorf("expected 1.2.3 not to match ^2.0.0")
	}
}

func TestPackageManagerFullLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nilpm_full_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manifest := `name: test-app
version: 0.1.0
dependencies:
  "@alap/ui-charts": "^1.2.0"
`
	err = os.WriteFile(filepath.Join(tempDir, "alap.yaml"), []byte(manifest), 0644)
	if err != nil {
		t.Fatalf("failed writing manifest: %v", err)
	}

	pm := New(tempDir)
	if err := pm.Install(); err != nil {
		t.Fatalf("pm.Install failed: %v", err)
	}
	// Integrity must be honestly labelled unresolved, never a fabricated
	// sha256 of the package name/version string.
	lockBytes, err := os.ReadFile(filepath.Join(tempDir, "alap.lock"))
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	if !strings.Contains(string(lockBytes), `integrity: "unresolved"`) || strings.Contains(string(lockBytes), "sha256-") {
		t.Errorf("lockfile must label integrity unresolved without a fabricated hash, got:\n%s", lockBytes)
	}

	// 1. Add dependency
	if err := pm.Add("@alap/sqlite", "^0.4.0"); err != nil {
		t.Fatalf("pm.Add failed: %v", err)
	}

	deps, err := pm.List()
	if err != nil || len(deps) != 2 {
		t.Errorf("expected 2 dependencies after add, got %d (err: %v)", len(deps), err)
	}

	// 2. Audit
	verified, issues, err := pm.Audit()
	if err != nil || verified != 2 || len(issues) > 0 {
		t.Errorf("audit failed: verified=%d, issues=%v, err=%v", verified, issues, err)
	}

	// 3. Remove dependency
	if err := pm.Remove("@alap/sqlite"); err != nil {
		t.Fatalf("pm.Remove failed: %v", err)
	}

	depsAfter, _ := pm.List()
	if len(depsAfter) != 1 {
		t.Errorf("expected 1 dependency after remove, got %d", len(depsAfter))
	}
}
