package widgets

import "testing"

func TestTextWidget(t *testing.T) {
	n := Text("hi")
	if n.Type != "Text" {
		t.Fatalf("Type = %q, want Text", n.Type)
	}
	if got := n.GetStringProp("text", ""); got != "hi" {
		t.Fatalf("text = %q, want hi", got)
	}
}

func TestContainerBuildsChildren(t *testing.T) {
	c := Column(Text("a"), Row(Text("b")))
	if len(c.Children) != 2 {
		t.Fatalf("Column children = %d, want 2", len(c.Children))
	}
	if c.Children[1].Type != "Row" || len(c.Children[1].Children) != 1 {
		t.Fatalf("nested Row was not built correctly")
	}
}

func TestButtonFiresHandler(t *testing.T) {
	fired := false
	b := Button("Click", func(args ...interface{}) { fired = true })
	if !b.TriggerEvent("onClick") {
		t.Fatalf("button has no onClick handler")
	}
	if !fired {
		t.Fatalf("onClick handler was not invoked")
	}
}

func TestSwitchValue(t *testing.T) {
	s := Switch(true, nil)
	if v, ok := s.Props["value"].(bool); !ok || !v {
		t.Fatalf("switch value prop was not set to true")
	}
}
