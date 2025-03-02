// pkg/effect/effect.go

package effect

import "fmt"

type EffectFunc func() error

func NewLogEffect(message string) EffectFunc {
	return func() error {
		fmt.Println(message)
		return nil
	}
}

func Perform(e EffectFunc) error {
	return e()
}
