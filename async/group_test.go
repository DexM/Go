package async_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"dexm.lol/async"
)

func ExampleGroup() {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (string, error) {
		// Perform some lengthy operation.

		return "string result of some lengthy operation", nil
	})

	promise2 := async.AddToExecutionGroup(&group, func() (int, error) {
		// Perform another lengthy operation.

		return 42, nil // Int result of some lengthy operation
	})

	if err := group.Execute(); err != nil {
		for _, err := range err {
			fmt.Println("Error:", err)
		}
		return
	}

	fmt.Println("Promise 1 result:", promise1())
	fmt.Println("Promise 2 result:", promise2())

	// Output:
	// Promise 1 result: string result of some lengthy operation
	// Promise 2 result: 42
}

func TestGroup_passesResultsToPromises(t *testing.T) {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (string, error) {
		return "dummy result 1", nil
	})

	promise2 := async.AddToExecutionGroup(&group, func() (string, error) {
		return "dummy result 2", nil
	})

	if err := group.Execute(); err != nil {
		t.Errorf("Unexpected error received from the execution group: %#v", err)
	}

	if res := promise1(); res != "dummy result 1" {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != "dummy result 2" {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}

func TestGroup_aggregatesErrors(t *testing.T) {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		return nil, dummyError1
	})

	promise2 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		return nil, dummyError2
	})

	err := group.Execute()
	if err == nil {
		t.Fatal("Expected to receive error from the execution group")
	}
	if len(err) != 2 {
		t.Fatalf("Expected aggregated error to have 2 errors, but got: %#v", err)
	}
	if !err.Has(dummyError1) {
		t.Errorf("Expected aggregated error to contain 1st error: %#v", err)
	}
	if !err.Has(dummyError2) {
		t.Errorf("Expected aggregated error to contain 2nd error: %#v", err)
	}

	if res := promise1(); res != nil {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != nil {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}

func TestGroup_passesResultsToPromisesAndAggregatesErrorsAtTheSameTime(t *testing.T) {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (string, error) {
		return "dummy result 1", dummyError1
	})

	promise2 := async.AddToExecutionGroup(&group, func() (string, error) {
		return "dummy result 2", dummyError2
	})

	err := group.Execute()
	if err == nil {
		t.Fatal("Expected to receive error from the execution group")
	}
	if len(err) != 2 {
		t.Fatalf("Expected aggregated error to have 2 errors, but got: %#v", err)
	}
	if !err.Has(dummyError1) {
		t.Errorf("Expected aggregated error to contain 1st error: %#v", err)
	}
	if !err.Has(dummyError2) {
		t.Errorf("Expected aggregated error to contain 2nd error: %#v", err)
	}

	if res := promise1(); res != "dummy result 1" {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != "dummy result 2" {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}

func TestGroup_launchesAsync(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (string, error) {
		wg.Done()
		wg.Wait()
		return "dummy result 1", nil
	})

	promise2 := async.AddToExecutionGroup(&group, func() (string, error) {
		wg.Done()
		wg.Wait()
		return "dummy result 2", nil
	})

	if err := group.Execute(); err != nil {
		t.Errorf("Unexpected error received from the execution group: %#v", err)
	}

	if res := promise1(); res != "dummy result 1" {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != "dummy result 2" {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}

func TestGroup_handlesPanics(t *testing.T) {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		panic("panic 1 happened")
	})

	promise2 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		panic("panic 2 happened")
	})

	err := group.Execute()
	if err == nil {
		t.Fatal("Expected to receive error from the execution group")
	}
	if len(err) != 2 {
		t.Fatalf("Expected aggregated error to have 2 errors, but got: %#v", err)
	}
	if !strings.HasSuffix(err[0].Error(), ": panic 1 happened") && !strings.HasSuffix(err[1].Error(), ": panic 1 happened") {
		t.Errorf("Expected aggregated error to contain 1st error: %#v", err)
	}
	if !strings.HasSuffix(err[0].Error(), ": panic 2 happened") && !strings.HasSuffix(err[1].Error(), ": panic 2 happened") {
		t.Errorf("Expected aggregated error to contain 2nd error: %#v", err)
	}

	if res := promise1(); res != nil {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != nil {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}

func TestGroup_handlesPanicsAndWrapsOriginalError(t *testing.T) {
	var group async.Group

	promise1 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		panic(dummyError1)
	})

	promise2 := async.AddToExecutionGroup(&group, func() (interface{}, error) {
		panic(dummyError2)
	})

	err := group.Execute()
	if err == nil {
		t.Fatal("Expected to receive error from the execution group")
	}
	if len(err) != 2 {
		t.Fatalf("Expected aggregated error to have 2 errors, but got: %#v", err)
	}
	if !err.Has(dummyError1) {
		t.Errorf("Expected aggregated error to contain 1st error: %#v", err)
	}
	if !err.Has(dummyError2) {
		t.Errorf("Expected aggregated error to contain 2nd error: %#v", err)
	}

	if res := promise1(); res != nil {
		t.Errorf("Unexpected result received from the promise 1: %#v", res)
	}

	if res := promise2(); res != nil {
		t.Errorf("Unexpected result received from the promise 2: %#v", res)
	}
}
