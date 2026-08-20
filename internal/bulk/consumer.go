package bulk

import "context"

func Consume(ctx context.Context, in <-chan Item) ([]Item, error) {
	out := []Item{}
	for {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		case item, ok := <-in:
			if !ok {
				return out, nil
			}
			out = append(out, item)
		}
	}
}
