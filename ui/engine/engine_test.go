package engine

import (
	"strings"
	"testing"
)

func TestBuildTreeLinksParentsAndChildren(t *testing.T) {
	tree := NewTree()
	root := tree.BeginNode("Column")
	child := tree.BeginNode("Text")
	tree.SetProp("text", "hello")
	tree.EndNode()
	tree.EndNode()

	if tree.Root != root {
		t.Fatalf("tree root mismatch")
	}
	if len(root.Children) != 1 || root.Children[0] != child {
		t.Fatalf("expected the child to be attached to the root")
	}
	if child.Parent != root {
		t.Fatalf("child.Parent was not set")
	}
	if got := child.GetStringProp("text", ""); got != "hello" {
		t.Fatalf("text prop = %q, want hello", got)
	}
	if tree.FindNodeByID(child.ID) != child {
		t.Fatalf("FindNodeByID did not return the child")
	}
}

func TestEndNodeOnEmptyStackReturnsNil(t *testing.T) {
	if got := NewTree().EndNode(); got != nil {
		t.Fatalf("EndNode() on an empty stack = %v, want nil", got)
	}
}

func TestHitTestAndDispatchTouch(t *testing.T) {
	tree := NewTree()
	btn := tree.BeginNode("Button")
	btn.Bounds = Rect{X: 0, Y: 0, Width: 100, Height: 50}
	clicked := false
	btn.SetEvent("onClick", func(args ...interface{}) { clicked = true })
	tree.EndNode()

	if got := tree.HitTest(10, 10); got != btn {
		t.Fatalf("HitTest(10,10) = %v, want the button", got)
	}
	if tree.HitTest(500, 500) != nil {
		t.Fatalf("HitTest outside bounds should be nil")
	}
	if !tree.DispatchTouch(10, 10) {
		t.Fatalf("DispatchTouch inside bounds should report handled")
	}
	if !clicked {
		t.Fatalf("onClick handler was not invoked")
	}
	if tree.DispatchTouch(500, 500) {
		t.Fatalf("DispatchTouch outside bounds should not report handled")
	}
}

func TestRenderTextTreeAndJSON(t *testing.T) {
	tree := NewTree()
	tree.BeginNode("Column")
	tree.BeginNode("Text")
	tree.EndNode()
	tree.EndNode()

	rendered := tree.RenderTextTree()
	if !strings.Contains(rendered, "Column") || !strings.Contains(rendered, "Text") {
		t.Fatalf("RenderTextTree missing nodes:\n%s", rendered)
	}

	data, err := tree.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error = %v", err)
	}
	if !strings.Contains(string(data), "Column") {
		t.Fatalf("ToJSON output = %s", data)
	}
}

func TestEmptyTreeRender(t *testing.T) {
	if got := NewTree().RenderTextTree(); got != "<Empty UI Tree>" {
		t.Fatalf("empty tree render = %q", got)
	}
}

func TestGetFloatPropCoercions(t *testing.T) {
	n := NewWidgetNode("id", "Box")
	n.SetProp("f", 1.5)
	n.SetProp("i", 2)

	if got := n.GetFloatProp("f", 0); got != 1.5 {
		t.Fatalf("float prop = %v, want 1.5", got)
	}
	if got := n.GetFloatProp("i", 0); got != 2 {
		t.Fatalf("int prop = %v, want 2", got)
	}
	if got := n.GetFloatProp("missing", 9); got != 9 {
		t.Fatalf("missing prop = %v, want fallback 9", got)
	}
}
