package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSpawnAndAwaitRunsWork(t *testing.T) {
	s := New(2)
	defer s.Stop()

	ran := false
	task := s.Spawn("work", func() error {
		ran = true
		return nil
	})
	if task.ID == 0 {
		t.Fatalf("expected a non-zero task ID")
	}
	if err := task.Await(); err != nil {
		t.Fatalf("Await() error = %v, want nil", err)
	}
	if !ran {
		t.Fatalf("task work function was not executed")
	}
}

func TestSpawnPropagatesError(t *testing.T) {
	s := New(1)
	defer s.Stop()

	want := errors.New("boom")
	task := s.Spawn("fail", func() error { return want })
	if err := task.Await(); !errors.Is(err, want) {
		t.Fatalf("Await() error = %v, want %v", err, want)
	}
}

func TestAwaitTimeoutWhileWorkerBusy(t *testing.T) {
	s := New(1)
	defer s.Stop()

	release := make(chan struct{})
	blocker := s.Spawn("blocker", func() error {
		<-release
		return nil
	})
	queued := s.Spawn("queued", func() error { return nil })

	if err := queued.AwaitTimeout(50 * time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AwaitTimeout() = %v, want context.DeadlineExceeded", err)
	}

	close(release)
	if err := blocker.Await(); err != nil {
		t.Fatalf("blocker Await() error = %v", err)
	}
}

func TestNewDefaultsWorkerCount(t *testing.T) {
	// A non-positive worker count must still start a usable scheduler.
	s := New(0)
	defer s.Stop()
	if err := s.Spawn("noop", func() error { return nil }).Await(); err != nil {
		t.Fatalf("Await() error = %v", err)
	}
}
