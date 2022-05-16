package async

type executeChMessageType[T any] struct {
	res T
	err error
}

func Execute[T any](f func() (T, error)) func() (T, error) {
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
