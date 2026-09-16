// nilc — NilLang Compiler CLI
// Usage:
//   nilc -in program.nil -out program.nabc   # compile to bytecode
//   nilc -run program.nil                     # compile & run immediately
//   nilc -check program.nil                   # type-check only
//   nilc -dump program.nil                    # print AST + disassembly
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
	"github.com/joysriramsarkar/alap-framework/compiler/parser"
	"github.com/joysriramsarkar/alap-framework/compiler/types"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
)

var (
	inFile    = flag.String("in", "", "Input .nil source file")
	outFile   = flag.String("out", "", "Output .nabc bytecode file")
	runFlag   = flag.Bool("run", false, "Compile and run immediately")
	checkFlag = flag.Bool("check", false, "Type-check only")
	dumpFlag  = flag.Bool("dump", false, "Print AST and bytecode disassembly")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `nilc — NilLang Compiler (Alap Framework)

Usage:
  nilc -in <file.nil> [options]

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  nilc -in hello.nil -run
  nilc -in hello.nil -out hello.nabc
  nilc -in hello.nil -check
  nilc -in hello.nil -dump
`)
	}
	flag.Parse()

	// Allow positional: nilc program.nil
	if *inFile == "" && flag.NArg() > 0 {
		*inFile = flag.Arg(0)
		if *outFile == "" && flag.NArg() > 1 {
			*outFile = flag.Arg(1)
		}
	}

	if *inFile == "" {
		fmt.Fprintln(os.Stderr, "nilc: no input file specified (-in or positional)")
		flag.Usage()
		os.Exit(1)
	}

	src, err := os.ReadFile(*inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nilc: cannot read %q: %v\n", *inFile, err)
		os.Exit(1)
	}

	result := compile(*inFile, string(src))
	if result == nil {
		os.Exit(1)
	}

	if *checkFlag {
		fmt.Printf("✓ %s type-checks OK\n", *inFile)
		return
	}

	if *dumpFlag {
		for _, fn := range result.Functions {
			fmt.Print(fn.Disassemble())
		}
		return
	}

	if *runFlag || *outFile == "" {
		runner := vm.New(result)
		if err := runner.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "nilc: runtime error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	out := *outFile
	if out == "" {
		base := strings.TrimSuffix(filepath.Base(*inFile), filepath.Ext(*inFile))
		out = base + ".nabc"
	}
	bytes := codegen.Serialize(result)
	if err := os.WriteFile(out, bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "nilc: cannot write %q: %v\n", out, err)
		os.Exit(1)
	}
	fmt.Printf("✓ compiled %s → %s (%d bytes)\n", *inFile, out, len(bytes))
}

// compile runs the full pipeline: lex → parse → typecheck → codegen.
func compile(filename, src string) *codegen.Module {
	// 1. Lex
	l := lexer.New(filename, src)
	tokens := l.Tokenize()
	if len(l.Errors()) > 0 {
		for _, e := range l.Errors() {
			fmt.Fprintln(os.Stderr, "lex:", e)
		}
		return nil
	}

	// 2. Parse
	p := parser.New(filename, tokens)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintln(os.Stderr, "parse:", e)
		}
		return nil
	}

	// 3. Type check
	checker := types.New()
	checker.CheckProgram(prog)
	if len(checker.Errors()) > 0 {
		for _, e := range checker.Errors() {
			fmt.Fprintln(os.Stderr, "type:", e)
		}
		return nil
	}

	// 4. Code generation
	gen := codegen.New(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	gen.GenerateProgram(prog)
	if len(gen.Errors()) > 0 {
		for _, e := range gen.Errors() {
			fmt.Fprintln(os.Stderr, "codegen:", e)
		}
		return nil
	}

	return gen.Module()
}
