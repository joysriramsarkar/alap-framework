package pos_test

import (
	"testing"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
)

var lastOutput string

func runPOS(t *testing.T, src string) *vm.VM {
	t.Helper()
	l := lexer.New("pos_run.nil", src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		t.Fatalf("lex: %v", l.Errors())
	}
	p := parser.New("pos_run.nil", tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse: %v", p.Errors())
	}
	checker := types.New()
	checker.CheckProgram(prog)
	gen := codegen.New("pos_run")
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		t.Fatalf("codegen: %v", gen.Errors())
	}
	runner := vm.New(gen.Module())
	if err := runner.Run(); err != nil {
		t.Fatalf("runtime: %v", err)
	}
	lastOutput = runner.Output()
	return runner
}

func readOutput() string { return lastOutput }
