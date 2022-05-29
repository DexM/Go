package async

import (
	"fmt"
)

// Promise is a function returned by async.Execute function.
// It can be called later to retrieve result of asynchronous execution.
type Promise[T any] func() (T, error)

type executeChMessageType[T any] struct {
	res T
	err error
}

// Execute function f asynchronously.
// Returns promise which can be called to retrieve function's f result.
// Calling promise will block execution until function f returns.
// It is advisable to use context to cancel function's f execution (see example code).
func Execute[T any](f func() (T, error)) Promise[T] {
	// This channel is buffered. It will be written to only once.
	// That way when function f completes, goroutine will end as well (even if promise is never called and channel not drained).
	ch := make(chan executeChMessageType[T], 1)

	go func() {
		var msg executeChMessageType[T]

		defer func() {
			// Panics in goroutines must be handled.
			// Otherwise caller will receive nothing: no result and no error.
			if panicArg := recover(); panicArg != nil {
				if err, ok := panicArg.(error); ok {
					msg.err = fmt.Errorf("asynchronous function panicked: %w", err)
				} else {
					msg.err = fmt.Errorf("asynchronous function panicked: %v", panicArg)
				}
			}

			// After potential panic is handled, result can be sent to channel
			ch <- msg
			close(ch)
		}()

		res, err := f()

		msg.res = res
		msg.err = err
	}()

	return func() (T, error) {
		msg := <-ch
		return msg.res, msg.err
	}
}
