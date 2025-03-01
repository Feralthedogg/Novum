// pkg/effect/effect.go

package effect

import "fmt"

type Effect interface {
	Handle() error
}

type LogEffect struct {
	Message string
}

func (le LogEffect) Handle() error {
	fmt.Println(le.Message)
	return nil
}

func Perform(e Effect) error {
	return e.Handle()
}
