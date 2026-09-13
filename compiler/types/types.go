// Package types implements the NilLang static type checker.
package types

import (
	"fmt"

	"github.com/joysriramsarkar/alap-framework/compiler/ast"
	"github.com/joysriramsarkar/alap-framework/compiler/lexer"
)

// ─── Built-in types ───────────────────────────────────────────────────────────

type Kind int

const (
	KindVoid Kind = iota
	KindBool
	KindI8; KindI16; KindI32; KindI64
	KindU8; KindU16; KindU32; KindU64
	KindF32; KindF64
	KindBigInt
	KindChar
	KindString
	KindBytes
	KindNull
	KindUndefined
	KindAny
	KindArray
	KindMap
	KindSet
	KindTuple
	KindUnion
	KindFunc
	KindStruct
	KindClass
	KindInterface
	KindEnum
	KindTypeAlias
	KindChannel
	KindFuture
	KindResult
	KindNever
	KindUnknown
)

// Type is a resolved NilLang type.
type Type struct {
	Kind       Kind
	Name       string
	Elem       *Type   // array element / channel payload / future result
	KeyType    *Type   // map key
	ValueType  *Type   // map value
	Fields     map[string]*Type
	Methods    map[string]*FuncType
	TypeParams []*Type
	Union      []*Type
	Tuple      []*Type
	Func       *FuncType
	Nullable   bool
	Sendable   bool
}

type FuncType struct {
	Params  []*Type
	Return  *Type
	Async   bool
}

// Builtin types (singletons)
var (
	Void      = &Type{Kind: KindVoid, Name: "void"}
	Bool      = &Type{Kind: KindBool, Name: "bool"}
	I8        = &Type{Kind: KindI8, Name: "i8"}
	I16       = &Type{Kind: KindI16, Name: "i16"}
	I32       = &Type{Kind: KindI32, Name: "i32"}
	I64       = &Type{Kind: KindI64, Name: "i64"}
	U8        = &Type{Kind: KindU8, Name: "u8"}
	U16       = &Type{Kind: KindU16, Name: "u16"}
	U32       = &Type{Kind: KindU32, Name: "u32"}
	U64       = &Type{Kind: KindU64, Name: "u64"}
	F32       = &Type{Kind: KindF32, Name: "f32"}
	F64       = &Type{Kind: KindF64, Name: "f64"}
	BigInt    = &Type{Kind: KindBigInt, Name: "bigint"}
	Char      = &Type{Kind: KindChar, Name: "char"}
	String    = &Type{Kind: KindString, Name: "string"}
	Bytes     = &Type{Kind: KindBytes, Name: "bytes"}
	Null      = &Type{Kind: KindNull, Name: "null"}
	Undefined = &Type{Kind: KindUndefined, Name: "undefined"}
	Any       = &Type{Kind: KindAny, Name: "any"}
	Never     = &Type{Kind: KindNever, Name: "never"}
)

func ArrayOf(elem *Type) *Type {
	return &Type{Kind: KindArray, Name: elem.Name + "[]", Elem: elem}
}

func ChannelOf(elem *Type) *Type {
	return &Type{Kind: KindChannel, Name: "Channel<" + elem.Name + ">", Elem: elem}
}

func FutureOf(elem *Type) *Type {
	return &Type{Kind: KindFuture, Name: "Future<" + elem.Name + ">", Elem: elem}
}

func ResultOf(ok, err *Type) *Type {
	return &Type{Kind: KindResult, Name: "Result<" + ok.Name + "," + err.Name + ">", Elem: ok, KeyType: err}
}

func NullableOf(t *Type) *Type {
	cp := *t
	cp.Nullable = true
	cp.Name = t.Name + "?"
	return &cp
}

// ─── Environment (scope chain) ────────────────────────────────────────────────

type Env struct {
	parent  *Env
	symbols map[string]*Symbol
}

type Symbol struct {
	Name  string
	Type  *Type
	Const bool
}

func newEnv(parent *Env) *Env {
	return &Env{parent: parent, symbols: make(map[string]*Symbol)}
}

func (e *Env) define(name string, t *Type, isConst bool) {
	e.symbols[name] = &Symbol{Name: name, Type: t, Const: isConst}
}

func (e *Env) lookup(name string) (*Symbol, bool) {
	if s, ok := e.symbols[name]; ok {
		return s, true
	}
	if e.parent != nil {
		return e.parent.lookup(name)
	}
	return nil, false
}

// ─── Checker ──────────────────────────────────────────────────────────────────

type Checker struct {
	global *Env
	errors []string
	types  map[string]*Type // named type registry
}

func New() *Checker {
	c := &Checker{
		global: newEnv(nil),
		types:  make(map[string]*Type),
	}
	c.registerBuiltins()
	return c
}

func (c *Checker) Errors() []string { return c.errors }

func (c *Checker) errorf(pos lexer.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf("%s:%d:%d: type error: "+format, append([]interface{}{pos.Filename, pos.Line, pos.Column}, args...)...)
	c.errors = append(c.errors, msg)
}

func (c *Checker) registerBuiltins() {
	builtins := map[string]*Type{
		"void": Void, "bool": Bool,
		"i8": I8, "i16": I16, "i32": I32, "i64": I64,
		"u8": U8, "u16": U16, "u32": U32, "u64": U64,
		"f32": F32, "f64": F64, "bigint": BigInt,
		"char": Char, "string": String, "bytes": Bytes,
		"null": Null, "undefined": Undefined, "any": Any, "never": Never,
		"number": I64, // compatibility alias → i64
	}
	for name, t := range builtins {
		c.types[name] = t
	}
	// built-in functions
	printFn := &Type{Kind: KindFunc, Name: "print", Func: &FuncType{
		Params: []*Type{Any}, Return: Void,
	}}
	c.global.define("print", printFn, true)

	logFn := &Type{Kind: KindFunc, Name: "log", Func: &FuncType{
		Params: []*Type{Any}, Return: Void,
	}}
	c.global.define("log", logFn, true)

	// Database built-in functions
	c.global.define("db_open", &Type{Kind: KindFunc, Name: "db_open", Func: &FuncType{Params: []*Type{String}, Return: String}}, true)
	c.global.define("db_close", &Type{Kind: KindFunc, Name: "db_close", Func: &FuncType{Params: []*Type{String}, Return: Bool}}, true)
	c.global.define("db_create_table", &Type{Kind: KindFunc, Name: "db_create_table", Func: &FuncType{Params: []*Type{String, String}, Return: Bool}}, true)
	c.global.define("db_insert", &Type{Kind: KindFunc, Name: "db_insert", Func: &FuncType{Params: []*Type{String, String, String}, Return: String}}, true)
	c.global.define("db_select", &Type{Kind: KindFunc, Name: "db_select", Func: &FuncType{Params: []*Type{String, String, String, I32}, Return: String}}, true)
	c.global.define("db_update", &Type{Kind: KindFunc, Name: "db_update", Func: &FuncType{Params: []*Type{String, String, String, String}, Return: I32}}, true)
	c.global.define("db_delete", &Type{Kind: KindFunc, Name: "db_delete", Func: &FuncType{Params: []*Type{String, String, String}, Return: I32}}, true)
	c.global.define("db_count", &Type{Kind: KindFunc, Name: "db_count", Func: &FuncType{Params: []*Type{String, String, String}, Return: I32}}, true)
	c.global.define("db_get", &Type{Kind: KindFunc, Name: "db_get", Func: &FuncType{Params: []*Type{String, String, String}, Return: String}}, true)
	c.global.define("db_tables", &Type{Kind: KindFunc, Name: "db_tables", Func: &FuncType{Params: []*Type{String}, Return: String}}, true)

	// Time & Math & Crypto & FS builtins
	c.global.define("time_now", &Type{Kind: KindFunc, Name: "time_now", Func: &FuncType{Params: nil, Return: I32}}, true)
	c.global.define("time_sleep", &Type{Kind: KindFunc, Name: "time_sleep", Func: &FuncType{Params: []*Type{I32}, Return: Void}}, true)
	c.global.define("crypto_sha256", &Type{Kind: KindFunc, Name: "crypto_sha256", Func: &FuncType{Params: []*Type{String}, Return: String}}, true)
	c.global.define("crypto_md5", &Type{Kind: KindFunc, Name: "crypto_md5", Func: &FuncType{Params: []*Type{String}, Return: String}}, true)
	c.global.define("crypto_uuid", &Type{Kind: KindFunc, Name: "crypto_uuid", Func: &FuncType{Params: nil, Return: String}}, true)
	c.global.define("math_sqrt", &Type{Kind: KindFunc, Name: "math_sqrt", Func: &FuncType{Params: []*Type{Any}, Return: F64}}, true)
	c.global.define("math_abs", &Type{Kind: KindFunc, Name: "math_abs", Func: &FuncType{Params: []*Type{Any}, Return: F64}}, true)
	c.global.define("math_min", &Type{Kind: KindFunc, Name: "math_min", Func: &FuncType{Params: []*Type{Any, Any}, Return: F64}}, true)
	c.global.define("math_max", &Type{Kind: KindFunc, Name: "math_max", Func: &FuncType{Params: []*Type{Any, Any}, Return: F64}}, true)
	c.global.define("fs_read", &Type{Kind: KindFunc, Name: "fs_read", Func: &FuncType{Params: []*Type{String}, Return: String}}, true)
	c.global.define("fs_write", &Type{Kind: KindFunc, Name: "fs_write", Func: &FuncType{Params: []*Type{String, String}, Return: Bool}}, true)
	c.global.define("fs_exists", &Type{Kind: KindFunc, Name: "fs_exists", Func: &FuncType{Params: []*Type{String}, Return: Bool}}, true)

	// Map, Set, Channel
	c.types["Map"] = &Type{Kind: KindMap, Name: "Map"}
	c.types["Set"] = &Type{Kind: KindSet, Name: "Set"}
	c.types["Channel"] = &Type{Kind: KindChannel, Name: "Channel"}
	c.types["Future"] = &Type{Kind: KindFuture, Name: "Future"}
	c.types["Result"] = &Type{Kind: KindResult, Name: "Result"}
}

// CheckProgram type-checks a full program.
func (c *Checker) CheckProgram(prog *ast.Program) {
	// first pass: register top-level declarations
	for _, stmt := range prog.Statements {
		c.registerDecl(stmt, c.global)
	}
	// second pass: check bodies
	for _, stmt := range prog.Statements {
		c.checkStmt(stmt, c.global)
	}
}

func (c *Checker) registerDecl(stmt ast.Statement, env *Env) {
	switch s := stmt.(type) {
	case *ast.FunctionDecl:
		t := c.buildFuncType(s.Params, s.ReturnType, s.Async)
		env.define(s.Name, t, true)
	case *ast.ClassDecl:
		t := &Type{Kind: KindClass, Name: s.Name, Fields: make(map[string]*Type), Methods: make(map[string]*FuncType)}
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.StructDecl:
		t := c.buildStructType(s)
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.InterfaceDecl:
		t := &Type{Kind: KindInterface, Name: s.Name, Fields: make(map[string]*Type), Methods: make(map[string]*FuncType)}
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.EnumDecl:
		t := &Type{Kind: KindEnum, Name: s.Name}
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.TypeDecl:
		t := c.resolveTypeAnnotation(s.Type)
		t.Name = s.Name
		c.types[s.Name] = t
	case *ast.ComponentDecl:
		t := &Type{Kind: KindClass, Name: s.Name, Fields: make(map[string]*Type), Methods: make(map[string]*FuncType)}
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.ActorDecl:
		t := &Type{Kind: KindClass, Name: s.Name, Fields: make(map[string]*Type), Methods: make(map[string]*FuncType)}
		c.types[s.Name] = t
		env.define(s.Name, t, true)
	case *ast.ExportDecl:
		if s.Declaration != nil {
			c.registerDecl(s.Declaration, env)
		}
	}
}

func (c *Checker) buildFuncType(params []*ast.Param, ret *ast.TypeAnnotation, async bool) *Type {
	var ptypes []*Type
	for _, p := range params {
		if p.Type != nil {
			ptypes = append(ptypes, c.resolveTypeAnnotation(p.Type))
		} else {
			ptypes = append(ptypes, Any)
		}
	}
	var retT *Type
	if ret != nil {
		retT = c.resolveTypeAnnotation(ret)
	} else {
		retT = Void
	}
	if async {
		retT = FutureOf(retT)
	}
	return &Type{Kind: KindFunc, Func: &FuncType{Params: ptypes, Return: retT, Async: async}}
}

func (c *Checker) buildStructType(s *ast.StructDecl) *Type {
	t := &Type{Kind: KindStruct, Name: s.Name, Fields: make(map[string]*Type)}
	for _, f := range s.Fields {
		ft := c.resolveTypeAnnotation(f.Type)
		if f.Optional {
			ft = NullableOf(ft)
		}
		t.Fields[f.Name] = ft
	}
	return t
}

func (c *Checker) resolveTypeAnnotation(ann *ast.TypeAnnotation) *Type {
	if ann == nil {
		return Any
	}
	if ann.Name == "union" && len(ann.Union) > 0 {
		var parts []*Type
		for _, u := range ann.Union {
			parts = append(parts, c.resolveTypeAnnotation(u))
		}
		return &Type{Kind: KindUnion, Name: "union", Union: parts}
	}
	if ann.Array {
		elem := c.resolveTypeAnnotation(&ast.TypeAnnotation{Name: ann.Name, Generic: ann.Generic})
		return ArrayOf(elem)
	}
	t, found := c.types[ann.Name]
	if !found {
		t = &Type{Kind: KindUnknown, Name: ann.Name}
	}
	if ann.Nullable {
		t = NullableOf(t)
	}
	return t
}

// ─── Statement checking ───────────────────────────────────────────────────────

func (c *Checker) checkStmt(stmt ast.Statement, env *Env) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		c.checkVarDecl(s, env)
	case *ast.FunctionDecl:
		c.checkFunctionDecl(s, env)
	case *ast.ExprStatement:
		c.checkExpr(s.Expr, env)
	case *ast.ReturnStatement:
		if s.Value != nil {
			c.checkExpr(s.Value, env)
		}
	case *ast.IfStatement:
		c.checkExpr(s.Condition, env)
		c.checkBlock(s.Consequent, env)
		if s.Alternative != nil {
			c.checkStmt(s.Alternative, env)
		}
	case *ast.WhileStatement:
		c.checkExpr(s.Condition, env)
		c.checkBlock(s.Body, env)
	case *ast.ForStatement:
		inner := newEnv(env)
		if s.Init != nil {
			c.checkStmt(s.Init, inner)
		}
		if s.Condition != nil {
			c.checkExpr(s.Condition, inner)
		}
		if s.Update != nil {
			c.checkExpr(s.Update, inner)
		}
		c.checkBlock(s.Body, inner)
	case *ast.ForInStatement:
		inner := newEnv(env)
		iterT := c.checkExpr(s.Iterable, env)
		var elemT *Type = Any
		if iterT.Kind == KindArray {
			elemT = iterT.Elem
		}
		inner.define(s.VarName, elemT, false)
		c.checkBlock(s.Body, inner)
	case *ast.LoopStatement:
		c.checkBlock(s.Body, env)
	case *ast.BlockStatement:
		c.checkBlock(s, env)
	case *ast.ClassDecl:
		c.checkClassDecl(s, env)
	case *ast.StructDecl:
		// already registered
	case *ast.TryCatchStatement:
		c.checkBlock(s.Try, env)
		if s.Catch != nil {
			inner := newEnv(env)
			if s.Catch.Param != nil {
				inner.define(s.Catch.Param.Name, Any, false)
			}
			c.checkBlock(s.Catch.Body, inner)
		}
		if s.Finally != nil {
			c.checkBlock(s.Finally, env)
		}
	case *ast.MatchStatement:
		c.checkExpr(s.Subject, env)
		for _, arm := range s.Arms {
			inner := newEnv(env)
			if arm.Pattern != nil {
				c.checkExpr(arm.Pattern, inner)
			}
			c.checkBlock(arm.Body, inner)
		}
	case *ast.TaskStatement:
		c.checkBlock(s.Body, env)
	case *ast.SpawnStatement:
		c.checkExpr(s.Call, env)
	case *ast.ThrowStatement:
		c.checkExpr(s.Value, env)
	case *ast.ComponentDecl:
		c.checkComponentDecl(s, env)
	case *ast.ExportDecl:
		if s.Declaration != nil {
			c.checkStmt(s.Declaration, env)
		}
	case *ast.ImportDecl:
		// imports are resolved at link time
	}
}

func (c *Checker) checkBlock(block *ast.BlockStatement, parent *Env) {
	if block == nil {
		return
	}
	inner := newEnv(parent)
	for _, stmt := range block.Body {
		c.checkStmt(stmt, inner)
	}
}

func (c *Checker) checkVarDecl(s *ast.VarDecl, env *Env) {
	var declaredType *Type
	if s.Type != nil {
		declaredType = c.resolveTypeAnnotation(s.Type)
	}
	var initType *Type
	if s.Init != nil {
		initType = c.checkExpr(s.Init, env)
	}
	finalType := declaredType
	if finalType == nil {
		finalType = initType
	}
	if finalType == nil {
		finalType = Any
	}
	if declaredType != nil && initType != nil {
		if !c.isAssignable(declaredType, initType) {
			c.errorf(s.Pos, "cannot assign %s to %s", initType.Name, declaredType.Name)
		}
	}
	env.define(s.Name, finalType, s.Kind == "const")
}

func (c *Checker) checkFunctionDecl(s *ast.FunctionDecl, env *Env) {
	inner := newEnv(env)
	for _, p := range s.Params {
		var pt *Type
		if p.Type != nil {
			pt = c.resolveTypeAnnotation(p.Type)
		} else {
			pt = Any
		}
		inner.define(p.Name, pt, false)
	}
	c.checkBlock(s.Body, inner)
}

func (c *Checker) checkClassDecl(s *ast.ClassDecl, env *Env) {
	inner := newEnv(env)
	classType, _ := c.types[s.Name]
	if classType == nil {
		classType = &Type{Kind: KindClass, Name: s.Name, Fields: make(map[string]*Type), Methods: make(map[string]*FuncType)}
	}
	inner.define("this", classType, true)
	for _, m := range s.Members {
		switch mem := m.(type) {
		case *ast.MethodMember:
			minner := newEnv(inner)
			for _, p := range mem.Params {
				var pt *Type
				if p.Type != nil {
					pt = c.resolveTypeAnnotation(p.Type)
				} else {
					pt = Any
				}
				minner.define(p.Name, pt, false)
			}
			c.checkBlock(mem.Body, minner)
		case *ast.FieldMember:
			var ft *Type
			if mem.Type != nil {
				ft = c.resolveTypeAnnotation(mem.Type)
			} else {
				ft = Any
			}
			if classType.Fields == nil {
				classType.Fields = make(map[string]*Type)
			}
			classType.Fields[mem.Name] = ft
		}
	}
}

func (c *Checker) checkComponentDecl(s *ast.ComponentDecl, env *Env) {
	inner := newEnv(env)
	for _, f := range s.StateFields {
		var ft *Type
		if f.Type != nil {
			ft = c.resolveTypeAnnotation(f.Type)
		} else {
			ft = Any
		}
		inner.define(f.Name, ft, false)
	}
	for _, lc := range s.Lifecycle {
		minner := newEnv(inner)
		c.checkBlock(lc.Body, minner)
	}
}

// ─── Expression checking (returns type) ──────────────────────────────────────

func (c *Checker) checkExpr(expr ast.Expression, env *Env) *Type {
	if expr == nil {
		return Void
	}
	switch e := expr.(type) {
	case *ast.IntLiteral:
		return I32
	case *ast.FloatLiteral:
		return F64
	case *ast.StringLiteral:
		return String
	case *ast.BoolLiteral:
		return Bool
	case *ast.NullLiteral:
		return Null
	case *ast.ArrayLiteral:
		if len(e.Elements) == 0 {
			return ArrayOf(Any)
		}
		elemT := c.checkExpr(e.Elements[0], env)
		return ArrayOf(elemT)
	case *ast.ObjectLiteral:
		t := &Type{Kind: KindStruct, Name: "object", Fields: make(map[string]*Type)}
		for _, f := range e.Fields {
			t.Fields[f.Name] = c.checkExpr(f.Value, env)
		}
		return t
	case *ast.Identifier:
		sym, ok := env.lookup(e.Name)
		if !ok {
			c.errorf(e.Pos, "undefined: %s", e.Name)
			return Any
		}
		return sym.Type
	case *ast.BinaryExpr:
		left := c.checkExpr(e.Left, env)
		right := c.checkExpr(e.Right, env)
		return c.checkBinary(e.Op, left, right, e.Pos)
	case *ast.UnaryExpr:
		t := c.checkExpr(e.Operand, env)
		if e.Op == "!" {
			return Bool
		}
		return t
	case *ast.AssignExpr:
		left := c.checkExpr(e.Left, env)
		right := c.checkExpr(e.Right, env)
		if !c.isAssignable(left, right) {
			c.errorf(e.Pos, "cannot assign %s to %s", right.Name, left.Name)
		}
		return left
	case *ast.CallExpr:
		calleeT := c.checkExpr(e.Callee, env)
		if calleeT.Kind == KindFunc && calleeT.Func != nil {
			return calleeT.Func.Return
		}
		return Any
	case *ast.MemberExpr:
		objT := c.checkExpr(e.Object, env)
		if objT.Fields != nil {
			if ft, ok := objT.Fields[e.Property]; ok {
				return ft
			}
		}
		return Any
	case *ast.IndexExpr:
		arrT := c.checkExpr(e.Object, env)
		if arrT.Kind == KindArray && arrT.Elem != nil {
			return arrT.Elem
		}
		return Any
	case *ast.NewExpr:
		name := ""
		if id, ok := e.Constructor.(*ast.Identifier); ok {
			name = id.Name
		}
		if t, ok := c.types[name]; ok {
			return t
		}
		return Any
	case *ast.AwaitExpr:
		t := c.checkExpr(e.Operand, env)
		if t.Kind == KindFuture && t.Elem != nil {
			return t.Elem
		}
		return t
	case *ast.TryExpr:
		t := c.checkExpr(e.Operand, env)
		if t.Kind == KindResult && t.Elem != nil {
			return t.Elem
		}
		return t
	case *ast.ChanReceiveExpr:
		t := c.checkExpr(e.Channel, env)
		if t.Kind == KindChannel && t.Elem != nil {
			return t.Elem
		}
		return Any
	case *ast.TypeAssertExpr:
		return c.resolveTypeAnnotation(e.Type)
	case *ast.ArrowFuncExpr:
		return c.buildFuncType(e.Params, e.ReturnType, false)
	case *ast.TernaryExpr:
		c.checkExpr(e.Condition, env)
		t1 := c.checkExpr(e.Consequent, env)
		c.checkExpr(e.Alternative, env)
		return t1
	case *ast.ChanSendStatement:
		c.checkExpr(e.Channel, env)
		c.checkExpr(e.Value, env)
		return Void
	}
	return Any
}

func (c *Checker) checkBinary(op string, left, right *Type, pos lexer.Position) *Type {
	numeric := func(t *Type) bool {
		return t.Kind >= KindI8 && t.Kind <= KindF64
	}
	switch op {
	case "+":
		if left.Kind == KindString || right.Kind == KindString {
			return String
		}
		if numeric(left) && numeric(right) {
			return left
		}
		return Any
	case "-", "*", "/", "%":
		if numeric(left) && numeric(right) {
			return left
		}
		c.errorf(pos, "operator %s requires numeric operands, got %s and %s", op, left.Name, right.Name)
		return Any
	case "==", "!=", "<", ">", "<=", ">=":
		return Bool
	case "&&", "||":
		return Bool
	case "??":
		return left
	}
	return Any
}

func (c *Checker) isAssignable(target, source *Type) bool {
	if target.Kind == KindAny || source.Kind == KindAny {
		return true
	}
	if target.Nullable && source.Kind == KindNull {
		return true
	}
	if target.Kind == source.Kind {
		return true
	}
	// numeric widening
	numericKinds := map[Kind]int{KindI8: 1, KindI16: 2, KindI32: 3, KindI64: 4, KindF32: 5, KindF64: 6}
	tw, tok := numericKinds[target.Kind]
	sw, sok := numericKinds[source.Kind]
	if tok && sok {
		return sw <= tw
	}
	return false
}
