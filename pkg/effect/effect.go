// pkg/effect/effect.go

package effect

import "fmt"

type EffectFunc func() error

func LogEffect(message string) EffectFunc {
	return func() error {
		fmt.Println(message)
		return nil
	}
}

func Perform(e EffectFunc) error {
	return e()
}
