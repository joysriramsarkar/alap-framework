package parser

import (
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
)

func parseSource(src string) *Parser {
	l := lexer.New("test.nil", src)
	p := New("test.nil", l.Tokenize())
	p.Parse()
	return p
}

func TestParseValidProgramHasNoErrors(t *testing.T) {
	src := `
function add(a: i32, b: i32): i32 {
    return a + b
}
let total: i32 = add(2, 3)
print(total)
`
	if errs := parseSource(src).Errors(); len(errs) > 0 {
		t.Fatalf("expected no parse errors, got %v", errs)
	}
}

// TestParseCanonicalFnSpelling pins M4 convergence: the canonical `fn`
// keyword and `->` return annotation parse to the same AST as `function`/`:`.
func TestParseCanonicalFnSpelling(t *testing.T) {
	src := `
fn add(a: i32, b: i32) -> i32 {
    return a + b
}
let total: i32 = add(2, 3)
print(total)
`
	p := parseSource(src)
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("expected no parse errors for canonical fn spelling, got %v", errs)
	}
}

func TestParseReportsUnexpectedToken(t *testing.T) {
	if errs := parseSource("}").Errors(); len(errs) == 0 {
		t.Fatalf("expected a parse error for a stray closing brace")
	}
}

func TestParseReportsUnterminatedBlock(t *testing.T) {
	if errs := parseSource("function f() {").Errors(); len(errs) == 0 {
		t.Fatalf("expected a parse error for an unterminated block")
	}
}

func TestParseBuildsProgramStatements(t *testing.T) {
	l := lexer.New("test.nil", "let a: i32 = 1\nlet b: i32 = 2\n")
	p := New("test.nil", l.Tokenize())
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("statements = %d, want 2", len(prog.Statements))
	}
}

func TestParseStructAllowsCommaSeparators(t *testing.T) {
	src := "struct Todo {\n    id: i32,\n    title: string,\n}\n"
	if errs := parseSource(src).Errors(); len(errs) > 0 {
		t.Fatalf("expected no parse errors for comma-separated struct, got %v", errs)
	}
}

func TestParseNamedStructLiteral(t *testing.T) {
	l := lexer.New("test.nil", "let c = Counter { value: 1 }\n")
	p := New("test.nil", l.Tokenize())
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("statements = %d, want 1", len(prog.Statements))
	}
}

// TestParseControlBraceNotEaten pins the named-literal disambiguation: a `{`
// after an identifier opens a block unless `IDENT :` follows, so `if x { ... }`
// keeps working after M4 added `Name { ... }` literals.
func TestParseControlBraceNotEaten(t *testing.T) {
	src := "let x = 1\nlet y = 0\nif x {\n    y = 2\n}\nwhile y < 10 {\n    y = y + 1\n}\n"
	p := parseSource(src)
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("control-flow braces must not parse as literals, got %v", errs)
	}
}

// TestParseMatchExpressionArms pins M4 convergence: canonical expression arms
// (`"+" => a + b`) parse alongside classic block arms.
func TestParseMatchExpressionArms(t *testing.T) {
	src := "fn pick(op: string) -> i32 {\n    match op {\n        \"+\" => 100,\n        _ => 200\n    }\n}\n"
	p := parseSource(src)
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("expected no parse errors for expression match arms, got %v", errs)
	}
}
