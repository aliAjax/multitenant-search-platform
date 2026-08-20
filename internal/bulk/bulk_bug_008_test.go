package bulk

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestProcessErrorBranchCompletes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			items, err := Process(ctx, []Item{{Value: 1}, {Err: errors.New("bad row")}, {Value: 3}})
			if err != nil || len(items) != 1 || items[0].Value != 1 {
				results <- errors.New("error branch did not complete")
				return
			}
			results <- nil
		}()
	}
	close(start)
	wg.Wait()
	for n := 0; n < 2; n++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestProcessNormalBranchCompletes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			items, err := Process(ctx, []Item{{Value: 1}, {Value: 3}})
			if err != nil || len(items) != 2 {
				results <- errors.New("normal branch did not complete")
				return
			}
			results <- nil
		}()
	}
	close(start)
	wg.Wait()
	for n := 0; n < 2; n++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestProduceClosesOnError(t *testing.T) {
	ch := Produce(context.Background(), []Item{{Value: 1}, {Err: errors.New("bad row")}})
	if _, ok := <-ch; !ok {
		t.Fatal("producer closed before delivering the first item")
	}
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("producer emitted an item after the error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("producer did not close after the error")
	}
}
