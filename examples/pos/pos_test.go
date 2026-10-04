package pos_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/platform/android"
	"github.com/joysriramsarkar/alap-framework/platform/onuron"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	"github.com/joysriramsarkar/alap-framework/stdlib/database"
)

func TestPOSVMWithDatabase(t *testing.T) {
	tmp := "/tmp/alap_pos_vm.db"
	os.Remove(tmp)
	defer os.Remove(tmp)

	src := `
function main(): void {
    let db: string = db_open("` + tmp + `")
    db_create_table(db, "test_table")
    db_insert(db, "test_table", "{\"id\":1,\"name\":\"hello\"}")
    let count: i32 = db_count(db, "test_table", "{}")
    print("Count:" + count.toString())
    let row: string = db_get(db, "test_table", "{\"id\":\"1\"}")
    print("Row:" + row)
    db_close(db)
}
`
	runPOS(t, src)
	out := readOutput()
	if !strings.Contains(out, "Count:1") {
		t.Errorf("expected Count:1, got: %s", out)
	}
}

func TestPOSFullPipeline(t *testing.T) {
	tmp := "/tmp/alap_pos_pipeline.db"
	os.Remove(tmp)
	defer os.Remove(tmp)

	src := `
function json_product(id, name, price, qty): string {
    return "{\"id\":" + id.toString() + ",\"name\":\"" + name + "\",\"price\":" + price.toString() + ",\"qty\":" + qty.toString() + "}"
}
function main(): void {
    let db: string = db_open("` + tmp + `")
    db_create_table(db, "products")
    db_create_table(db, "cart")
    db_insert(db, "products", json_product("1", "Apple", "150", "100"))
    db_insert(db, "products", json_product("2", "Bread", "45", "50"))
    let count: i32 = db_count(db, "products", "{}")
    print("ProductCount:" + count.toString())
    let row: string = db_get(db, "products", "{\"id\":\"1\"}")
    print("Product1:" + row)
    let tables: string = db_tables(db)
    print("Tables:" + tables)
    db_close(db)
}
`
	runPOS(t, src)
	out := readOutput()
	if !strings.Contains(out, "ProductCount:2") {
		t.Errorf("ProductCount:2: got %s", out)
	}
	if !strings.Contains(out, "Product1:") {
		t.Errorf("Product1:: got %s", out)
	}
	if !strings.Contains(out, "Tables:") {
		t.Errorf("Tables:: got %s", out)
	}

	db, err := database.New(tmp)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db.Close()
	if db.Count("products", map[string]interface{}{}) != 2 {
		t.Error("expected 2 products persisted")
	}
}

func TestPOSNilLangCompile(t *testing.T) {
	src := `
function calculateTotal(items: i32, price: i32): i32 {
    return items * price
}
function main(): void {
    let total: i32 = calculateTotal(6, 90)
    print("Total:" + total.toString())
    let discount: i32 = 50
    let final: i32 = total - discount
    print("Final:" + final.toString())
}
`
	runPOS(t, src)
	out := readOutput()
	if !strings.Contains(out, "Total:540") {
		t.Errorf("Total:540: got %s", out)
	}
	if !strings.Contains(out, "Final:490") {
		t.Errorf("Final:490: got %s", out)
	}
}

func TestPOSDeclarativeUIHierarchy(t *testing.T) {
	posSrc, err := os.ReadFile("src/pos_app.nil")
	if err != nil {
		t.Fatalf("failed reading src/pos_app.nil: %v", err)
	}

	l := lexer.New("pos_app.nil", string(posSrc))
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}

	p := parser.New("pos_app.nil", tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	checker := types.New()
	checker.CheckProgram(prog)

	gen := codegen.New("pos_app")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		t.Fatalf("codegen errors: %v", gen.Errors())
	}

	runner := vm.New(gen.Module())
	if err := runner.Run(); err != nil {
		t.Fatalf("runtime error: %v", err)
	}

	tree := runner.GetUITree()
	if tree == nil || tree.Root == nil {
		t.Fatalf("expected UI Tree root to be built from POSApp component")
	}

	runner.ComputeUILayout(1080, 1920)
	if tree.Root.Bounds.Width != 1080 || tree.Root.Bounds.Height != 1920 {
		t.Errorf("expected bounds 1080x1920, got %fx%f", tree.Root.Bounds.Width, tree.Root.Bounds.Height)
	}

	textTree := tree.RenderTextTree()
	if !strings.Contains(textTree, "Column") || !strings.Contains(textTree, "Row") {
		t.Errorf("expected rendered text tree to contain Column and Row containers:\n%s", textTree)
	}
	if !strings.Contains(textTree, "Button") {
		t.Errorf("expected rendered tree to contain Buttons:\n%s", textTree)
	}
}

func TestPOSMultiPlatformPackaging(t *testing.T) {
	posSrc, err := os.ReadFile("src/pos_app.nil")
	if err != nil {
		t.Fatalf("failed reading pos_app.nil: %v", err)
	}

	l := lexer.New("pos_app.nil", string(posSrc))
	tokens := l.Tokenize()
	p := parser.New("pos_app.nil", tokens)
	prog := p.Parse()
	gen := codegen.New("pos_app")
	gen.GenerateProgram(prog)
	bytecode := codegen.Serialize(gen.Module())

	if len(bytecode) == 0 {
		t.Fatalf("expected non-empty bytecode for POS app")
	}

	tmpBuild := filepath.Join(os.TempDir(), "alap_pos_build_test")
	_ = os.RemoveAll(tmpBuild)
	defer os.RemoveAll(tmpBuild)

	// 1. Onuron .nilax package generation
	onuronAdapter := onuron.New()
	if err := onuronAdapter.GenerateProject(tmpBuild, bytecode); err != nil {
		t.Fatalf("Onuron GenerateProject failed: %v", err)
	}

	manifestPath := filepath.Join(tmpBuild, "onuron", "app.alapmanifest")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("expected Onuron manifest at %s", manifestPath)
	}
	nabcPath := filepath.Join(tmpBuild, "onuron", "bin", "main.nabc")
	if _, err := os.Stat(nabcPath); os.IsNotExist(err) {
		t.Errorf("expected Onuron bundled bytecode at %s", nabcPath)
	}

	// 2. Android Gradle/JNI project generation
	androidAdapter := android.New()
	androidDir := filepath.Join(tmpBuild, "android")
	if err := androidAdapter.GenerateProject(androidDir, bytecode); err != nil {
		t.Fatalf("Android GenerateProject failed: %v", err)
	}

	gradlePath := filepath.Join(androidDir, "build.gradle.kts")
	if _, err := os.Stat(gradlePath); os.IsNotExist(err) {
		t.Errorf("expected Android build.gradle.kts at %s", gradlePath)
	}
	androidAsset := filepath.Join(androidDir, "app", "src", "main", "assets", "main.nabc")
	if _, err := os.Stat(androidAsset); os.IsNotExist(err) {
		t.Errorf("expected Android asset main.nabc at %s", androidAsset)
	}
}

func TestPOSConcurrentCheckoutSimulation(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "alap_pos_concurrency.db")
	_ = os.Remove(tmp)
	defer os.Remove(tmp)

	db, err := database.New(tmp)
	if err != nil {
		t.Fatalf("failed creating db: %v", err)
	}
	defer db.Close()

	db.CreateTable("inventory")
	db.CreateTable("orders")

	// 100 items of Apple stock
	db.Insert("inventory", database.Record{"id": float64(1), "name": "Apple", "stock": float64(100)})

	var wg sync.WaitGroup
	cashiers := 20
	var mu sync.Mutex
	successfulCheckouts := 0

	for i := 0; i < cashiers; i++ {
		wg.Add(1)
		go func(cashierID int) {
			defer wg.Done()
			mu.Lock()
			rec, ok := db.GetRecord("inventory", map[string]interface{}{"id": float64(1)})
			if ok {
				currentStock := rec["stock"].(float64)
				if currentStock >= 2 {
					db.Update("inventory", map[string]interface{}{"id": float64(1)}, database.Record{"stock": currentStock - 2})
					db.Insert("orders", database.Record{"cashier": float64(cashierID), "qty": float64(2)})
					successfulCheckouts++
				}
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if successfulCheckouts != cashiers {
		t.Errorf("expected %d successful checkouts, got %d", cashiers, successfulCheckouts)
	}

	apple, _ := db.GetRecord("inventory", map[string]interface{}{"id": float64(1)})
	expectedStock := float64(100 - (cashiers * 2))
	if apple["stock"] != expectedStock {
		t.Errorf("expected stock %v, got %v", expectedStock, apple["stock"])
	}

	orderCount := db.Count("orders", map[string]interface{}{})
	if orderCount != cashiers {
		t.Errorf("expected %d orders logged, got %d", cashiers, orderCount)
	}
}

func TestMiniPOSSourceEndToEnd(t *testing.T) {
	src, err := os.ReadFile("src/main.nil")
	if err != nil {
		t.Fatalf("failed reading src/main.nil: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "alap_minipos_test_*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	runner := runPOS(t, string(src))
	out := runner.Output()

	if !strings.Contains(out, "ONURON FRESH MARKET — MINI POS") {
		t.Errorf("expected header, got:\n%s", out)
	}
	if !strings.Contains(out, "Lookup by Barcode (8901001)") {
		t.Errorf("expected barcode lookup, got:\n%s", out)
	}
	if !strings.Contains(out, "Net Amount Due:     $603") {
		t.Errorf("expected Net Amount Due: $603, got:\n%s", out)
	}
	if !strings.Contains(out, "INV-20261004-001") {
		t.Errorf("expected invoice number, got:\n%s", out)
	}
	if !strings.Contains(out, "Mini POS Execution & Persistence Verified Successfully") {
		t.Errorf("expected verification success, got:\n%s", out)
	}
}

func TestMiniPOSOnuronLifecycleExecution(t *testing.T) {
	src, err := os.ReadFile("src/main.nil")
	if err != nil {
		t.Fatalf("failed reading src/main.nil: %v", err)
	}

	l := lexer.New("mini_pos.nil", string(src))
	tokens := l.Tokenize()
	p := parser.New("mini_pos.nil", tokens)
	prog := p.Parse()
	gen := codegen.New("mini_pos")
	gen.GenerateProgram(prog)
	bytecode := codegen.Serialize(gen.Module())

	tmpDir, err := os.MkdirTemp("", "onuron_lifecycle_test_*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	adapter := onuron.New()
	adapter.AppName = "mini-pos"
	adapter.Version = "1.0.0"

	if err := adapter.GenerateProject(tmpDir, bytecode); err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(tmpDir, "onuron", "app.alapmanifest"))
	if err != nil {
		t.Fatalf("manifest read: %v", err)
	}
	if !strings.Contains(string(manifestBytes), "Name = mini-pos") {
		t.Errorf("expected manifest Name = mini-pos, got:\n%s", string(manifestBytes))
	}
	if !strings.Contains(string(manifestBytes), "Version = 1.0.0") {
		t.Errorf("expected manifest Version = 1.0.0, got:\n%s", string(manifestBytes))
	}

	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	if err := adapter.RunApp(filepath.Join(tmpDir, "onuron")); err != nil {
		t.Fatalf("RunApp failed: %v", err)
	}
}

