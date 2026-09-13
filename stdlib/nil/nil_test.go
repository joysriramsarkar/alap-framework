package nil_test

import (
	"os"
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
	nilpkg "github.com/joysriramsarkar/alap-framework/stdlib/nil"
)

func TestNilDBFunctions(t *testing.T) {
	tmp := "/tmp/alap_stdlib_nil_test.db"
	os.Remove(tmp)
	defer os.Remove(tmp)

	mod := &codegen.Module{
		Name: "test_mod",
	}
	runner := vm.New(mod)
	nilpkg.RegisterDBNativeFunctions(runner)

	openFn, ok := runner.GetGlobal("db_open")
	if !ok || openFn.NativeVal == nil {
		t.Fatalf("expected db_open to be registered")
	}

	hVal, err := openFn.NativeVal([]vm.Value{vm.StrVal(tmp)})
	if err != nil {
		t.Fatalf("db_open failed: %v", err)
	}

	insertFn, _ := runner.GetGlobal("db_insert")
	_, err = insertFn.NativeVal([]vm.Value{hVal, vm.StrVal("items"), vm.StrVal(`{"name":"Notebook","price":40}`)})
	if err != nil {
		t.Fatalf("db_insert failed: %v", err)
	}

	countFn, _ := runner.GetGlobal("db_count")
	countVal, err := countFn.NativeVal([]vm.Value{hVal, vm.StrVal("items"), vm.StrVal(`{}`)})
	if err != nil {
		t.Fatalf("db_count failed: %v", err)
	}
	if countVal.IntVal != 1 {
		t.Errorf("expected count 1, got %d", countVal.IntVal)
	}
}
