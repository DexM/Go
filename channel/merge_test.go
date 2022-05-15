package channel_test

import (
	"context"
	"fmt"
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
