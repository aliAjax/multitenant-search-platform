package platform

import (
	"context"
	"errors"
	"io"
	"testing"
)

type trackedWriter struct{ active, max *int }

func (w *trackedWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *trackedWriter) Close() error                { *w.active--; return nil }

var _ io.WriteCloser = (*trackedWriter)(nil)

func TestAppendBatchClosesEachWriter(t *testing.T) {
	active, max := 0, 0
	var log EventLog
	open := func() (io.WriteCloser, error) {
		active++
		if active > max {
			max = active
		}
		return &trackedWriter{active: &active, max: &max}, nil
	}
	if err := log.AppendBatchWithWriter(context.Background(), []any{"a", "b", "c"}, open); err != nil {
		t.Fatal(err)
	}
	if max != 1 || active != 0 {
		t.Fatalf("writers stayed open during batch: max=%d active=%d", max, active)
	}
}

func TestRunWithFinalizerPreservesWorkError(t *testing.T) {
	workErr := errors.New("work failed")
	finalErr := errors.New("finalizer failed")
	r := NewTaskRunner(RealClock{})
	err := r.RunWithFinalizer(context.Background(), "task", func(context.Context) error { return workErr }, func() error { return finalErr })
	if !errors.Is(err, workErr) {
		t.Fatalf("work error was swallowed: %v", err)
	}
}
