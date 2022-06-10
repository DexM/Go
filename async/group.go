package async

import (
	"fmt"
	"sync"
)

// Group functions for asynchronous execution.
// Method Execute() will block until all functions have executed and will return aggregated errors.
// Each function added to the group will have its own promise to return successful result.
type Group struct {
	funcs    []func(*sync.WaitGroup, chan<- error)
	executed bool
}

// Execute functions added to the group.
// Will block until all functions have executed and will return aggregated errors.
//
// Execute can be called only once.
// Calling Execute() repeatedly will result in an error.
func (g *Group) Execute() (errs AggregatedError) {
	// Check whether this group was already executed.
	if g.executed {
		return AggregatedError{ErrGroupAlreadyExecuted}
	}

	// Mark group as executed.
	g.executed = true

	// Wait group to track when all the functions are complete
	var wg sync.WaitGroup
	wg.Add(len(g.funcs))

	// Channel to store all errors.
	// Channel size is preallocated, so all function can finish execution without being blocked.
	errCh := make(chan error, len(g.funcs))

	// Asynchronous function which will close error channel when all the functions have finished.
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Launch all functions.
	for _, f := range g.funcs {
		go f(&wg, errCh)
	}

	// Collect and return errors.
	for err := range errCh {
		errs = append(errs, err)
	}

	return
}

// AddToExecutionGroup registers function f with the execution group.
// Returns promise which will return successful result.
//
// Promise can be called only once.
// Calling promise repeatedly will result in panic.
func AddToExecutionGroup[T any](group *Group, f func() (T, error)) Promise[T] {
	// This channel is buffered. It will be written to only once.
	// That way when function f completes, goroutine will end as well (even if promise is never called and channel not drained).
	resCh := make(chan T, 1)

	group.funcs = append(group.funcs, func(wg *sync.WaitGroup, errCh chan<- error) {
		// Make sure wait group is notified about completion of this function.
		defer wg.Done()

		// Make sure channel is always closed when asynchronous function completes.
		defer close(resCh)

		// Make sure result message is always sent to a promise.
		var resData T
		defer func() { resCh <- resData }()

		// Make sure error is sent to ExecutionGroup if present.
		var resErr error
		defer func() {
			if resErr != nil {
				errCh <- resErr
			}
		}()

		// Make sure panics are handles.
		// Otherwise caller will receive nothing - neither result, nor error.
		defer func() {
			if panicArg := recover(); panicArg != nil {
				if err, ok := panicArg.(error); ok {
					resErr = fmt.Errorf("asynchronous function panicked: %w", err)
				} else {
					resErr = fmt.Errorf("asynchronous function panicked: %v", panicArg)
				}
			}
		}()

		// Execute function f and store the result and error.
		resData, resErr = f()
	})

	return func() T {
		msg, ok := <-resCh
		if !ok {
			panic(ErrPromiseAlreadyExecuted)
		}
		return msg
	}
}
