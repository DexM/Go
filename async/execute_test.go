package async_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"testing"

	"dexm.lol/async"
)

func ExampleExecute() {
	promise := async.Execute(func() (string, error) {
		// Perform some lengthy operation.

		return "string result of some lengthy operation", nil
	})

	// Perform another lengthy operation.

	res, err := promise()
	fmt.Println("Result:", res)
	fmt.Println("Error:", err)

	// Output:
	// Result: string result of some lengthy operation
	// Error: <nil>
}

func ExampleExecute_waitGroups() {
	// Use wait groups to execute multiple asynchronous functions and read results only when all of them succeeded.
	var wg sync.WaitGroup
	wg.Add(2)

	promise1 := async.Execute(func() (string, error) {
		defer wg.Done()

		return "string result", nil
	})

	promise2 := async.Execute(func() (int, error) {
		defer wg.Done()

		return 42, nil
	})

	// Wait for both asynchronous functions to complete before getting results.
	wg.Wait()

	stringRes, err := promise1()
	fmt.Printf("Result of 1st asynchronous function: %s. Error: %v\n", stringRes, err)

	intRes, err := promise2()
	fmt.Printf("Result of 2nd asynchronous function: %d. Error: %v\n", intRes, err)

	// Output:
	// Result of 1st asynchronous function: string result. Error: <nil>
	// Result of 2nd asynchronous function: 42. Error: <nil>
}

func ExampleExecute_context() {
	// Cancel context when main function exits to signal asynchronous function to exit as well.
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	promise := async.Execute(func() (string, error) {
		// Use context to cancel asynchronous function if base function exits early.
		_, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/", nil)
		if err != nil {
			return "", err
		}

		return "request was successful", nil
	})

	_, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/", nil)
	if err != nil {
		// It is safe to exit this function early, context will abort asynchronous function.
		return
	}

	res, err := promise()
	fmt.Println("Result:", res)
	fmt.Println("Error:", err)

	// Output:
	// Result: request was successful
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

func TestExecutePassesResultAndErrorToPromiseAtTheSameTime(t *testing.T) {
	promise := async.Execute(func() (string, error) {
		return "dummy result", errors.New("dummy error")
	})

	res, err := promise()
	if err == nil {
		t.Error("Expected to receive error from the promise")
	}
	if err.Error() != "dummy error" {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if res != "dummy result" {
		t.Errorf("Unexpected result received from the promise: %s", res)
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

func TestExecuteHandlesPanics(t *testing.T) {
	promise := async.Execute(func() (interface{}, error) {
		panic("panic happened")
	})

	res, err := promise()
	if err == nil {
		t.Error("Expected to receive error from the promise")
	}
	if err.Error() != "asynchronous function panicked: panic happened" {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if res != nil {
		t.Errorf("Unexpected result received from the promise: %#v", res)
	}
}

func TestExecuteHandlesPanicsAndWrapsOriginalError(t *testing.T) {
	promise := async.Execute(func() (interface{}, error) {
		panic(errors.New("panic happened"))
	})

	res, err := promise()
	if err == nil {
		t.Error("Expected to receive error from the promise")
	}
	if err.Error() != "asynchronous function panicked: panic happened" {
		t.Errorf("Unexpected error received from the promise: %s", err.Error())
	}
	if originalErr := errors.Unwrap(err); originalErr == nil || originalErr.Error() != "panic happened" {
		t.Errorf("Original error was not properly wrapped: %#v", originalErr)
	}
	if res != nil {
		t.Errorf("Unexpected result received from the promise: %#v", res)
	}
}

func TestExecuteDoesNotLeakGoroutines(t *testing.T) {
	const dataSize = 1024 * 1024
	var memStatsBefore, memStatsAfter runtime.MemStats
	var memUsageDiff uint64
	var wg sync.WaitGroup

	wg.Add(1)

	// Disable parallelism for more predictable results.
	oldGoMaxProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldGoMaxProcs)

	// Take initial memory reading (for some reason running GC twice gives more stable results).
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&memStatsBefore)

	func() {
		// Simulate situation when promise is never called.
		_ = async.Execute(func() ([]byte, error) {
			defer wg.Done()
			return make([]byte, dataSize), nil
		})

		wg.Wait()
	}()

	// Take final memory reading (for some reason running GC twice gives more stable results).
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&memStatsAfter)

	if memStatsBefore.HeapInuse > memStatsAfter.HeapInuse {
		memUsageDiff = memStatsBefore.HeapInuse - memStatsAfter.HeapInuse
	} else {
		memUsageDiff = memStatsAfter.HeapInuse - memStatsBefore.HeapInuse
	}

	if memUsageDiff != 0 {
		t.Errorf("Memory usage diff is greater than 0: %d", memUsageDiff)
	}
}
