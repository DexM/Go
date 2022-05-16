package async_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"dexm.lol/async"
)

func ExampleExecute() {
	promise := async.Execute(func() (string, error) {
		return "string result of some lengthy operation", nil
	})

	res, err := promise()
	fmt.Println("Result:", res)
	fmt.Println("Error:", err)

	// Output:
	// Result: string result of some lengthy operation
	// Error: <nil>
}

func TestExecutePassesResultToPromise(t *testing.T) {
	promise := async.Execute(func() (string, error) {
		return "dummy result", nil
	})

	res, err := promise()
	if err != nil {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if res != "dummy result" {
		t.Errorf("Unexpected result received from the promise: %s", res)
	}
}

func TestExecutePassesErrorToPromise(t *testing.T) {
	promise := async.Execute(func() (interface{}, error) {
		return nil, errors.New("dummy error")
	})

	res, err := promise()
	if err == nil {
		t.Error("Expected to receive error from the promise")
	}
	if err.Error() != "dummy error" {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if res != nil {
		t.Errorf("Unexpected result received from the promise: %#v", res)
	}
}

func TestExecuteLaunchesAsync(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	promise := async.Execute(func() (string, error) {
		defer wg.Done()
		return "dummy result", nil
	})

	wg.Wait()

	res, err := promise()
	if err != nil {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if res != "dummy result" {
		t.Errorf("Unexpected result received from the promise: %s", res)
	}
}
