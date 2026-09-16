// Package vm implements the nilrt virtual machine that executes NABC bytecode.
package vm

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	nilcrypto "github.com/joysriramsarkar/alap-framework/stdlib/crypto"
	nilfs "github.com/joysriramsarkar/alap-framework/stdlib/fs"
	niljson "github.com/joysriramsarkar/alap-framework/stdlib/json"
	nillog "github.com/joysriramsarkar/alap-framework/stdlib/log"
	nilmath "github.com/joysriramsarkar/alap-framework/stdlib/math"
	nilnet "github.com/joysriramsarkar/alap-framework/stdlib/net"
	niltime "github.com/joysriramsarkar/alap-framework/stdlib/time"
	"github.com/joysriramsarkar/alap-framework/ui/engine"
	"github.com/joysriramsarkar/alap-framework/ui/layout"
)

// ─── Value types ─────────────────────────────────────────────────────────────

type ValueKind int

const (
	ValNil ValueKind = iota
	ValBool
	ValInt
	ValFloat
	ValString
	ValBytes
	ValArray
	ValMap
	ValObject
	ValFunc
	ValNativeFunc
	ValChannel
)

// Value is a NilLang runtime value (tagged union).
type Value struct {
	Kind      ValueKind
	BoolVal   bool
	IntVal    int64
	FloatVal  float64
	StrVal    string
	Slice     []Value
	MapVal    map[string]Value
	FuncVal   *Closure
	NativeVal func(args []Value) (Value, error)
	Chan      *Channel
}

// Closure wraps a compiled function with a captured environment.
type Closure struct {
	Fn      *codegen.Function
	Capture []Value
}

// Channel is a Go-channel-backed NilLang Channel<T>.
type Channel struct {
	C    chan Value
	mu   sync.Mutex
	closed bool
}

func NewChannel(cap int) *Channel {
	return &Channel{C: make(chan Value, cap)}
}

func (ch *Channel) Send(v Value) {
	ch.C <- v
}

func (ch *Channel) Recv() (Value, bool) {
	v, ok := <-ch.C
	return v, ok
}

// ─── Value constructors ───────────────────────────────────────────────────────

var Nil = Value{Kind: ValNil}
var True = Value{Kind: ValBool, BoolVal: true}
var False = Value{Kind: ValBool, BoolVal: false}

func IntVal(n int64) Value    { return Value{Kind: ValInt, IntVal: n} }
func FloatVal(f float64) Value { return Value{Kind: ValFloat, FloatVal: f} }
func StrVal(s string) Value   { return Value{Kind: ValString, StrVal: s} }
func BoolVal(b bool) Value {
	if b {
		return True
	}
	return False
}
func ArrayVal(s []Value) Value { return Value{Kind: ValArray, Slice: s} }
func ObjVal(m map[string]Value) Value { return Value{Kind: ValObject, MapVal: m} }
func FuncVal(fn *codegen.Function) Value {
	return Value{Kind: ValFunc, FuncVal: &Closure{Fn: fn}}
}
func NativeFuncVal(fn func(args []Value) (Value, error)) Value {
	return Value{Kind: ValNativeFunc, NativeVal: fn}
}

func toVMValue(v interface{}) Value {
	if v == nil {
		return Nil
	}
	switch val := v.(type) {
	case bool:
		return BoolVal(val)
	case float64:
		return FloatVal(val)
	case int:
		return IntVal(int64(val))
	case int64:
		return IntVal(val)
	case string:
		return StrVal(val)
	case []interface{}:
		slice := make([]Value, len(val))
		for i, el := range val {
			slice[i] = toVMValue(el)
		}
		return ArrayVal(slice)
	case map[string]interface{}:
		m := make(map[string]Value)
		for k, el := range val {
			m[k] = toVMValue(el)
		}
		return ObjVal(m)
	}
	return StrVal(fmt.Sprintf("%v", v))
}

// ─── Value operations ─────────────────────────────────────────────────────────

func (v Value) String() string {
	switch v.Kind {
	case ValNil:
		return "null"
	case ValBool:
		if v.BoolVal {
			return "true"
		}
		return "false"
	case ValInt:
		return fmt.Sprintf("%d", v.IntVal)
	case ValFloat:
		return fmt.Sprintf("%g", v.FloatVal)
	case ValString:
		return v.StrVal
	case ValArray:
		parts := make([]string, len(v.Slice))
		for i, el := range v.Slice {
			parts[i] = el.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case ValMap, ValObject:
		parts := make([]string, 0, len(v.MapVal))
		for k, val := range v.MapVal {
			parts = append(parts, fmt.Sprintf("%s: %s", k, val.String()))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case ValFunc:
		if v.FuncVal != nil && v.FuncVal.Fn != nil {
			return "<fn:" + v.FuncVal.Fn.Name + ">"
		}
		return "<fn>"
	case ValChannel:
		return "<channel>"
	}
	return "<unknown>"
}

func (v Value) toRawValue() interface{} {
	switch v.Kind {
	case ValBool:
		return v.BoolVal
	case ValInt:
		return float64(v.IntVal)
	case ValFloat:
		return v.FloatVal
	case ValString:
		return v.StrVal
	case ValArray:
		slice := make([]interface{}, len(v.Slice))
		for i, el := range v.Slice {
			slice[i] = el.toRawValue()
		}
		return slice
	case ValMap, ValObject:
		m := make(map[string]interface{})
		for k, el := range v.MapVal {
			m[k] = el.toRawValue()
		}
		return m
	}
	return v.String()
}

func (v Value) Truthy() bool {
	switch v.Kind {
	case ValNil:
		return false
	case ValBool:
		return v.BoolVal
	case ValInt:
		return v.IntVal != 0
	case ValFloat:
		return v.FloatVal != 0
	case ValString:
		return v.StrVal != ""
	case ValArray:
		return len(v.Slice) > 0
	}
	return true
}

func (v Value) Equal(other Value) bool {
	if v.Kind != other.Kind {
		if v.Kind == ValNil || other.Kind == ValNil {
			return v.Kind == other.Kind
		}
		// numeric cross-kind equality
		return v.toFloat() == other.toFloat()
	}
	switch v.Kind {
	case ValNil:
		return true
	case ValBool:
		return v.BoolVal == other.BoolVal
	case ValInt:
		return v.IntVal == other.IntVal
	case ValFloat:
		return v.FloatVal == other.FloatVal
	case ValString:
		return v.StrVal == other.StrVal
	}
	return false
}

func (v Value) toFloat() float64 {
	switch v.Kind {
	case ValInt:
		return float64(v.IntVal)
	case ValFloat:
		return v.FloatVal
	}
	return 0
}

// ─── VM frame ─────────────────────────────────────────────────────────────────

type Frame struct {
	fn      *codegen.Function
	ip      int
	locals  []Value
	base    int // stack base
}

func newFrame(fn *codegen.Function) *Frame {
	locals := make([]Value, fn.LocalCount+fn.ParamCount+4)
	return &Frame{fn: fn, locals: locals}
}

// ─── VM ───────────────────────────────────────────────────────────────────────

// VM is the NilLang bytecode virtual machine.
type VM struct {
	module   *codegen.Module
	globals  map[string]Value
	stack    []Value
	frames   []*Frame
	output   strings.Builder
	errors   []string
	maxStack int
	UITree   *engine.UITree
}

// New creates a VM for the given compiled module.
func New(mod *codegen.Module) *VM {
	vm := &VM{
		module:   mod,
		globals:  make(map[string]Value),
		maxStack: 8192,
		UITree:   engine.NewUITree(),
	}
	vm.registerBuiltins()
	return vm
}

// GetUITree returns the active UI tree.
func (vm *VM) GetUITree() *engine.UITree { return vm.UITree }

// ComputeUILayout calculates bounds for all UI elements.
func (vm *VM) ComputeUILayout(width, height float64) {
	if vm.UITree != nil && vm.UITree.Root != nil {
		layout.ComputeLayout(vm.UITree.Root, layout.LayoutContext{
			ViewportWidth:  width,
			ViewportHeight: height,
			Scale:          1.0,
		})
	}
}

// Output returns everything printed during execution.
func (vm *VM) Output() string { return vm.output.String() }

// Errors returns runtime errors.
func (vm *VM) Errors() []string { return vm.errors }

func (vm *VM) registerBuiltins() {
	// Built-in constants
	vm.globals["__nil__"] = Nil
	vm.globals["__true__"] = True
	vm.globals["__false__"] = False

	// Time
	vm.globals["time_now"] = NativeFuncVal(func(args []Value) (Value, error) {
		return IntVal(niltime.NowMs()), nil
	})
	vm.globals["time_sleep"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) > 0 {
			niltime.Sleep(args[0].IntVal)
		}
		return Nil, nil
	})

	// Math
	vm.globals["math_abs"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Abs(args[0].toFloat())), nil
	})
	vm.globals["math_sqrt"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Sqrt(args[0].toFloat())), nil
	})
	vm.globals["math_floor"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Floor(args[0].toFloat())), nil
	})
	vm.globals["math_ceil"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Ceil(args[0].toFloat())), nil
	})
	vm.globals["math_round"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Round(args[0].toFloat())), nil
	})
	vm.globals["math_random"] = NativeFuncVal(func(args []Value) (Value, error) {
		return FloatVal(nilmath.Random()), nil
	})
	vm.globals["math_min"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Min(args[0].toFloat(), args[1].toFloat())), nil
	})
	vm.globals["math_max"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return FloatVal(0), nil
		}
		return FloatVal(nilmath.Max(args[0].toFloat(), args[1].toFloat())), nil
	})

	// Crypto
	vm.globals["crypto_sha256"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return StrVal(""), nil
		}
		return StrVal(nilcrypto.Sha256(args[0].String())), nil
	})
	vm.globals["crypto_md5"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return StrVal(""), nil
		}
		return StrVal(nilcrypto.Md5(args[0].String())), nil
	})
	vm.globals["crypto_uuid"] = NativeFuncVal(func(args []Value) (Value, error) {
		return StrVal(nilcrypto.UUID()), nil
	})

	// JSON
	vm.globals["json_stringify"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return StrVal("null"), nil
		}
		return StrVal(args[0].String()), nil
	})
	vm.globals["json_parse"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil, nil
		}
		parsed, err := niljson.Parse(args[0].String())
		if err != nil {
			return Nil, err
		}
		return toVMValue(parsed), nil
	})

	// Filesystem
	vm.globals["fs_read"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil, nil
		}
		content, err := nilfs.ReadTextFile(args[0].String())
		if err != nil {
			return Nil, err
		}
		return StrVal(content), nil
	})
	vm.globals["fs_write"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return Nil, nil
		}
		err := nilfs.WriteTextFile(args[0].String(), args[1].String())
		if err != nil {
			return Nil, err
		}
		return True, nil
	})
	vm.globals["fs_exists"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return False, nil
		}
		return BoolVal(nilfs.Exists(args[0].String())), nil
	})

	// Network
	vm.globals["http_get"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return Nil, nil
		}
		body, err := nilnet.Get(args[0].String())
		if err != nil {
			return Nil, err
		}
		return StrVal(body), nil
	})

	// Logging
	vm.globals["log_info"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) > 0 {
			nillog.Info(args[0].String())
		}
		return Nil, nil
	})
	vm.globals["log_warn"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) > 0 {
			nillog.Warn(args[0].String())
		}
		return Nil, nil
	})
	vm.globals["log_error"] = NativeFuncVal(func(args []Value) (Value, error) {
		if len(args) > 0 {
			nillog.Error(args[0].String())
		}
		return Nil, nil
	})

	// Functions from compiled module
	if vm.module != nil {
		for _, fn := range vm.module.Functions {
			vm.globals[fn.Name] = FuncVal(fn)
		}
	}
}

// Run executes the module's main function.
func (vm *VM) Run() error {
	if vm.module.MainFunc == nil {
		return fmt.Errorf("no main function")
	}
	return vm.callFunction(vm.module.MainFunc, nil)
}

func (vm *VM) push(v Value) {
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() Value {
	if len(vm.stack) == 0 {
		return Nil
	}
	v := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return v
}

func (vm *VM) peek() Value {
	if len(vm.stack) == 0 {
		return Nil
	}
	return vm.stack[len(vm.stack)-1]
}

func (vm *VM) callFunction(fn *codegen.Function, args []Value) error {
	frame := newFrame(fn)
	// load args into locals
	for i, arg := range args {
		if i < len(frame.locals) {
			frame.locals[i] = arg
		}
	}
	vm.frames = append(vm.frames, frame)
	err := vm.execute(frame)
	if len(vm.frames) > 0 {
		vm.frames = vm.frames[:len(vm.frames)-1]
	}
	return err
}

// findFunction looks up a named function in the module.
func (vm *VM) findFunction(name string) *codegen.Function {
	for _, fn := range vm.module.Functions {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

// ─── Execution loop ───────────────────────────────────────────────────────────

func (vm *VM) execute(frame *Frame) error {
	fn := frame.fn
	code := fn.Code
	consts := fn.Constants

	getConst := func(idx int32) codegen.Constant {
		if int(idx) < len(consts) {
			return consts[idx]
		}
		return codegen.Constant{}
	}

	for frame.ip < len(code) {
		ins := code[frame.ip]
		frame.ip++
		op := ins.Op
		oper := ins.Operand

		switch op {
		case codegen.OP_NOP:
			// nothing

		case codegen.OP_HALT:
			return nil

		// ── Constants ───────────────────────────────────────────────────────
		case codegen.OP_LOAD_CONST:
			c := getConst(oper)
			switch c.Kind {
			case codegen.ConstInt:
				vm.push(IntVal(c.IntVal))
			case codegen.ConstFloat:
				vm.push(FloatVal(c.FloatVal))
			case codegen.ConstString:
				vm.push(StrVal(c.StrVal))
			case codegen.ConstBool:
				vm.push(BoolVal(c.BoolVal))
			case codegen.ConstNull:
				vm.push(Nil)
			}

		case codegen.OP_LOAD_NULL:
			vm.push(Nil)
		case codegen.OP_LOAD_TRUE:
			vm.push(True)
		case codegen.OP_LOAD_FALSE:
			vm.push(False)
		case codegen.OP_LOAD_INT:
			vm.push(IntVal(int64(oper)))
		case codegen.OP_LOAD_FLOAT:
			bits := math.Float64frombits(uint64(oper))
			vm.push(FloatVal(bits))

		// ── Locals ──────────────────────────────────────────────────────────
		case codegen.OP_LOAD_LOCAL:
			idx := int(oper)
			if idx < len(frame.locals) {
				vm.push(frame.locals[idx])
			} else {
				vm.push(Nil)
			}
		case codegen.OP_STORE_LOCAL:
			v := vm.pop()
			idx := int(oper)
			for idx >= len(frame.locals) {
				frame.locals = append(frame.locals, Nil)
			}
			frame.locals[idx] = v

		// ── Globals ─────────────────────────────────────────────────────────
		case codegen.OP_LOAD_GLOBAL:
			nameVal := vm.pop()
			name := nameVal.StrVal
			if v, ok := vm.globals[name]; ok {
				vm.push(v)
			} else {
				// try to find function
				if f := vm.findFunction(name); f != nil {
					vm.push(FuncVal(f))
				} else {
					vm.push(Nil)
				}
			}
		case codegen.OP_STORE_GLOBAL:
			nameVal := vm.pop()
			v := vm.pop()
			vm.globals[nameVal.StrVal] = v

		// ── Fields ──────────────────────────────────────────────────────────
		case codegen.OP_LOAD_FIELD:
			nameVal := vm.pop()
			obj := vm.pop()
			name := nameVal.StrVal
			switch obj.Kind {
			case ValObject, ValMap:
				if v, ok := obj.MapVal[name]; ok {
					vm.push(v)
				} else if name == "length" {
					vm.push(IntVal(int64(len(obj.MapVal))))
				} else {
					vm.push(Nil)
				}
			case ValArray:
				if name == "length" {
					vm.push(IntVal(int64(len(obj.Slice))))
				} else {
					vm.push(vm.callArrayMethod(obj, name, nil))
				}
			case ValString:
				if name == "length" {
					vm.push(IntVal(int64(len(obj.StrVal))))
				} else {
					vm.push(vm.callStringMethod(obj, name, nil))
				}
			default:
				vm.push(Nil)
			}
		case codegen.OP_STORE_FIELD:
			nameVal := vm.pop()
			obj := vm.pop()
			v := vm.peek()
			if obj.MapVal == nil {
				obj.MapVal = make(map[string]Value)
			}
			obj.MapVal[nameVal.StrVal] = v

		// ── Index ────────────────────────────────────────────────────────────
		case codegen.OP_LOAD_INDEX:
			idx := vm.pop()
			arr := vm.pop()
			switch arr.Kind {
			case ValArray:
				i := int(idx.IntVal)
				if i >= 0 && i < len(arr.Slice) {
					vm.push(arr.Slice[i])
				} else {
					vm.push(Nil)
				}
			case ValMap, ValObject:
				key := idx.StrVal
				if v, ok := arr.MapVal[key]; ok {
					vm.push(v)
				} else {
					vm.push(Nil)
				}
			case ValString:
				i := int(idx.IntVal)
				if i >= 0 && i < len(arr.StrVal) {
					vm.push(StrVal(string(arr.StrVal[i])))
				} else {
					vm.push(Nil)
				}
			default:
				vm.push(Nil)
			}
		case codegen.OP_STORE_INDEX:
			val := vm.pop()
			idx := vm.pop()
			arr := vm.pop()
			if arr.Kind == ValArray {
				i := int(idx.IntVal)
				for i >= len(arr.Slice) {
					arr.Slice = append(arr.Slice, Nil)
				}
				arr.Slice[i] = val
			} else if arr.Kind == ValMap || arr.Kind == ValObject {
				if arr.MapVal == nil {
					arr.MapVal = make(map[string]Value)
				}
				arr.MapVal[idx.StrVal] = val
			}

		// ── Stack ops ────────────────────────────────────────────────────────
		case codegen.OP_POP:
			vm.pop()
		case codegen.OP_DUP:
			vm.push(vm.peek())

		// ── Arithmetic ───────────────────────────────────────────────────────
		case codegen.OP_ADD:
			b, a := vm.pop(), vm.pop()
			vm.push(vm.opAdd(a, b))
		case codegen.OP_STR_CONCAT:
			b, a := vm.pop(), vm.pop()
			vm.push(StrVal(a.String() + b.String()))
		case codegen.OP_SUB:
			b, a := vm.pop(), vm.pop()
			result, err := vm.numOp(a, b, "-")
			if err != nil {
				return err
			}
			vm.push(result)
		case codegen.OP_MUL:
			b, a := vm.pop(), vm.pop()
			result, err := vm.numOp(a, b, "*")
			if err != nil {
				return err
			}
			vm.push(result)
		case codegen.OP_DIV:
			b, a := vm.pop(), vm.pop()
			result, err := vm.numOp(a, b, "/")
			if err != nil {
				return err
			}
			vm.push(result)
		case codegen.OP_MOD:
			b, a := vm.pop(), vm.pop()
			result, err := vm.numOp(a, b, "%")
			if err != nil {
				return err
			}
			vm.push(result)
		case codegen.OP_POW:
			b, a := vm.pop(), vm.pop()
			vm.push(FloatVal(math.Pow(a.toFloat(), b.toFloat())))
		case codegen.OP_NEG:
			a := vm.pop()
			if a.Kind == ValInt {
				vm.push(IntVal(-a.IntVal))
			} else {
				vm.push(FloatVal(-a.FloatVal))
			}

		// ── Logic ────────────────────────────────────────────────────────────
		case codegen.OP_NOT:
			a := vm.pop()
			vm.push(BoolVal(!a.Truthy()))
		case codegen.OP_AND:
			b, a := vm.pop(), vm.pop()
			if a.Truthy() {
				vm.push(b)
			} else {
				vm.push(a)
			}
		case codegen.OP_OR:
			b, a := vm.pop(), vm.pop()
			if a.Truthy() {
				vm.push(a)
			} else {
				vm.push(b)
			}

		// ── Comparison ───────────────────────────────────────────────────────
		case codegen.OP_EQ:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(a.Equal(b)))
		case codegen.OP_NEQ:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(!a.Equal(b)))
		case codegen.OP_LT:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(a.toFloat() < b.toFloat()))
		case codegen.OP_GT:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(a.toFloat() > b.toFloat()))
		case codegen.OP_LTE:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(a.toFloat() <= b.toFloat()))
		case codegen.OP_GTE:
			b, a := vm.pop(), vm.pop()
			vm.push(BoolVal(a.toFloat() >= b.toFloat()))

		// ── Control flow ─────────────────────────────────────────────────────
		case codegen.OP_JUMP:
			frame.ip = int(oper)
		case codegen.OP_JUMP_IF_FALSE:
			cond := vm.pop()
			if !cond.Truthy() {
				frame.ip = int(oper)
			}
		case codegen.OP_JUMP_IF_TRUE:
			cond := vm.pop()
			if cond.Truthy() {
				frame.ip = int(oper)
			}
		case codegen.OP_JUMP_IF_NOT_NULL:
			cond := vm.pop()
			if cond.Kind != ValNil {
				frame.ip = int(oper)
			}

		// ── Function calls ───────────────────────────────────────────────────
		case codegen.OP_CALL:
			argc := int(oper)
			args := make([]Value, argc)
			for i := argc - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			callee := vm.pop()
			ret, err := vm.invoke(callee, args)
			if err != nil {
				return err
			}
			vm.push(ret)

		case codegen.OP_CALL_METHOD:
			argc := int(oper)
			methodNameVal := vm.pop()
			args := make([]Value, argc)
			for i := argc - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			obj := vm.pop()
			ret := vm.callMethod(obj, methodNameVal.StrVal, args)
			vm.push(ret)

		case codegen.OP_RETURN:
			retVal := vm.pop()
			vm.push(retVal) // return value left on stack for caller
			return nil
		case codegen.OP_RETURN_VOID:
			vm.push(Nil)
			return nil

		// ── Object/array creation ─────────────────────────────────────────────
		case codegen.OP_NEW_ARRAY:
			count := int(oper)
			elems := make([]Value, count)
			for i := count - 1; i >= 0; i-- {
				elems[i] = vm.pop()
			}
			vm.push(ArrayVal(elems))

		case codegen.OP_NEW_OBJECT:
			count := int(oper)
			m := make(map[string]Value)
			pairs := make([]struct{ k, v Value }, count)
			for i := count - 1; i >= 0; i-- {
				v := vm.pop()
				k := vm.pop()
				pairs[i] = struct{ k, v Value }{k, v}
			}
			for _, p := range pairs {
				m[p.k.String()] = p.v
			}
			vm.push(ObjVal(m))

		case codegen.OP_NEW_MAP:
			vm.push(ObjVal(make(map[string]Value)))

		case codegen.OP_NEW_CLASS:
			nameVal := vm.pop()
			obj := ObjVal(make(map[string]Value))
			obj.MapVal["__class__"] = nameVal
			vm.push(obj)

		// ── Print ─────────────────────────────────────────────────────────────
		case codegen.OP_PRINT:
			argc := int(oper)
			parts := make([]string, argc)
			for i := argc - 1; i >= 0; i-- {
				parts[i] = vm.pop().String()
			}
			line := strings.Join(parts, " ")
			fmt.Fprintln(&vm.output, line)
			fmt.Println(line)

		// ── Async / channels ──────────────────────────────────────────────────
		case codegen.OP_AWAIT:
			// In VM mode, await is synchronous (no actual async scheduler yet)
			// Value is already resolved on stack
		case codegen.OP_SPAWN_TASK:
			callee := vm.pop()
			go func(c Value) {
				_, _ = vm.invoke(c, nil)
			}(callee)
		case codegen.OP_CHAN_SEND:
			v := vm.pop()
			ch := vm.pop()
			if ch.Kind == ValChannel && ch.Chan != nil {
				ch.Chan.Send(v)
			}
		case codegen.OP_CHAN_RECV:
			ch := vm.pop()
			if ch.Kind == ValChannel && ch.Chan != nil {
				v, _ := ch.Chan.Recv()
				vm.push(v)
			} else {
				vm.push(Nil)
			}

		// ── UI ops ───────────────────────────────────────────────────────────
		case codegen.OP_UI_CREATE:
			widgetTypeVal := vm.pop()
			if vm.UITree != nil {
				vm.UITree.BeginNode(widgetTypeVal.String())
			}
		case codegen.OP_UI_PROP:
			keyVal := vm.pop()
			val := vm.pop()
			if vm.UITree != nil {
				vm.UITree.SetProp(keyVal.String(), val.toRawValue())
			}
		case codegen.OP_UI_EVENT:
			handlerVal := vm.pop()
			eventVal := vm.pop()
			if vm.UITree != nil {
				hCopy := handlerVal
				vm.UITree.SetEvent(eventVal.String(), func(args ...interface{}) {
					_, _ = vm.invoke(hCopy, nil)
				})
			}
		case codegen.OP_UI_CHILD:
			// Automatically structured by BeginNode / EndNode
		case codegen.OP_UI_END:
			if vm.UITree != nil {
				vm.UITree.EndNode()
			}

		// ── Error handling ────────────────────────────────────────────────────
		case codegen.OP_THROW:
			v := vm.pop()
			return fmt.Errorf("NilLang panic: %s", v.String())
		case codegen.OP_TRY_BEGIN:
			// patch handled by VM try-catch (simplified: just continue)
		case codegen.OP_TRY_END:
		case codegen.OP_CATCH_BEGIN:
		}
	}
	return nil
}

// invoke calls a Value as a function.
func (vm *VM) invoke(callee Value, args []Value) (Value, error) {
	if callee.Kind == ValString {
		if fn := vm.findFunction(callee.StrVal); fn != nil {
			callee = FuncVal(fn)
		} else if g, ok := vm.globals[callee.StrVal]; ok {
			callee = g
		}
	}
	if callee.Kind == ValNativeFunc && callee.NativeVal != nil {
		return callee.NativeVal(args)
	}
	if callee.Kind != ValFunc || callee.FuncVal == nil {
		return Nil, fmt.Errorf("cannot call non-function value: %s", callee.String())
	}
	fn := callee.FuncVal.Fn
	stackBefore := len(vm.stack)
	err := vm.callFunction(fn, args)
	if err != nil {
		return Nil, err
	}
	if len(vm.stack) > stackBefore {
		ret := vm.pop()
		return ret, nil
	}
	return Nil, nil
}

// FindFunction returns a function by name if defined in the module.
func (vm *VM) FindFunction(name string) *codegen.Function {
	return vm.findFunction(name)
}

// InvokeFunction invokes a named function with arguments in the VM.
func (vm *VM) InvokeFunction(name string, args []Value) (Value, error) {
	return vm.invoke(StrVal(name), args)
}

// opAdd handles + (numeric or string concat).
func (vm *VM) opAdd(a, b Value) Value {
	if a.Kind == ValString || b.Kind == ValString {
		return StrVal(a.String() + b.String())
	}
	if a.Kind == ValInt && b.Kind == ValInt {
		return IntVal(a.IntVal + b.IntVal)
	}
	return FloatVal(a.toFloat() + b.toFloat())
}

func (vm *VM) numOp(a, b Value, op string) (Value, error) {
	if a.Kind == ValInt && b.Kind == ValInt {
		switch op {
		case "-":
			return IntVal(a.IntVal - b.IntVal), nil
		case "*":
			return IntVal(a.IntVal * b.IntVal), nil
		case "/":
			if b.IntVal == 0 {
				return Nil, fmt.Errorf("integer division by zero")
			}
			return IntVal(a.IntVal / b.IntVal), nil
		case "%":
			if b.IntVal == 0 {
				return Nil, fmt.Errorf("integer modulo by zero")
			}
			return IntVal(a.IntVal % b.IntVal), nil
		}
	}
	fa, fb := a.toFloat(), b.toFloat()
	switch op {
	case "-":
		return FloatVal(fa - fb), nil
	case "*":
		return FloatVal(fa * fb), nil
	case "/":
		if fb == 0 {
			return Nil, fmt.Errorf("floating-point division by zero")
		}
		return FloatVal(fa / fb), nil
	case "%":
		if fb == 0 {
			return Nil, fmt.Errorf("floating-point modulo by zero")
		}
		return FloatVal(math.Mod(fa, fb)), nil
	}
	return Nil, fmt.Errorf("unsupported numeric operator %q", op)
}

// callMethod dispatches method calls on values.
func (vm *VM) callMethod(obj Value, method string, args []Value) Value {
	if method == "toString" {
		return StrVal(obj.String())
	}
	switch obj.Kind {
	case ValArray:
		return vm.callArrayMethod(obj, method, args)
	case ValString:
		return vm.callStringMethod(obj, method, args)
	case ValMap, ValObject:
		return vm.callMapMethod(obj, method, args)
	case ValFunc:
		if method == "call" || method == "apply" {
			ret, _ := vm.invoke(obj, args)
			return ret
		}
	}
	return Nil
}

func (vm *VM) callArrayMethod(arr Value, method string, args []Value) Value {
	switch method {
	case "push":
		for _, a := range args {
			arr.Slice = append(arr.Slice, a)
		}
		return IntVal(int64(len(arr.Slice)))
	case "pop":
		if len(arr.Slice) == 0 {
			return Nil
		}
		v := arr.Slice[len(arr.Slice)-1]
		arr.Slice = arr.Slice[:len(arr.Slice)-1]
		return v
	case "length":
		return IntVal(int64(len(arr.Slice)))
	case "join":
		sep := ", "
		if len(args) > 0 {
			sep = args[0].StrVal
		}
		parts := make([]string, len(arr.Slice))
		for i, v := range arr.Slice {
			parts[i] = v.String()
		}
		return StrVal(strings.Join(parts, sep))
	case "includes":
		if len(args) > 0 {
			for _, v := range arr.Slice {
				if v.Equal(args[0]) {
					return True
				}
			}
		}
		return False
	case "indexOf":
		if len(args) > 0 {
			for i, v := range arr.Slice {
				if v.Equal(args[0]) {
					return IntVal(int64(i))
				}
			}
		}
		return IntVal(-1)
	case "reverse":
		n := len(arr.Slice)
		for i := 0; i < n/2; i++ {
			arr.Slice[i], arr.Slice[n-1-i] = arr.Slice[n-1-i], arr.Slice[i]
		}
		return arr
	case "slice":
		start, end := 0, len(arr.Slice)
		if len(args) > 0 {
			start = int(args[0].IntVal)
		}
		if len(args) > 1 {
			end = int(args[1].IntVal)
		}
		return ArrayVal(arr.Slice[start:end])
	case "toString":
		return StrVal(arr.String())
	}
	return Nil
}

func (vm *VM) callStringMethod(s Value, method string, args []Value) Value {
	str := s.StrVal
	switch method {
	case "length":
		return IntVal(int64(len(str)))
	case "toUpperCase":
		return StrVal(strings.ToUpper(str))
	case "toLowerCase":
		return StrVal(strings.ToLower(str))
	case "trim":
		return StrVal(strings.TrimSpace(str))
	case "trimStart":
		return StrVal(strings.TrimLeft(str, " \t\n\r"))
	case "trimEnd":
		return StrVal(strings.TrimRight(str, " \t\n\r"))
	case "includes":
		if len(args) > 0 {
			return BoolVal(strings.Contains(str, args[0].StrVal))
		}
		return False
	case "startsWith":
		if len(args) > 0 {
			return BoolVal(strings.HasPrefix(str, args[0].StrVal))
		}
		return False
	case "endsWith":
		if len(args) > 0 {
			return BoolVal(strings.HasSuffix(str, args[0].StrVal))
		}
		return False
	case "indexOf":
		if len(args) > 0 {
			return IntVal(int64(strings.Index(str, args[0].StrVal)))
		}
		return IntVal(-1)
	case "split":
		sep := ""
		if len(args) > 0 {
			sep = args[0].StrVal
		}
		parts := strings.Split(str, sep)
		vals := make([]Value, len(parts))
		for i, p := range parts {
			vals[i] = StrVal(p)
		}
		return ArrayVal(vals)
	case "replace":
		if len(args) >= 2 {
			return StrVal(strings.ReplaceAll(str, args[0].StrVal, args[1].StrVal))
		}
		return s
	case "charAt":
		if len(args) > 0 {
			i := int(args[0].IntVal)
			if i >= 0 && i < len(str) {
				return StrVal(string(str[i]))
			}
		}
		return StrVal("")
	case "substring":
		start, end := 0, len(str)
		if len(args) > 0 {
			start = int(args[0].IntVal)
		}
		if len(args) > 1 {
			end = int(args[1].IntVal)
		}
		if start < 0 {
			start = 0
		}
		if end > len(str) {
			end = len(str)
		}
		return StrVal(str[start:end])
	case "toString":
		return s
	case "repeat":
		if len(args) > 0 {
			n := int(args[0].IntVal)
			return StrVal(strings.Repeat(str, n))
		}
		return s
	}
	return Nil
}

func (vm *VM) callMapMethod(obj Value, method string, args []Value) Value {
	switch method {
	case "get":
		if len(args) > 0 {
			if v, ok := obj.MapVal[args[0].String()]; ok {
				return v
			}
		}
		return Nil
	case "set":
		if len(args) >= 2 {
			if obj.MapVal == nil {
				obj.MapVal = make(map[string]Value)
			}
			obj.MapVal[args[0].String()] = args[1]
		}
		return obj
	case "has":
		if len(args) > 0 {
			_, ok := obj.MapVal[args[0].String()]
			return BoolVal(ok)
		}
		return False
	case "delete":
		if len(args) > 0 {
			delete(obj.MapVal, args[0].String())
		}
		return obj
	case "keys":
		keys := make([]Value, 0, len(obj.MapVal))
		for k := range obj.MapVal {
			keys = append(keys, StrVal(k))
		}
		return ArrayVal(keys)
	case "values":
		vals := make([]Value, 0, len(obj.MapVal))
		for _, v := range obj.MapVal {
			vals = append(vals, v)
		}
		return ArrayVal(vals)
	case "size", "length":
		return IntVal(int64(len(obj.MapVal)))
	}
	return Nil
}
