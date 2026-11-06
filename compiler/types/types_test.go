package types

import (
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
)

func checkSource(src string) *Checker {
	l := lexer.New("test.nil", src)
	p := parser.New("test.nil", l.Tokenize())
	prog := p.Parse()
	c := New()
	c.CheckProgram(prog)
	return c
}

func TestCheckProgramValidHasNoErrors(t *testing.T) {
	c := checkSource(`
let count: i32 = 3
let label: string = "ok"
let flag: bool = true
`)
	if errs := c.Errors(); len(errs) > 0 {
		t.Fatalf("expected no type errors, got %v", errs)
	}
}

func TestCheckProgramDetectsAssignmentMismatch(t *testing.T) {
	c := checkSource(`let count: i32 = "not a number"`)
	if errs := c.Errors(); len(errs) == 0 {
		t.Fatalf("expected a type error assigning a string to i32")
	}
}

func TestCheckProgramCanonicalAliasesAndMatchWildcard(t *testing.T) {
	c := checkSource(`
fn pick(op: string) -> Int {
    match op {
        "+" => 100,
        _ => 200
    }
}
let wide: Int = 1
let nothing: Void = print("x")
let names: Array<string> = ["a", "b"]
let empty: Array<Int> = []
`)
	if errs := c.Errors(); len(errs) > 0 {
		t.Fatalf("expected canonical Int/Void and _ wildcard to type-check, got %v", errs)
	}
}

func TestCheckProgramAllowsNumericWidening(t *testing.T) {
	c := checkSource(`
let small: i32 = 1
let wide: i64 = small
`)
	if errs := c.Errors(); len(errs) > 0 {
		t.Fatalf("expected numeric widening to type-check, got %v", errs)
	}
}
