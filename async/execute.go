package async

type executeChMessageType[T any] struct {
	res T
	err error
}

// Execute function f asynchronously.
// Returns promise which can be called to retrieve function's f result.
// Calling promise will block execution until function f returns.
// It is advisable to use context to cancel function's f execution (see example code).
func Execute[T any](f func() (T, error)) func() (T, error) {
	// This channel is buffered. It will be written to only once.
	// That way when function f completes, goroutine will end as well (even if promise is never called and channel not drained).
	ch := make(chan executeChMessageType[T], 1)

	go func() {
		defer close(ch)

		res, err := f()
		if err != nil {
			ch <- executeChMessageType[T]{err: err}
			return
		}

		ch <- executeChMessageType[T]{res: res}
	}()

	return func() (T, error) {
		msg := <-ch
		return msg.res, msg.err
	}
}
