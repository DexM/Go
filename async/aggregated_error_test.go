package async_test

import (
	"errors"
	"fmt"
	"testing"

	"dexm.lol/async"
)

// Ensure interface implementation
var (
	_ error = async.AggregatedError{}
)

func TestAggregatedError_beingWrapped(t *testing.T) {
	err := fmt.Errorf("wrapping error: %w", async.AggregatedError{
		dummyError1,
		dummyError2,
		dummyError3,
	})

	if err.Error() != "wrapping error: error aggregated from 3 errors" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	var aggregatedError async.AggregatedError
	if !errors.As(err, &aggregatedError) {
		t.Errorf("errors.As() did not detect wrapped async.AggregatedError")
	}

	if aggregatedError.Error() != "error aggregated from 3 errors" {
		t.Errorf("aggregated error was not unwrapped properly: %#v", aggregatedError)
	}
}

func TestAggregatedError_Has(t *testing.T) {
	e := async.AggregatedError{
		dummyError1,
		fmt.Errorf("wrapping error: %w", dummyError2),
	}

	if !e.Has(dummyError1) {
		t.Error("Expected to find error in aggregated error")
	}

	if !e.Has(dummyError2) {
		t.Error("Expected to find error in aggregated error")
	}

	if e.Has(dummyError3) {
		t.Error("Did not expected to find error in aggregated error")
	}
}

func TestAggregatedError_Find(t *testing.T) {
	e := async.AggregatedError{
		customErrorType1{"custom error 1"},
		fmt.Errorf("wrapping error: %w", customErrorType2{"custom error 2"}),
	}

	var customError1 customErrorType1
	if !e.Find(&customError1) {
		t.Fatal("Expected to find error in aggregated error")
	}
	if customError1.msg != "custom error 1" {
		t.Errorf("Unexpected error received: %#v", customError1)
	}

	var customError2 customErrorType2
	if !e.Find(&customError2) {
		t.Fatal("Expected to find error in aggregated error")
	}
	if customError2.msg != "custom error 2" {
		t.Errorf("Unexpected error received: %#v", customError2)
	}

	var customError3 customErrorType3
	if e.Find(&customError3) {
		t.Error("Did not expected to find error in aggregated error")
	}
}
