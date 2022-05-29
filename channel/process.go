package channel

import (
	"context"
)

// Process channel concurrently.
// Concurrency must be greater than 0, but it makes no sense to have it less than 2.
// You must close input channel for output and error channels to be closed.
func Process[T, R any](
	ctx context.Context,
	concurrency int,
	channel <-chan T,
	f func(context.Context, T) (R, error),
) (<-chan R, <-chan error) {
	chRes := make(chan R)
	chErr := make(chan error)

	go func(chRes chan<- R, chErr chan<- error) {
		defer close(chRes)
		defer close(chErr)

		for message := range channel {
			if res, err := f(ctx, message); err != nil {
				chErr <- err
			} else {
				chRes <- res
			}
		}
	}(chRes, chErr)

	return chRes, chErr
}
