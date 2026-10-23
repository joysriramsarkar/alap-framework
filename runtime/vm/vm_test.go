package vm

import "testing"

func TestNumOpRejectsIntegerDivisionByZero(t *testing.T) {
	_, err := (&VM{}).numOp(IntVal(10), IntVal(0), "/")
	if err == nil {
		t.Fatal("expected integer division by zero to return an error")
	}
}

func TestNumOpRejectsFloatModuloByZero(t *testing.T) {
	_, err := (&VM{}).numOp(FloatVal(10), FloatVal(0), "%")
	if err == nil {
		t.Fatal("expected floating-point modulo by zero to return an error")
	}
}
