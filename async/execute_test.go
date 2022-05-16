package async_test

import (
	"fmt"

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
