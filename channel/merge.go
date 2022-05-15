package channel

import (
	"context"
)

// Merge multiple channels into a single one.
func Merge[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	res := make(chan T)

	go func(res chan<- T) {
		defer close(res)

		for _, ch := range channels {
			for msg := range ch {
				res <- msg
			}
		}
	}(res)

	return res
}
