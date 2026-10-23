// Package engine provides the core UI tree model, reactive state binding, and rendering abstractions for Alap.
package engine

import (
	"encoding/json"
	"fmt"
	"sync"
)

// UITree manages the active declarative UI component tree and event dispatching.
type UITree struct {
	mu        sync.RWMutex
	Root      *WidgetNode
	nodeStack []*WidgetNode
	nextID    int
	nodeMap   map[string]*WidgetNode
}

// NewUITree creates a new UI tree manager.
func NewUITree() *UITree {
	return &UITree{
		nodeStack: make([]*WidgetNode, 0),
		nodeMap:   make(map[string]*WidgetNode),
	}
}

// NewTree is an alias for NewUITree.
func NewTree() *UITree {
	return NewUITree()
}

// BeginNode creates a new widget node and pushes it onto the active parent stack.
func (t *UITree) BeginNode(widgetType string) *WidgetNode {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nextID++
	id := fmt.Sprintf("node_%d", t.nextID)
	node := NewWidgetNode(id, widgetType)
	t.nodeMap[id] = node

	if len(t.nodeStack) > 0 {
		parent := t.nodeStack[len(t.nodeStack)-1]
		parent.AddChild(node)
	} else if t.Root == nil {
		t.Root = node
	}

	t.nodeStack = append(t.nodeStack, node)
	return node
}

// CurrentNode returns the node currently at the top of the stack.
func (t *UITree) CurrentNode() *WidgetNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.nodeStack) == 0 {
		return nil
	}
	return t.nodeStack[len(t.nodeStack)-1]
}

// SetProp sets a property on the currently active node.
func (t *UITree) SetProp(key string, val interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.nodeStack) > 0 {
		curr := t.nodeStack[len(t.nodeStack)-1]
		curr.SetProp(key, val)
	}
}

// SetEvent sets an event handler on the currently active node.
func (t *UITree) SetEvent(name string, handler EventHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.nodeStack) > 0 {
		curr := t.nodeStack[len(t.nodeStack)-1]
		curr.SetEvent(name, handler)
	}
}

// EndNode pops the current node from the stack.
func (t *UITree) EndNode() *WidgetNode {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.nodeStack) == 0 {
		return nil
	}
	popped := t.nodeStack[len(t.nodeStack)-1]
	t.nodeStack = t.nodeStack[:len(t.nodeStack)-1]
	return popped
}

// FindNodeByID returns the widget node with the specified ID.
func (t *UITree) FindNodeByID(id string) *WidgetNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodeMap[id]
}

// DispatchEvent finds a node and invokes its event handler.
func (t *UITree) DispatchEvent(nodeID, eventName string, args ...interface{}) bool {
	t.mu.RLock()
	node, exists := t.nodeMap[nodeID]
	t.mu.RUnlock()

	if !exists || node == nil {
		return false
	}
	return node.TriggerEvent(eventName, args...)
}

// HitTest finds the topmost widget node containing (x, y).
func (t *UITree) HitTest(x, y float64) *WidgetNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return hitTestNode(t.Root, x, y)
}

func hitTestNode(node *WidgetNode, x, y float64) *WidgetNode {
	if node == nil {
		return nil
	}
	for i := len(node.Children) - 1; i >= 0; i-- {
		if hit := hitTestNode(node.Children[i], x, y); hit != nil {
			return hit
		}
	}
	b := node.Bounds
	if x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height {
		return node
	}
	return nil
}

// DispatchTouch checks which leaf node contains (x, y) and triggers "onClick".
func (t *UITree) DispatchTouch(x, y float64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Root == nil {
		return false
	}
	return hitTestAndTrigger(t.Root, x, y)
}

func hitTestAndTrigger(node *WidgetNode, x, y float64) bool {
	if node == nil {
		return false
	}
	// Check children in reverse order (topmost first)
	for i := len(node.Children) - 1; i >= 0; i-- {
		if hitTestAndTrigger(node.Children[i], x, y) {
			return true
		}
	}
	// Check self bounds
	b := node.Bounds
	if x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height {
		return node.TriggerEvent("onClick", x, y)
	}
	return false
}

// ToJSON serializes the UI tree to JSON for bridge transmission.
func (t *UITree) ToJSON() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return json.Marshal(t.Root)
}

// RenderTextTree prints an ASCII representation of the UI tree.
func (t *UITree) RenderTextTree() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Root == nil {
		return "<Empty UI Tree>"
	}
	return t.Root.DumpString(0)
}
