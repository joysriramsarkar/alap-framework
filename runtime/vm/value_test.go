package vm

import "testing"

func TestValueConstructorsToString(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want string
	}{
		{"nil", Nil, "null"},
		{"true", True, "true"},
		{"false", False, "false"},
		{"int", IntVal(42), "42"},
		{"string", StrVal("hi"), "hi"},
		{"array", ArrayVal([]Value{IntVal(1), IntVal(2)}), "[1, 2]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBoolValSetsKindAndValue(t *testing.T) {
	if v := BoolVal(true); v.Kind != ValBool || !v.BoolVal {
		t.Fatalf("BoolVal(true) = %+v, want a true bool value", v)
	}
	if v := BoolVal(false); v.Kind != ValBool || v.BoolVal {
		t.Fatalf("BoolVal(false) = %+v, want a false bool value", v)
	}
}

func TestTruthy(t *testing.T) {
	falsy := []Value{Nil, False, IntVal(0), FloatVal(0), StrVal(""), ArrayVal(nil)}
	for i, v := range falsy {
		if v.Truthy() {
			t.Errorf("falsy[%d] (%s) reported truthy", i, v.String())
		}
	}
	truthy := []Value{True, IntVal(1), StrVal("x"), ArrayVal([]Value{IntVal(1)})}
	for i, v := range truthy {
		if !v.Truthy() {
			t.Errorf("truthy[%d] (%s) reported falsy", i, v.String())
		}
	}
}

func TestEqual(t *testing.T) {
	if !IntVal(1).Equal(IntVal(1)) {
		t.Errorf("1 == 1 should be true")
	}
	if IntVal(1).Equal(IntVal(2)) {
		t.Errorf("1 == 2 should be false")
	}
	if !StrVal("a").Equal(StrVal("a")) {
		t.Errorf("string equality failed")
	}
	// Numeric cross-kind equality: int 1 and float 1.0 compare by value.
	if !IntVal(1).Equal(FloatVal(1.0)) {
		t.Errorf("1 (int) should equal 1.0 (float)")
	}
	if !Nil.Equal(Nil) {
		t.Errorf("nil == nil should be true")
	}
	if True.Equal(IntVal(1)) {
		t.Errorf("true should not equal int 1")
	}
}

func TestChannelSendRecv(t *testing.T) {
	ch := NewChannel(1)
	ch.Send(IntVal(7))
	v, ok := ch.Recv()
	if !ok {
		t.Fatalf("Recv reported a closed channel")
	}
	if v.Kind != ValInt || v.IntVal != 7 {
		t.Fatalf("Recv() = %s, want 7", v.String())
	}
}
