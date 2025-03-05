// pkg/state/state.go

package state

import "golang.org/x/exp/constraints"

// StateLayer represents an immutable state with a numeric counter.
type StateLayer[T constraints.Integer] struct {
	Counter T
}

// NewStateLayer creates a new state with the given initial counter value.
func NewStateLayer[T constraints.Integer](initial T) StateLayer[T] {
	return StateLayer[T]{Counter: initial}
}

// Increment returns a new state with the counter incremented.
func (s StateLayer[T]) Increment() StateLayer[T] {
	return StateLayer[T]{Counter: s.Counter + 1}
}
