// pkg/effect/effect.go

package effect

import "fmt"

// EffectFunc represents a side effect function.
type EffectFunc func() error

// NewLogEffect returns an EffectFunc that logs the provided message.
func NewLogEffect(message string) EffectFunc {
	return func() error {
		fmt.Println(message)
		return nil
	}
}

// Perform executes the provided side effect.
func Perform(e EffectFunc) error {
	return e()
}
