package actor

import "testing"

func TestSendAndAsk(t *testing.T) {
	a := New("Counter", 16)
	var total int64

	a.On("Add", func(msg Message) interface{} {
		total += msg.Payload.(int64)
		return total
	})
	a.On("Get", func(msg Message) interface{} { return total })

	a.Start()
	defer a.Stop()

	if err := a.Send("Add", int64(4)); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := a.Send("Add", int64(6)); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	got, err := a.Ask("Get", nil)
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if got.(int64) != 10 {
		t.Fatalf("counter = %v, want 10", got)
	}
}

func TestSendAndAskBeforeStartFail(t *testing.T) {
	a := New("Idle", 4)
	if err := a.Send("x", nil); err == nil {
		t.Fatalf("expected Send() to fail before Start()")
	}
	if _, err := a.Ask("x", nil); err == nil {
		t.Fatalf("expected Ask() to fail before Start()")
	}
}

func TestUnhandledAskReturnsNil(t *testing.T) {
	a := New("Quiet", 4)
	a.Start()
	defer a.Stop()

	res, err := a.Ask("missing", nil)
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if res != nil {
		t.Fatalf("unhandled Ask() = %v, want nil", res)
	}
}

func TestDefaultMailboxCapacity(t *testing.T) {
	a := New("Default", 0)
	a.On("Ping", func(msg Message) interface{} { return "pong" })
	a.Start()
	defer a.Stop()

	if res, err := a.Ask("Ping", nil); err != nil || res != "pong" {
		t.Fatalf("Ask() = (%v, %v), want (pong, nil)", res, err)
	}
}
