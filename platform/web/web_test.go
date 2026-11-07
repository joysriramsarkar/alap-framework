package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWebPlatformAdapter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "alap_web_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	adapter := New()
	bytecode := []byte("WASM_BYTECODE_DUMMY")

	err = adapter.GenerateProject(tempDir, bytecode)
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// Verify index.html
	indexFile, err := os.ReadFile(filepath.Join(tempDir, "web", "index.html"))
	if err != nil || !strings.Contains(string(indexFile), "AlapRuntime.boot") {
		t.Errorf("expected valid index.html with runtime boot, got: %s", string(indexFile))
	}

	// Verify runtime.js
	runtimeFile, err := os.ReadFile(filepath.Join(tempDir, "web", "js", "runtime.js"))
	if err != nil || !strings.Contains(string(runtimeFile), "AlapRuntime =") {
		t.Errorf("expected valid runtime.js, got: %s", string(runtimeFile))
	}

	// Verify manifest.json
	manifestFile, err := os.ReadFile(filepath.Join(tempDir, "web", "manifest.json"))
	if err != nil || !strings.Contains(string(manifestFile), "AlapWebApp") {
		t.Errorf("expected valid manifest.json, got: %s", string(manifestFile))
	}

	// Bytecode is NABC, not WebAssembly: it must be written as main.nabc
	// and never mislabelled as .wasm.
	nabcFile, err := os.ReadFile(filepath.Join(tempDir, "web", "main.nabc"))
	if err != nil || string(nabcFile) != string(bytecode) {
		t.Errorf("expected main.nabc containing the bytecode, got err=%v content=%q", err, string(nabcFile))
	}
	if _, err := os.Stat(filepath.Join(tempDir, "web", "alap.wasm")); !os.IsNotExist(err) {
		t.Errorf("alap.wasm must not be produced: NABC bytecode is not WebAssembly")
	}
	if strings.Contains(string(indexFile), "alap.wasm") || !strings.Contains(string(indexFile), "main.nabc") {
		t.Errorf("index.html must reference main.nabc, not alap.wasm, got: %s", string(indexFile))
	}
}
