package json

import "testing"

func TestStringifyAndParseRoundtrip(t *testing.T) {
	in := map[string]interface{}{"name": "onuron", "port": float64(4324)}

	s, err := Stringify(in, false)
	if err != nil {
		t.Fatalf("Stringify error = %v", err)
	}

	out, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok || m["name"] != "onuron" {
		t.Fatalf("roundtrip mismatch: %v", out)
	}
}

func TestStringifyPrettyIsNonEmpty(t *testing.T) {
	s, err := Stringify(map[string]interface{}{"a": float64(1)}, true)
	if err != nil {
		t.Fatalf("Stringify error = %v", err)
	}
	if len(s) == 0 {
		t.Fatalf("pretty output is empty")
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("{not json"); err == nil {
		t.Fatalf("expected Parse to fail on invalid JSON")
	}
	if Valid("{not json") {
		t.Fatalf("Valid reported invalid JSON as valid")
	}
	if !Valid(`{"ok":true}`) {
		t.Fatalf("Valid reported valid JSON as invalid")
	}
}
