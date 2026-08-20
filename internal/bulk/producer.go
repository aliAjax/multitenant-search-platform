package bulk

import "context"

type Item struct {
	Value int
	Err   error
}

func Produce(ctx context.Context, in []Item) <-chan Item {
	out := make(chan Item)
	go func() {
		for _, item := range in {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if item.Err != nil {
				return
			}
			out <- item
		}
		close(out)
	}()
	return out
}
