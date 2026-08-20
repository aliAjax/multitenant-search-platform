package platform

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TaskState string

const (
	TaskPending TaskState = "pending"
	TaskRunning TaskState = "running"
	TaskPaused  TaskState = "paused"
	TaskDone    TaskState = "done"
	TaskFailed  TaskState = "failed"
)

type Task struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	State     TaskState `json:"state"`
	Progress  int       `json:"progress"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
type TaskRunner struct {
	mu    sync.Mutex
	tasks map[string]Task
	clock Clock
}

func NewTaskRunner(c Clock) *TaskRunner { return &TaskRunner{tasks: map[string]Task{}, clock: c} }
func (r *TaskRunner) Create(kind string) Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := Task{ID: NewID("task"), Kind: kind, State: TaskPending, UpdatedAt: r.clock.Now()}
	r.tasks[t.ID] = t
	return t
}
func (r *TaskRunner) Update(id string, state TaskState, progress int, err error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return ErrNotFound
	}
	t.State = state
	t.Progress = progress
	t.UpdatedAt = r.clock.Now()
	if err != nil {
		t.Error = err.Error()
	}
	r.tasks[id] = t
	return nil
}
func (r *TaskRunner) Get(id string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return t, fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return t, nil
}
func (r *TaskRunner) Run(ctx context.Context, id string, fn func(context.Context) error) error {
	if e := r.Update(id, TaskRunning, 0, nil); e != nil {
		return e
	}
	e := fn(ctx)
	if e != nil {
		_ = r.Update(id, TaskFailed, 0, e)
		return e
	}
	return r.Update(id, TaskDone, 100, nil)
}

func (r *TaskRunner) RunWithFinalizer(ctx context.Context, id string, fn func(context.Context) error, finalizer func() error) (err error) {
	err = fn(ctx)
	// Run the finalizer unconditionally, but never let its error clobber a
	// work error that already occurred — the original failure is what callers
	// need to see.
	if e := finalizer(); e != nil && err == nil {
		err = e
	}
	return err
}
