package channel_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"dexm.lol/channel"
)

func ExampleMerge() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ch1 := make(chan string, 2)
	ch1 <- "message 1"
	ch1 <- "message 2"
	close(ch1)

	ch2 := make(chan string, 2)
	ch2 <- "message 3"
	ch2 <- "message 4"
	close(ch2)

	ch3 := channel.Merge(ctx, ch1, ch2)

	for message := range ch3 {
		fmt.Println("Received message:", message)
	}

	// Unordered output:
	// Received message: message 1
	// Received message: message 2
	// Received message: message 3
	// Received message: message 4
}

func TestMergeReadsChannelsConcurrently(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	chInput1 := make(chan string)
	chInput2 := make(chan string)

	go func() {
		defer close(chInput1)
		defer close(chInput2)

		chInput1 <- "message 1 from channel 1"
		chInput2 <- "message 1 from channel 2"

		chInput1 <- "message 2 from channel 1"
		chInput2 <- "message 2 from channel 2"
	}()

	chRes := channel.Merge(ctx, chInput1, chInput2)

	expected := map[string]bool{
		"message 1 from channel 1": true,
		"message 1 from channel 2": true,
		"message 2 from channel 1": true,
		"message 2 from channel 2": true,
	}

	actual := make(map[string]bool)
	for res := range chRes {
		actual[res] = true
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Error(diff)
	}
}

func TestMergeDoesNotLeakGoroutines(t *testing.T) {
	const dataSize = 1024 * 1024
	var memStatsBefore, memStatsAfter runtime.MemStats
	var memUsageDiff uint64

	// Disable parallelism for more predictable results.
	oldGoMaxProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldGoMaxProcs)

	// Take initial memory reading (for some reason running GC twice gives more stable results).
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&memStatsBefore)

	func() {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		chInput := make(chan []byte, 2)
		chInput <- make([]byte, dataSize)
		chInput <- make([]byte, dataSize)
		close(chInput)

		chRes := channel.Merge(ctx, chInput)
		<-chRes
		// Do not drain channel fully, leave 2nd message in the channel.
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
