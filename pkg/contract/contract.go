// pkg/contract/contract.go

package contract

import "fmt"

// Assert triggers a runtime error with the given message if the condition is false.
func Assert(condition bool, msg string) {
	if !condition {
		panic(fmt.Sprintf("Contract violation: %s", msg))
	}
}
