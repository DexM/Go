package async

import (
	"errors"
	"fmt"
)

type (
	// AggregatedError is a collection of several individual errors.
	AggregatedError []error
)

func (e AggregatedError) Error() string {
	return fmt.Sprintf("error aggregated from %d errors", len(e))
}

// Has reports whether any error in collection matches target.
//
// See errors.Is() documentation for additional information:
// https://pkg.go.dev/errors#Is
func (e AggregatedError) Has(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// Find the first error in collection that matches target, and if one is found, sets target to that error value and returns true.
// Otherwise, it returns false.
//
// See errors.As() documentation for additional information:
// https://pkg.go.dev/errors#As
func (e AggregatedError) Find(target any) bool {
	for _, err := range e {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}
