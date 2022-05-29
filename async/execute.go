package async

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
		defer close(ch)

		res, err := f()
		ch <- executeChMessageType[T]{res, err}
	}()

	return func() (T, error) {
		msg := <-ch
		return msg.res, msg.err
	}
}
