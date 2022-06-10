package async

type (
	// Promise will return result of an asynchronous execution.
	Promise[T any] func() T

	// PromiseWithError will return result of an asynchronous execution or an encountered error.
	PromiseWithError[T any] func() (T, error)
)
