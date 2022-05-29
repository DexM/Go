package channel

import (
	"context"
)

// Consume channel concurrently.
// Concurrency must be greater than 0, but it makes no sense to have it less than 2.
// You must close input channel for error channel to be closed.
func Consume[T any](
	ctx context.Context,
	concurrency int,
	channel <-chan T,
	f func(context.Context, T) error,
) <-chan error {
	chErr := make(chan error)

	go func(chErr chan<- error) {
		defer close(chErr)

		for message := range channel {
			if err := f(ctx, message); err != nil {
				chErr <- err
			}
		}
	}(chErr)

	return chErr
}
