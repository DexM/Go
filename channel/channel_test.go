package channel_test

import (
	"context"
	"fmt"

	"dexm.lol/channel"
)

func Example() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	type processInput struct {
		message           string
		shouldFailProcess bool
		shouldFailConsume bool
	}

	type processOutput struct {
		message           string
		shouldFailConsume bool
	}

	chIn := make(chan processInput, 3)
	chIn <- processInput{message: "message 1", shouldFailProcess: false, shouldFailConsume: false}
	chIn <- processInput{message: "message 2", shouldFailProcess: false, shouldFailConsume: true}
	chIn <- processInput{message: "message 3", shouldFailProcess: true, shouldFailConsume: true}
	close(chIn)

	chProcessRes, chProcessErr := channel.Process(ctx, 2, chIn, func(ctx context.Context, in processInput) (processOutput, error) {
		if in.shouldFailProcess {
			return processOutput{}, fmt.Errorf("error processing message: %s", in.message)
		}

		res := processOutput{
			message:           fmt.Sprintf("processed message: %s", in.message),
			shouldFailConsume: in.shouldFailConsume,
		}
		return res, nil
	})

	chConsumeErr := channel.Consume(ctx, 2, chProcessRes, func(ctx context.Context, in processOutput) error {
		if in.shouldFailConsume {
			return fmt.Errorf("error consuming message: %s", in.message)
		}

		fmt.Println("Consumed message:", in.message)
		return nil
	})

	chErr := channel.Merge(ctx, chProcessErr, chConsumeErr)

	for err := range chErr {
		fmt.Println("Error received:", err.Error())
	}

	// Unordered output:
	// Consumed message: processed message: message 1
	// Error received: error consuming message: processed message: message 2
	// Error received: error processing message: message 3
}
