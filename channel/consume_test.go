package channel_test

import (
	"context"
	"fmt"

	"dexm.lol/channel"
)

func ExampleConsume() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	type input struct {
		message    string
		shouldFail bool
	}

	chIn := make(chan input, 2)
	chIn <- input{message: "message 1", shouldFail: false}
	chIn <- input{message: "message 2", shouldFail: true}
	close(chIn)

	chErr := channel.Consume(ctx, 2, chIn, func(ctx context.Context, in input) error {
		if in.shouldFail {
			return fmt.Errorf("error consuming message: %s", in.message)
		}

		fmt.Println("Consumed message:", in.message)
		return nil
	})

	for err := range chErr {
		fmt.Println("Error received:", err.Error())
	}

	// Unordered output:
	// Consumed message: message 1
	// Error received: error consuming message: message 2
}
