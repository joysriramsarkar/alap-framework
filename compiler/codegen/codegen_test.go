package codegen

import (
	"strings"
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
)

func buildModule(t *testing.T, src string) *Module {
	t.Helper()

	l := lexer.New("test.nil", src)
	tokens := l.Tokenize()
	if errs := l.Errors(); len(errs) > 0 {
		t.Fatalf("lex errors: %v", errs)
	}

	p := parser.New("test.nil", tokens)
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	g := New("demo")
	g.GenerateProgram(prog)
	if errs := g.Errors(); len(errs) > 0 {
		t.Fatalf("codegen errors: %v", errs)
	}
	return g.Module()
}

func TestSerializeDeserializeRoundtrip(t *testing.T) {
	mod := buildModule(t, `
function add(a: i32, b: i32): i32 {
    return a + b
}
print(add(2, 3))
`)
	data := Serialize(mod)
	if len(data) == 0 {
		t.Fatalf("Serialize returned an empty byte slice")
	}

	got, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize error = %v", err)
	}
	if got.Name != mod.Name {
		t.Fatalf("module name = %q, want %q", got.Name, mod.Name)
	}
	if len(got.Functions) != len(mod.Functions) {
		t.Fatalf("function count = %d, want %d", len(got.Functions), len(mod.Functions))
	}
}

func TestDeserializeRejectsInvalidInput(t *testing.T) {
	if _, err := Deserialize([]byte("not-bytecode")); err == nil {
		t.Fatalf("expected Deserialize to reject invalid input")
	}
	if _, err := Deserialize([]byte("NAB")); err == nil {
		t.Fatalf("expected Deserialize to reject a too-short input")
	}
}

func TestOpcodeString(t *testing.T) {
	if got := OP_ADD.String(); got != "ADD" {
		t.Fatalf("OP_ADD.String() = %q, want ADD", got)
	}
	if got := Opcode(250).String(); got != "OP_250" {
		t.Fatalf("unknown opcode String() = %q, want OP_250", got)
	}
}

func TestFunctionDisassemble(t *testing.T) {
	fn := &Function{Name: "demo", Code: []Instruction{{Op: OP_NOP}}}
	out := fn.Disassemble()
	if !strings.Contains(out, "demo") {
		t.Fatalf("Disassemble() = %q, want it to mention the function name", out)
	}
}
