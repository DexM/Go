package async

func Execute[T any](f func() (T, error)) func() (T, error) {
	return f
}
