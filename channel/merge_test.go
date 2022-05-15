package channel_test

import (
	"context"
	"fmt"

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
