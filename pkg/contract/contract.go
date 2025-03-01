package contract

import "fmt"

func Assert(condition bool, msg string) {
	if !condition {
		panic(fmt.Sprintf("Contract violation: %s", msg))
	}
}
