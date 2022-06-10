package async_test

import (
	"errors"
)

type (
	customErrorType1 struct {
		msg string
	}

	customErrorType2 struct {
		msg string
	}

	customErrorType3 struct {
		msg string
	}
)

var (
	dummyError = errors.New("dummy error")

	dummyError1 = errors.New("dummy error 1")
	dummyError2 = errors.New("dummy error 2")
	dummyError3 = errors.New("dummy error 3")
)

var (
	_ error = customErrorType1{}
	_ error = customErrorType2{}
	_ error = customErrorType3{}
)

func (e customErrorType1) Error() string {
	return e.msg
}

func (e customErrorType2) Error() string {
	return e.msg
}

func (e customErrorType3) Error() string {
	return e.msg
}
