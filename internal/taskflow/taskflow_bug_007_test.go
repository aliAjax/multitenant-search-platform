package taskflow

import "testing"

func TestRetryTransitionsToSucceeded(t *testing.T) {
	s := New()
	s.Set("job", Failed)
	if err := s.Retry("job"); err != nil {
		t.Fatal(err)
	}
	if err := s.Complete("job"); err != nil {
		t.Fatal(err)
	}
	if s.State("job") != Succeeded {
		t.Fatalf("wrong final state: %s", s.State("job"))
	}
}

func TestRetryingIsActive(t *testing.T) {
	s := New()
	s.Set("job", Retrying)
	active := Active(s.states)
	if len(active) != 1 || active[0] != "job" {
		t.Fatalf("retrying job disappeared: %#v", active)
	}
}
