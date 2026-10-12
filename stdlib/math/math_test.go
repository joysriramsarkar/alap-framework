package math

import "testing"

func TestBasicMath(t *testing.T) {
	if got := Abs(-3); got != 3 {
		t.Errorf("Abs(-3) = %v, want 3", got)
	}
	if got := Sqrt(16); got != 4 {
		t.Errorf("Sqrt(16) = %v, want 4", got)
	}
	if got := Min(2, 5); got != 2 {
		t.Errorf("Min(2,5) = %v, want 2", got)
	}
	if got := Max(2, 5); got != 5 {
		t.Errorf("Max(2,5) = %v, want 5", got)
	}
	if got := Pow(2, 10); got != 1024 {
		t.Errorf("Pow(2,10) = %v, want 1024", got)
	}
	if got := Floor(1.9); got != 1 {
		t.Errorf("Floor(1.9) = %v, want 1", got)
	}
	if got := Ceil(1.1); got != 2 {
		t.Errorf("Ceil(1.1) = %v, want 2", got)
	}
	if got := Round(1.5); got != 2 {
		t.Errorf("Round(1.5) = %v, want 2", got)
	}
}

func TestClamp(t *testing.T) {
	if got := Clamp(10, 0, 5); got != 5 {
		t.Errorf("Clamp(10,0,5) = %v, want 5", got)
	}
	if got := Clamp(-1, 0, 5); got != 0 {
		t.Errorf("Clamp(-1,0,5) = %v, want 0", got)
	}
	if got := Clamp(3, 0, 5); got != 3 {
		t.Errorf("Clamp(3,0,5) = %v, want 3", got)
	}
}

func TestRandomRanges(t *testing.T) {
	if r := Random(); r < 0 || r >= 1 {
		t.Errorf("Random() = %v, want [0,1)", r)
	}
	for i := 0; i < 100; i++ {
		if n := RandomInt(5, 10); n < 5 || n > 10 {
			t.Fatalf("RandomInt(5,10) = %d, want [5,10]", n)
		}
	}
}
