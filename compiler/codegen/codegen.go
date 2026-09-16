// Package codegen generates NABC (NilLang Application Bytecode) from the AST.
// NABC is a simple stack-based bytecode that the nilrt VM executes.
package codegen

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/joysriramsarkar/alap-framework/compiler/ast"
)

// ─── Opcode definitions ──────────────────────────────────────────────────────

type Opcode byte

const (
	OP_NOP Opcode = iota

	// Stack / locals
	OP_LOAD_CONST  // load from constant pool
	OP_LOAD_LOCAL  // load local variable
	OP_STORE_LOCAL // store local variable
	OP_LOAD_GLOBAL
	OP_STORE_GLOBAL
	OP_LOAD_FIELD
	OP_STORE_FIELD
	OP_LOAD_INDEX
	OP_STORE_INDEX
	OP_POP
	OP_DUP

	// Literals
	OP_LOAD_NULL
	OP_LOAD_TRUE
	OP_LOAD_FALSE
	OP_LOAD_INT    // inline int32
	OP_LOAD_FLOAT  // inline float64

	// Arithmetic
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_NEG
	OP_POW

	// Logic
	OP_NOT
	OP_AND
	OP_OR

	// Comparison
	OP_EQ
	OP_NEQ
	OP_LT
	OP_GT
	OP_LTE
	OP_GTE

	// String
	OP_STR_CONCAT

	// Control flow
	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_JUMP_IF_TRUE
	OP_JUMP_IF_NOT_NULL

	// Functions
	OP_CALL        // call function: arg count in operand
	OP_CALL_METHOD // call method: name idx, arg count
	OP_RETURN
	OP_RETURN_VOID

	// Objects / arrays / maps
	OP_NEW_OBJECT  // field count in operand
	OP_NEW_ARRAY   // element count
	OP_NEW_MAP
	OP_NEW_CLASS

	// Async / concurrency
	OP_AWAIT
	OP_SPAWN_TASK
	OP_CHAN_SEND
	OP_CHAN_RECV

	// UI instructions
	OP_UI_CREATE
	OP_UI_PROP
	OP_UI_EVENT
	OP_UI_CHILD
	OP_UI_END

	// Error handling
	OP_THROW
	OP_TRY_BEGIN
	OP_TRY_END
	OP_CATCH_BEGIN

	// Misc
	OP_PRINT
	OP_HALT
)

var opcodeNames = map[Opcode]string{
	OP_NOP: "NOP", OP_LOAD_CONST: "LOAD_CONST",
	OP_LOAD_LOCAL: "LOAD_LOCAL", OP_STORE_LOCAL: "STORE_LOCAL",
	OP_LOAD_GLOBAL: "LOAD_GLOBAL", OP_STORE_GLOBAL: "STORE_GLOBAL",
	OP_LOAD_FIELD: "LOAD_FIELD", OP_STORE_FIELD: "STORE_FIELD",
	OP_LOAD_INDEX: "LOAD_INDEX", OP_STORE_INDEX: "STORE_INDEX",
	OP_POP: "POP", OP_DUP: "DUP",
	OP_LOAD_NULL: "LOAD_NULL", OP_LOAD_TRUE: "LOAD_TRUE", OP_LOAD_FALSE: "LOAD_FALSE",
	OP_LOAD_INT: "LOAD_INT", OP_LOAD_FLOAT: "LOAD_FLOAT",
	OP_ADD: "ADD", OP_SUB: "SUB", OP_MUL: "MUL", OP_DIV: "DIV", OP_MOD: "MOD", OP_NEG: "NEG",
	OP_NOT: "NOT", OP_AND: "AND", OP_OR: "OR",
	OP_EQ: "EQ", OP_NEQ: "NEQ", OP_LT: "LT", OP_GT: "GT", OP_LTE: "LTE", OP_GTE: "GTE",
	OP_STR_CONCAT: "STR_CONCAT",
	OP_JUMP: "JUMP", OP_JUMP_IF_FALSE: "JUMP_IF_FALSE", OP_JUMP_IF_TRUE: "JUMP_IF_TRUE",
	OP_JUMP_IF_NOT_NULL: "JUMP_IF_NOT_NULL",
	OP_CALL: "CALL", OP_CALL_METHOD: "CALL_METHOD", OP_RETURN: "RETURN", OP_RETURN_VOID: "RETURN_VOID",
	OP_NEW_OBJECT: "NEW_OBJECT", OP_NEW_ARRAY: "NEW_ARRAY", OP_NEW_MAP: "NEW_MAP",
	OP_AWAIT: "AWAIT", OP_SPAWN_TASK: "SPAWN_TASK", OP_CHAN_SEND: "CHAN_SEND", OP_CHAN_RECV: "CHAN_RECV",
	OP_UI_CREATE: "UI_CREATE", OP_UI_PROP: "UI_PROP", OP_UI_EVENT: "UI_EVENT",
	OP_UI_CHILD: "UI_CHILD", OP_UI_END: "UI_END",
	OP_THROW: "THROW", OP_TRY_BEGIN: "TRY_BEGIN", OP_TRY_END: "TRY_END", OP_CATCH_BEGIN: "CATCH_BEGIN",
	OP_PRINT: "PRINT", OP_HALT: "HALT",
}

func (op Opcode) String() string {
	if s, ok := opcodeNames[op]; ok {
		return s
	}
	return fmt.Sprintf("OP_%d", op)
}

// ─── Instruction ─────────────────────────────────────────────────────────────

type Instruction struct {
	Op      Opcode
	Operand int32 // inline integer (index / count / jump target)
}

func (i Instruction) String() string {
	return fmt.Sprintf("%-20s %d", i.Op, i.Operand)
}

// ─── Constant pool ────────────────────────────────────────────────────────────

type ConstKind int

const (
	ConstInt ConstKind = iota
	ConstFloat
	ConstString
	ConstBool
	ConstNull
)

type Constant struct {
	Kind    ConstKind
	IntVal  int64
	FloatVal float64
	StrVal  string
	BoolVal bool
}

func (c Constant) String() string {
	switch c.Kind {
	case ConstInt:
		return fmt.Sprintf("Int(%d)", c.IntVal)
	case ConstFloat:
		return fmt.Sprintf("Float(%g)", c.FloatVal)
	case ConstString:
		return fmt.Sprintf("String(%q)", c.StrVal)
	case ConstBool:
		return fmt.Sprintf("Bool(%v)", c.BoolVal)
	case ConstNull:
		return "Null"
	}
	return "?"
}

// ─── Function object ─────────────────────────────────────────────────────────

type Function struct {
	Name       string
	ParamCount int
	LocalCount int
	Code       []Instruction
	Constants  []Constant
	LocalNames []string
	Async      bool
}

func (f *Function) Disassemble() string {
	s := fmt.Sprintf("=== function %s (params=%d locals=%d async=%v) ===\n", f.Name, f.ParamCount, f.LocalCount, f.Async)
	for i, ins := range f.Code {
		s += fmt.Sprintf("  %4d  %s\n", i, ins)
	}
	return s
}

// ─── Module (compilation unit) ───────────────────────────────────────────────

type Module struct {
	Name      string
	Functions []*Function
	Globals   []string
	MainFunc  *Function
}

// ─── Generator ────────────────────────────────────────────────────────────────

type Generator struct {
	module  *Module
	current *Function
	locals  map[string]int
	errors  []string
	loopBreaks  [][]int
	loopContinues [][]int
}

func New(moduleName string) *Generator {
	return &Generator{
		module: &Module{Name: moduleName},
	}
}

func (g *Generator) Errors() []string { return g.errors }
func (g *Generator) Module() *Module  { return g.module }

func (g *Generator) errorf(format string, args ...interface{}) {
	g.errors = append(g.errors, fmt.Sprintf("codegen: "+format, args...))
}

// ─── Emit helpers ─────────────────────────────────────────────────────────────

func (g *Generator) emit(op Opcode, operand ...int32) int {
	var oper int32
	if len(operand) > 0 {
		oper = operand[0]
	}
	ins := Instruction{Op: op, Operand: oper}
	g.current.Code = append(g.current.Code, ins)
	return len(g.current.Code) - 1
}

func (g *Generator) patch(at int, target int32) {
	g.current.Code[at].Operand = target
}

func (g *Generator) currentPos() int {
	return len(g.current.Code)
}

func (g *Generator) addConst(c Constant) int32 {
	for i, existing := range g.current.Constants {
		if existing.Kind == c.Kind && existing.StrVal == c.StrVal && existing.IntVal == c.IntVal {
			return int32(i)
		}
	}
	g.current.Constants = append(g.current.Constants, c)
	return int32(len(g.current.Constants) - 1)
}

func (g *Generator) defineLocal(name string) int32 {
	idx := len(g.current.LocalNames)
	g.current.LocalNames = append(g.current.LocalNames, name)
	g.locals[name] = idx
	if idx+1 > g.current.LocalCount {
		g.current.LocalCount = idx + 1
	}
	return int32(idx)
}

func (g *Generator) lookupLocal(name string) (int, bool) {
	idx, ok := g.locals[name]
	return idx, ok
}

// ─── Code generation entry ────────────────────────────────────────────────────

// GenerateProgram compiles a whole program into the module.
func (g *Generator) GenerateProgram(prog *ast.Program) {
	main := &Function{Name: "__main__"}
	g.module.MainFunc = main
	g.module.Functions = append(g.module.Functions, main)
	g.enterFunction(main)

	for _, stmt := range prog.Statements {
		g.genStmt(stmt)
	}
	g.emit(OP_HALT)
	g.leaveFunction()
}

func (g *Generator) enterFunction(fn *Function) {
	g.current = fn
	g.locals = make(map[string]int)
}

func (g *Generator) leaveFunction() {
	g.current = g.module.MainFunc
	g.locals = make(map[string]int)
}

// ─── Statement codegen ────────────────────────────────────────────────────────

func (g *Generator) genStmt(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.genVarDecl(s)
	case *ast.FunctionDecl:
		g.genFunctionDecl(s)
	case *ast.ExprStatement:
		g.genExpr(s.Expr)
		g.emit(OP_POP)
	case *ast.ReturnStatement:
		if s.Value != nil {
			g.genExpr(s.Value)
			g.emit(OP_RETURN)
		} else {
			g.emit(OP_RETURN_VOID)
		}
	case *ast.IfStatement:
		g.genIf(s)
	case *ast.WhileStatement:
		g.genWhile(s)
	case *ast.ForStatement:
		g.genFor(s)
	case *ast.ForInStatement:
		g.genForIn(s)
	case *ast.LoopStatement:
		g.genLoop(s)
	case *ast.BreakStatement:
		if len(g.loopBreaks) > 0 {
			jmp := g.emit(OP_JUMP, 0)
			g.loopBreaks[len(g.loopBreaks)-1] = append(g.loopBreaks[len(g.loopBreaks)-1], jmp)
		}
	case *ast.ContinueStatement:
		if len(g.loopContinues) > 0 {
			jmp := g.emit(OP_JUMP, 0)
			g.loopContinues[len(g.loopContinues)-1] = append(g.loopContinues[len(g.loopContinues)-1], jmp)
		}
	case *ast.BlockStatement:
		for _, st := range s.Body {
			g.genStmt(st)
		}
	case *ast.ThrowStatement:
		g.genExpr(s.Value)
		g.emit(OP_THROW)
	case *ast.TryCatchStatement:
		g.genTryCatch(s)
	case *ast.SpawnStatement:
		g.genExpr(s.Call)
		g.emit(OP_SPAWN_TASK)
	case *ast.TaskStatement:
		// inline task block as spawn
		fn := g.genAnonymousFunc(nil, s.Body)
		idx := g.addFuncToModule(fn)
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: fn.Name})
		g.emit(OP_LOAD_CONST, cidx)
		_ = idx
		g.emit(OP_SPAWN_TASK)
	case *ast.ChanSendStatement:
		g.genExpr(s.Channel)
		g.genExpr(s.Value)
		g.emit(OP_CHAN_SEND)
	case *ast.ComponentDecl:
		g.genComponentDecl(s)
	case *ast.ClassDecl:
		g.genClassDecl(s)
	case *ast.MatchStatement:
		g.genMatch(s)
	case *ast.ImportDecl, *ast.ExportDecl, *ast.InterfaceDecl,
		*ast.TypeDecl, *ast.EnumDecl, *ast.StructDecl,
		*ast.ActorDecl, *ast.StoreDecl:
		// metadata only; no code emitted at runtime
	}
}

func (g *Generator) genVarDecl(s *ast.VarDecl) {
	if s.Init != nil {
		g.genExpr(s.Init)
	} else {
		g.emit(OP_LOAD_NULL)
	}
	idx := g.defineLocal(s.Name)
	g.emit(OP_STORE_LOCAL, idx)
}

func (g *Generator) genFunctionDecl(s *ast.FunctionDecl) {
	fn := &Function{Name: s.Name, Async: s.Async}
	savedFunc := g.current
	savedLocals := g.locals
	g.current = fn
	g.locals = make(map[string]int)

	for _, p := range s.Params {
		g.defineLocal(p.Name)
	}
	fn.ParamCount = len(s.Params)
	g.genBlock(s.Body)
	g.emit(OP_RETURN_VOID)

	g.current = savedFunc
	g.locals = savedLocals
	g.module.Functions = append(g.module.Functions, fn)
}

func (g *Generator) genAnonymousFunc(params []*ast.Param, body *ast.BlockStatement) *Function {
	fn := &Function{Name: fmt.Sprintf("__anon_%d__", len(g.module.Functions))}
	savedFunc := g.current
	savedLocals := g.locals
	g.current = fn
	g.locals = make(map[string]int)
	for _, p := range params {
		g.defineLocal(p.Name)
	}
	fn.ParamCount = len(params)
	g.genBlock(body)
	g.emit(OP_RETURN_VOID)
	g.current = savedFunc
	g.locals = savedLocals
	return fn
}

func (g *Generator) addFuncToModule(fn *Function) int {
	g.module.Functions = append(g.module.Functions, fn)
	return len(g.module.Functions) - 1
}

func (g *Generator) genBlock(block *ast.BlockStatement) {
	if block == nil {
		return
	}
	for _, stmt := range block.Body {
		g.genStmt(stmt)
	}
}

func (g *Generator) genIf(s *ast.IfStatement) {
	g.genExpr(s.Condition)
	jmpFalse := g.emit(OP_JUMP_IF_FALSE, 0)
	g.genBlock(s.Consequent)
	if s.Alternative != nil {
		jmpEnd := g.emit(OP_JUMP, 0)
		g.patch(jmpFalse, int32(g.currentPos()))
		g.genStmt(s.Alternative)
		g.patch(jmpEnd, int32(g.currentPos()))
	} else {
		g.patch(jmpFalse, int32(g.currentPos()))
	}
}

func (g *Generator) genWhile(s *ast.WhileStatement) {
	g.loopBreaks = append(g.loopBreaks, nil)
	g.loopContinues = append(g.loopContinues, nil)

	loopStart := g.currentPos()
	g.genExpr(s.Condition)
	jmpEnd := g.emit(OP_JUMP_IF_FALSE, 0)
	g.genBlock(s.Body)
	// patch continues to loop start
	for _, c := range g.loopContinues[len(g.loopContinues)-1] {
		g.patch(c, int32(loopStart))
	}
	g.emit(OP_JUMP, int32(loopStart))
	end := g.currentPos()
	g.patch(jmpEnd, int32(end))
	// patch breaks
	for _, b := range g.loopBreaks[len(g.loopBreaks)-1] {
		g.patch(b, int32(end))
	}
	g.loopBreaks = g.loopBreaks[:len(g.loopBreaks)-1]
	g.loopContinues = g.loopContinues[:len(g.loopContinues)-1]
}

func (g *Generator) genFor(s *ast.ForStatement) {
	if s.Init != nil {
		g.genStmt(s.Init)
	}
	g.loopBreaks = append(g.loopBreaks, nil)
	g.loopContinues = append(g.loopContinues, nil)

	loopStart := g.currentPos()
	var jmpEnd int = -1
	if s.Condition != nil {
		g.genExpr(s.Condition)
		jmpEnd = g.emit(OP_JUMP_IF_FALSE, 0)
	}
	g.genBlock(s.Body)
	continueTarget := g.currentPos()
	for _, c := range g.loopContinues[len(g.loopContinues)-1] {
		g.patch(c, int32(continueTarget))
	}
	if s.Update != nil {
		g.genExpr(s.Update)
		g.emit(OP_POP)
	}
	g.emit(OP_JUMP, int32(loopStart))
	end := g.currentPos()
	if jmpEnd >= 0 {
		g.patch(jmpEnd, int32(end))
	}
	for _, b := range g.loopBreaks[len(g.loopBreaks)-1] {
		g.patch(b, int32(end))
	}
	g.loopBreaks = g.loopBreaks[:len(g.loopBreaks)-1]
	g.loopContinues = g.loopContinues[:len(g.loopContinues)-1]
}

func (g *Generator) genForIn(s *ast.ForInStatement) {
	g.genExpr(s.Iterable)
	// simplified: iterate via index
	// Push array, then index=0
	idxLocal := g.defineLocal("__forin_idx__")
	arrLocal := g.defineLocal("__forin_arr__")
	g.emit(OP_STORE_LOCAL, arrLocal)
	g.emit(OP_LOAD_INT, 0)
	g.emit(OP_STORE_LOCAL, idxLocal)

	g.loopBreaks = append(g.loopBreaks, nil)
	g.loopContinues = append(g.loopContinues, nil)
	loopStart := g.currentPos()

	// condition: idx < arr.length
	g.emit(OP_LOAD_LOCAL, idxLocal)
	g.emit(OP_LOAD_LOCAL, arrLocal)
	lenIdx := g.addConst(Constant{Kind: ConstString, StrVal: "length"})
	g.emit(OP_LOAD_CONST, lenIdx)
	g.emit(OP_LOAD_FIELD)
	g.emit(OP_LT)
	jmpEnd := g.emit(OP_JUMP_IF_FALSE, 0)

	// body: let varName = arr[idx]
	g.emit(OP_LOAD_LOCAL, arrLocal)
	g.emit(OP_LOAD_LOCAL, idxLocal)
	g.emit(OP_LOAD_INDEX)
	varIdx := g.defineLocal(s.VarName)
	g.emit(OP_STORE_LOCAL, varIdx)
	g.genBlock(s.Body)

	// increment
	continueTarget := g.currentPos()
	for _, c := range g.loopContinues[len(g.loopContinues)-1] {
		g.patch(c, int32(continueTarget))
	}
	g.emit(OP_LOAD_LOCAL, idxLocal)
	g.emit(OP_LOAD_INT, 1)
	g.emit(OP_ADD)
	g.emit(OP_STORE_LOCAL, idxLocal)
	g.emit(OP_JUMP, int32(loopStart))

	end := g.currentPos()
	g.patch(jmpEnd, int32(end))
	for _, b := range g.loopBreaks[len(g.loopBreaks)-1] {
		g.patch(b, int32(end))
	}
	g.loopBreaks = g.loopBreaks[:len(g.loopBreaks)-1]
	g.loopContinues = g.loopContinues[:len(g.loopContinues)-1]
}

func (g *Generator) genLoop(s *ast.LoopStatement) {
	g.loopBreaks = append(g.loopBreaks, nil)
	g.loopContinues = append(g.loopContinues, nil)

	loopStart := g.currentPos()
	g.genBlock(s.Body)
	continueTarget := g.currentPos()
	for _, c := range g.loopContinues[len(g.loopContinues)-1] {
		g.patch(c, int32(continueTarget))
	}
	g.emit(OP_JUMP, int32(loopStart))
	end := g.currentPos()
	for _, b := range g.loopBreaks[len(g.loopBreaks)-1] {
		g.patch(b, int32(end))
	}
	g.loopBreaks = g.loopBreaks[:len(g.loopBreaks)-1]
	g.loopContinues = g.loopContinues[:len(g.loopContinues)-1]
}

func (g *Generator) genTryCatch(s *ast.TryCatchStatement) {
	catchTarget := g.emit(OP_TRY_BEGIN, 0)
	g.genBlock(s.Try)
	g.emit(OP_TRY_END)
	jmpSkipCatch := g.emit(OP_JUMP, 0)
	g.patch(catchTarget, int32(g.currentPos()))
	if s.Catch != nil {
		g.emit(OP_CATCH_BEGIN)
		if s.Catch.Param != nil {
			idx := g.defineLocal(s.Catch.Param.Name)
			g.emit(OP_STORE_LOCAL, idx)
		} else {
			g.emit(OP_POP)
		}
		g.genBlock(s.Catch.Body)
	}
	g.patch(jmpSkipCatch, int32(g.currentPos()))
	if s.Finally != nil {
		g.genBlock(s.Finally)
	}
}

func (g *Generator) genMatch(s *ast.MatchStatement) {
	g.genExpr(s.Subject)
	subjIdx := g.defineLocal("__match_subj__")
	g.emit(OP_STORE_LOCAL, subjIdx)
	var jmpEnds []int
	for _, arm := range s.Arms {
		g.emit(OP_LOAD_LOCAL, subjIdx)
		g.genExpr(arm.Pattern)
		g.emit(OP_EQ)
		jmpNext := g.emit(OP_JUMP_IF_FALSE, 0)
		g.genBlock(arm.Body)
		jmpEnd := g.emit(OP_JUMP, 0)
		jmpEnds = append(jmpEnds, jmpEnd)
		g.patch(jmpNext, int32(g.currentPos()))
	}
	end := g.currentPos()
	for _, j := range jmpEnds {
		g.patch(j, int32(end))
	}
}

func (g *Generator) genClassDecl(s *ast.ClassDecl) {
	// Emit class skeleton — full OOP is handled by runtime
	cidx := g.addConst(Constant{Kind: ConstString, StrVal: s.Name})
	g.emit(OP_LOAD_CONST, cidx)
	g.emit(OP_NEW_CLASS)
	idx := g.defineLocal(s.Name)
	g.emit(OP_STORE_LOCAL, idx)
}

func (g *Generator) genComponentDecl(s *ast.ComponentDecl) {
	// Emit UI component declaration
	cidx := g.addConst(Constant{Kind: ConstString, StrVal: s.Name})
	g.emit(OP_LOAD_CONST, cidx)
	g.emit(OP_UI_CREATE)
	if s.BuildBody != nil {
		g.genUINode(s.BuildBody)
	}
	g.emit(OP_UI_END)
	idx := g.defineLocal(s.Name)
	g.emit(OP_STORE_LOCAL, idx)
}

func (g *Generator) genUINode(node *ast.UINode) {
	cidx := g.addConst(Constant{Kind: ConstString, StrVal: node.Widget})
	g.emit(OP_LOAD_CONST, cidx)
	g.emit(OP_UI_CREATE)
	for _, arg := range node.Args {
		g.genExpr(arg)
		midx := g.addConst(Constant{Kind: ConstString, StrVal: "text"})
		g.emit(OP_LOAD_CONST, midx)
		g.emit(OP_UI_PROP)
	}
	for _, mod := range node.Modifiers {
		if len(mod.Args) == 0 {
			g.emit(OP_LOAD_TRUE)
		} else if len(mod.Args) == 1 {
			g.genExpr(mod.Args[0])
		} else {
			// Pack multi-arguments into an array
			for _, a := range mod.Args {
				g.genExpr(a)
			}
			g.emit(OP_NEW_ARRAY, int32(len(mod.Args)))
		}
		midx := g.addConst(Constant{Kind: ConstString, StrVal: mod.Name})
		g.emit(OP_LOAD_CONST, midx)
		g.emit(OP_UI_PROP)
	}
	for _, ev := range node.EventHandlers {
		eidx := g.addConst(Constant{Kind: ConstString, StrVal: ev.Name})
		g.emit(OP_LOAD_CONST, eidx)
		g.genExpr(ev.Handler)
		g.emit(OP_UI_EVENT)
	}
	for _, child := range node.Children {
		g.genUINode(child)
		g.emit(OP_UI_CHILD)
	}
	g.emit(OP_UI_END)
}

// ─── Expression codegen ───────────────────────────────────────────────────────

func (g *Generator) genExpr(expr ast.Expression) {
	if expr == nil {
		g.emit(OP_LOAD_NULL)
		return
	}
	switch e := expr.(type) {
	case *ast.IntLiteral:
		n, _ := strconv.ParseInt(e.Value, 10, 64)
		if n >= -32768 && n <= 32767 {
			g.emit(OP_LOAD_INT, int32(n))
		} else {
			cidx := g.addConst(Constant{Kind: ConstInt, IntVal: n})
			g.emit(OP_LOAD_CONST, cidx)
		}
	case *ast.FloatLiteral:
		f, _ := strconv.ParseFloat(e.Value, 64)
		bits := math.Float64bits(f)
		cidx := g.addConst(Constant{Kind: ConstFloat, FloatVal: f, IntVal: int64(bits)})
		g.emit(OP_LOAD_CONST, cidx)
	case *ast.StringLiteral:
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: e.Value})
		g.emit(OP_LOAD_CONST, cidx)
	case *ast.BoolLiteral:
		if e.Value {
			g.emit(OP_LOAD_TRUE)
		} else {
			g.emit(OP_LOAD_FALSE)
		}
	case *ast.NullLiteral:
		g.emit(OP_LOAD_NULL)
	case *ast.Identifier:
		if e.Name == "print" {
			return // handled inline
		}
		if idx, ok := g.lookupLocal(e.Name); ok {
			g.emit(OP_LOAD_LOCAL, int32(idx))
		} else {
			cidx := g.addConst(Constant{Kind: ConstString, StrVal: e.Name})
			g.emit(OP_LOAD_CONST, cidx)
			g.emit(OP_LOAD_GLOBAL)
		}
	case *ast.BinaryExpr:
		g.genBinary(e)
	case *ast.UnaryExpr:
		g.genExpr(e.Operand)
		switch e.Op {
		case "!":
			g.emit(OP_NOT)
		case "-":
			g.emit(OP_NEG)
		}
	case *ast.AssignExpr:
		g.genAssign(e)
	case *ast.CallExpr:
		g.genCall(e)
	case *ast.MemberExpr:
		g.genExpr(e.Object)
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: e.Property})
		g.emit(OP_LOAD_CONST, cidx)
		g.emit(OP_LOAD_FIELD)
	case *ast.IndexExpr:
		g.genExpr(e.Object)
		g.genExpr(e.Index)
		g.emit(OP_LOAD_INDEX)
	case *ast.NewExpr:
		for _, arg := range e.Arguments {
			g.genExpr(arg)
		}
		if id, ok := e.Constructor.(*ast.Identifier); ok {
			cidx := g.addConst(Constant{Kind: ConstString, StrVal: id.Name})
			g.emit(OP_LOAD_CONST, cidx)
		}
		g.emit(OP_NEW_OBJECT, int32(len(e.Arguments)))
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			g.genExpr(el)
		}
		g.emit(OP_NEW_ARRAY, int32(len(e.Elements)))
	case *ast.ObjectLiteral:
		for _, f := range e.Fields {
			cidx := g.addConst(Constant{Kind: ConstString, StrVal: f.Name})
			g.emit(OP_LOAD_CONST, cidx)
			g.genExpr(f.Value)
		}
		g.emit(OP_NEW_OBJECT, int32(len(e.Fields)))
	case *ast.AwaitExpr:
		g.genExpr(e.Operand)
		g.emit(OP_AWAIT)
	case *ast.TryExpr:
		g.genExpr(e.Operand)
		// try-expression: if Result.Err, throw
	case *ast.ChanReceiveExpr:
		g.genExpr(e.Channel)
		g.emit(OP_CHAN_RECV)
	case *ast.ChanSendStatement:
		g.genExpr(e.Channel)
		g.genExpr(e.Value)
		g.emit(OP_CHAN_SEND)
	case *ast.TypeAssertExpr:
		g.genExpr(e.Value) // just emit value; runtime checks type
	case *ast.TernaryExpr:
		g.genExpr(e.Condition)
		jmpFalse := g.emit(OP_JUMP_IF_FALSE, 0)
		g.genExpr(e.Consequent)
		jmpEnd := g.emit(OP_JUMP, 0)
		g.patch(jmpFalse, int32(g.currentPos()))
		g.genExpr(e.Alternative)
		g.patch(jmpEnd, int32(g.currentPos()))
	case *ast.ArrowFuncExpr:
		var params []*ast.Param
		params = e.Params
		var body *ast.BlockStatement
		switch b := e.Body.(type) {
		case *ast.BlockStatement:
			body = b
		default:
			// wrap expression in return
			body = &ast.BlockStatement{Body: []ast.Statement{
				&ast.ReturnStatement{Value: b.(ast.Expression)},
			}}
		}
		fn := g.genAnonymousFunc(params, body)
		g.addFuncToModule(fn)
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: fn.Name})
		g.emit(OP_LOAD_CONST, cidx)
	}
}

func (g *Generator) genBinary(e *ast.BinaryExpr) {
	g.genExpr(e.Left)
	if e.Op == "??" {
		// Keep the left value on the stack when it is non-null. When it is
		// null, discard it before evaluating the fallback expression.
		g.emit(OP_DUP)
		jmpEnd := g.emit(OP_JUMP_IF_NOT_NULL, 0)
		g.emit(OP_POP)
		g.genExpr(e.Right)
		g.patch(jmpEnd, int32(g.currentPos()))
		return
	}
	g.genExpr(e.Right)
	switch e.Op {
	case "+":
		g.emit(OP_ADD)
	case "-":
		g.emit(OP_SUB)
	case "*":
		g.emit(OP_MUL)
	case "/":
		g.emit(OP_DIV)
	case "%":
		g.emit(OP_MOD)
	case "**":
		g.emit(OP_POW)
	case "==":
		g.emit(OP_EQ)
	case "!=":
		g.emit(OP_NEQ)
	case "<":
		g.emit(OP_LT)
	case ">":
		g.emit(OP_GT)
	case "<=":
		g.emit(OP_LTE)
	case ">=":
		g.emit(OP_GTE)
	case "&&":
		g.emit(OP_AND)
	case "||":
		g.emit(OP_OR)
	}
}

func (g *Generator) genAssign(e *ast.AssignExpr) {
	if e.Op != "=" {
		// compound: load, operate, store
		g.genExpr(e.Left)
		g.genExpr(e.Right)
		switch e.Op {
		case "+=":
			g.emit(OP_ADD)
		case "-=":
			g.emit(OP_SUB)
		case "*=":
			g.emit(OP_MUL)
		case "/=":
			g.emit(OP_DIV)
		case "%=":
			g.emit(OP_MOD)
		}
	} else {
		g.genExpr(e.Right)
	}
	// store
	switch left := e.Left.(type) {
	case *ast.Identifier:
		if idx, ok := g.lookupLocal(left.Name); ok {
			g.emit(OP_DUP)
			g.emit(OP_STORE_LOCAL, int32(idx))
		} else {
			g.emit(OP_DUP)
			cidx := g.addConst(Constant{Kind: ConstString, StrVal: left.Name})
			g.emit(OP_LOAD_CONST, cidx)
			g.emit(OP_STORE_GLOBAL)
		}
	case *ast.MemberExpr:
		// already computed value is on stack; need object+field
		g.genExpr(left.Object)
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: left.Property})
		g.emit(OP_LOAD_CONST, cidx)
		g.emit(OP_STORE_FIELD)
	}
}

func (g *Generator) genCall(e *ast.CallExpr) {
	// special: print(…)
	if id, ok := e.Callee.(*ast.Identifier); ok && id.Name == "print" {
		for _, arg := range e.Arguments {
			g.genExpr(arg)
		}
		g.emit(OP_PRINT, int32(len(e.Arguments)))
		return
	}
	// method call
	if mem, ok := e.Callee.(*ast.MemberExpr); ok {
		g.genExpr(mem.Object)
		for _, arg := range e.Arguments {
			g.genExpr(arg)
		}
		cidx := g.addConst(Constant{Kind: ConstString, StrVal: mem.Property})
		g.emit(OP_LOAD_CONST, cidx) // method name
		g.emit(OP_CALL_METHOD, int32(len(e.Arguments)))
		return
	}
	// regular call
	g.genExpr(e.Callee)
	for _, arg := range e.Arguments {
		g.genExpr(arg)
	}
	g.emit(OP_CALL, int32(len(e.Arguments)))
}

// ─── NABC binary serialisation ────────────────────────────────────────────────

// Serialize converts a Module to NABC bytes.
func Serialize(mod *Module) []byte {
	var buf []byte
	writeU32 := func(n uint32) {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, n)
		buf = append(buf, b...)
	}
	writeStr := func(s string) {
		writeU32(uint32(len(s)))
		buf = append(buf, []byte(s)...)
	}
	// magic
	buf = append(buf, []byte("NABC")...)
	writeU32(1) // version
	writeStr(mod.Name)
	writeU32(uint32(len(mod.Functions)))
	for _, fn := range mod.Functions {
		writeStr(fn.Name)
		writeU32(uint32(fn.ParamCount))
		writeU32(uint32(fn.LocalCount))
		// constants
		writeU32(uint32(len(fn.Constants)))
		for _, c := range fn.Constants {
			buf = append(buf, byte(c.Kind))
			switch c.Kind {
			case ConstInt:
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(c.IntVal))
				buf = append(buf, b...)
			case ConstFloat:
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, math.Float64bits(c.FloatVal))
				buf = append(buf, b...)
			case ConstString:
				writeStr(c.StrVal)
			case ConstBool:
				if c.BoolVal {
					buf = append(buf, 1)
				} else {
					buf = append(buf, 0)
				}
			case ConstNull:
				// nothing extra
			}
		}
		// instructions
		writeU32(uint32(len(fn.Code)))
		for _, ins := range fn.Code {
			buf = append(buf, byte(ins.Op))
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(ins.Operand))
			buf = append(buf, b...)
		}
	}
	return buf
}

// Deserialize decodes NABC binary data into a Module.
func Deserialize(data []byte) (*Module, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid NABC bytecode: too short")
	}
	if string(data[:4]) != "NABC" {
		return nil, fmt.Errorf("invalid NABC bytecode header: %q", string(data[:4]))
	}

	offset := 4
	readU32 := func() (uint32, error) {
		if offset+4 > len(data) {
			return 0, fmt.Errorf("unexpected EOF reading uint32")
		}
		v := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
		return v, nil
	}

	readStr := func() (string, error) {
		l, err := readU32()
		if err != nil {
			return "", err
		}
		if offset+int(l) > len(data) {
			return "", fmt.Errorf("unexpected EOF reading string")
		}
		s := string(data[offset : offset+int(l)])
		offset += int(l)
		return s, nil
	}

	_, err := readU32() // version
	if err != nil {
		return nil, err
	}

	modName, err := readStr()
	if err != nil {
		return nil, err
	}

	fnCount, err := readU32()
	if err != nil {
		return nil, err
	}

	mod := &Module{
		Name:      modName,
		Functions: make([]*Function, 0, fnCount),
	}

	for i := 0; i < int(fnCount); i++ {
		fnName, err := readStr()
		if err != nil {
			return nil, err
		}
		paramCount, err := readU32()
		if err != nil {
			return nil, err
		}
		localCount, err := readU32()
		if err != nil {
			return nil, err
		}
		constCount, err := readU32()
		if err != nil {
			return nil, err
		}

		fn := &Function{
			Name:       fnName,
			ParamCount: int(paramCount),
			LocalCount: int(localCount),
			Constants:  make([]Constant, 0, constCount),
		}

		for c := 0; c < int(constCount); c++ {
			if offset >= len(data) {
				return nil, fmt.Errorf("unexpected EOF reading constant tag")
			}
			kind := ConstKind(data[offset])
			offset++

			var constant Constant
			constant.Kind = kind

			switch kind {
			case ConstInt:
				if offset+8 > len(data) {
					return nil, fmt.Errorf("unexpected EOF reading int const")
				}
				constant.IntVal = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
				offset += 8
			case ConstFloat:
				if offset+8 > len(data) {
					return nil, fmt.Errorf("unexpected EOF reading float const")
				}
				bits := binary.LittleEndian.Uint64(data[offset : offset+8])
				constant.FloatVal = math.Float64frombits(bits)
				offset += 8
			case ConstString:
				s, err := readStr()
				if err != nil {
					return nil, err
				}
				constant.StrVal = s
			case ConstBool:
				if offset >= len(data) {
					return nil, fmt.Errorf("unexpected EOF reading bool const")
				}
				constant.BoolVal = data[offset] == 1
				offset++
			case ConstNull:
				// no extra bytes
			}
			fn.Constants = append(fn.Constants, constant)
		}

		insCount, err := readU32()
		if err != nil {
			return nil, err
		}
		fn.Code = make([]Instruction, 0, insCount)

		for j := 0; j < int(insCount); j++ {
			if offset+5 > len(data) {
				return nil, fmt.Errorf("unexpected EOF reading instruction")
			}
			op := Opcode(data[offset])
			oper := int32(binary.LittleEndian.Uint32(data[offset+1 : offset+5]))
			offset += 5
			fn.Code = append(fn.Code, Instruction{Op: op, Operand: oper})
		}

		mod.Functions = append(mod.Functions, fn)
		if fn.Name == "" || fn.Name == "main" || mod.MainFunc == nil {
			mod.MainFunc = fn
		}
	}

	return mod, nil
}
