package bulk

import (
	"context"
	"sync"
)

func Process(ctx context.Context, in []Item) ([]Item, error) {
	producer := Produce(ctx, in)
	merged := make(chan Item)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for n := 0; n < 2; n++ {
		go func() {
			<-start
			wg.Add(1)
			defer wg.Done()
			for item := range producer {
				merged <- item
			}
		}()
	}
	close(start)
	go func() {
		wg.Wait()
		close(merged)
	}()
	return Consume(ctx, merged)
}
