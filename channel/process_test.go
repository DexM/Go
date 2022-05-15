package channel_test

import (
	"context"
	"fmt"

	"dexm.lol/channel"
)

func ExampleProcess() {
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

	chRes, chErr := channel.Process(ctx, 2, chIn, func(ctx context.Context, in input) (string, error) {
		if in.shouldFail {
			return "", fmt.Errorf("error processing message: %s", in.message)
		}
		return fmt.Sprintf("processed message: %s", in.message), nil
	})

	res := <-chRes
	fmt.Println("Result received:", res)

	err := <-chErr
	fmt.Println("Error received:", err.Error())

	// Unordered output:
	// Result received: processed message: message 1
	// Error received: error processing message: message 2
}
