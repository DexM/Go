package async

type (
	// PromiseWithError will return result of an asynchronous execution or an encountered error.
	PromiseWithError[T any] func() (T, error)
)
