package async

import (
	"errors"
)

// Predefined errors.
var (
	ErrPromiseAlreadyExecuted = errors.New("promise was already executed, calling promise multiple times is not supported")
	ErrGroupAlreadyExecuted   = errors.New("group was already executed, calling Execute() on the same group multiple times is not supported")
)
