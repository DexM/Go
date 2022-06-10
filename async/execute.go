package async

import (
	"fmt"
)

type executeChannelMessageType[T any] struct {
	res T
	err error
}

// Execute function f asynchronously.
// Returns promise which can be called to retrieve function's f result.
// Calling promise will block execution until function f returns result.
//
// Promise can be called only once.
// Calling promise repeatedly will result in an error.
//
// It is advisable to use context to cancel function's f execution (see example code).
func Execute[T any](f func() (T, error)) PromiseWithError[T] {
	// This channel is buffered. It will be written to only once.
	// That way when function f completes, goroutine will end as well (even if promise is never called and channel not drained).
	ch := make(chan executeChannelMessageType[T], 1)

	go func() {
		// Make sure channel is always closed when asynchronous function completes.
		defer close(ch)

		// Make sure result message is always sent to a promise.
		var msg executeChannelMessageType[T]
		defer func() { ch <- msg }()

		// Make sure panics are handles.
		// Otherwise caller will receive nothing - neither result, nor error.
		defer func() {
			if panicArg := recover(); panicArg != nil {
				if err, ok := panicArg.(error); ok {
					msg.err = fmt.Errorf("asynchronous function panicked: %w", err)
				} else {
					msg.err = fmt.Errorf("asynchronous function panicked: %v", panicArg)
				}
			}
		}()

		// Execute function f and store the result and error.
		res, err := f()

		msg.res = res
		msg.err = err
	}()

	return func() (T, error) {
		msg, ok := <-ch
		if !ok {
			msg.err = ErrPromiseAlreadyExecuted
		}

		return msg.res, msg.err
	}
}
