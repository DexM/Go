package channel

import (
	"context"
	"sync"
)

// Merge multiple channels into a single one.
func Merge[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	wg.Add(len(channels))

	res := make(chan T)
	go func() {
		wg.Wait()
		close(res)
	}()

	for _, ch := range channels {
		go func(ch <-chan T) {
			defer wg.Done()

			for msg := range ch {
				res <- msg
			}
		}(ch)
	}

	return res
}
