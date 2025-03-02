// pkg/composite/composite.go

package composite

import (
	"errors"
	"fmt"

	"github.com/Feralthedogg/Novum/pkg/effect"
	st "github.com/Feralthedogg/Novum/pkg/state"
)

// NovumComposite chains computations with state transitions, side effects, and contract checks.
// The Deps generic parameter includes compile‑time dependencies (including modules).
type NovumComposite[T any, Deps any] struct {
	value      T
	stateFn    func(st.StateLayer) st.StateLayer
	effects    []effect.EffectFunc
	contractFn func(T) bool
	err        error
	deps       Deps
}

// Return wraps a value and dependencies into a NovumComposite.
func Return[T any, Deps any](value T, deps Deps) NovumComposite[T, Deps] {
	return NovumComposite[T, Deps]{
		value: value,
		stateFn: func(s st.StateLayer) st.StateLayer {
			return s
		},
		effects:    nil,
		contractFn: func(val T) bool { return true },
		err:        nil,
		deps:       deps,
	}
}

// Bind chains the current composite with function f and combines state transitions and effects.
func (m NovumComposite[T, Deps]) Bind(f func(T, Deps) NovumComposite[T, Deps]) NovumComposite[T, Deps] {
	if m.err != nil {
		return m
	}
	if !m.contractFn(m.value) {
		m.err = errors.New("contract violation before Bind: invalid value")
		return m
	}
	next := f(m.value, m.deps)
	if next.err != nil {
		m.err = fmt.Errorf("error in Bind: %w", next.err)
		return m
	}
	combinedStateFn := func(s st.StateLayer) st.StateLayer {
		return next.stateFn(m.stateFn(s))
	}
	combinedEffects := append(m.effects, next.effects...)
	return NovumComposite[T, Deps]{
		value:      next.value,
		stateFn:    combinedStateFn,
		effects:    combinedEffects,
		contractFn: next.contractFn,
		err:        m.err,
		deps:       m.deps,
	}
}

// WithEffect appends a side effect.
func (m NovumComposite[T, Deps]) WithEffect(e effect.EffectFunc) NovumComposite[T, Deps] {
	m.effects = append(m.effects, e)
	return m
}

// WithContract sets a new contract function.
func (m NovumComposite[T, Deps]) WithContract(fn func(T) bool) NovumComposite[T, Deps] {
	m.contractFn = fn
	return m
}

// Run executes the composite chain with an initial state and returns the final value,
// state, accumulated side effects, and any error.
func (m NovumComposite[T, Deps]) Run(initialState st.StateLayer) (T, st.StateLayer, []effect.EffectFunc, error) {
	finalState := m.stateFn(initialState)
	if !m.contractFn(m.value) {
		return m.value, finalState, m.effects, errors.New("final contract violation")
	}
	return m.value, finalState, m.effects, m.err
}
