// pkg/composite/composite.go

package composite

import (
	"errors"
	"fmt"

	"github.com/Feralthedogg/Novum/pkg/effect"
	"github.com/Feralthedogg/Novum/pkg/module"
	"github.com/Feralthedogg/Novum/pkg/state"
)

// NovumComposite is a generic structure that combines a value with state transitions, side effects,
// contract validation, error handling, and an optional module container.
type NovumComposite[T any] struct {
	value      T
	stateFn    func(state.StateLayer) state.StateLayer
	effects    []effect.Effect
	contractFn func(T) bool
	err        error
	modules    *module.Container
}

// Return wraps a value into a NovumComposite with default settings.
func Return[T any](value T) NovumComposite[T] {
	return NovumComposite[T]{
		value:      value,
		stateFn:    func(s state.StateLayer) state.StateLayer { return s },
		effects:    []effect.Effect{},
		contractFn: func(val T) bool { return true },
		err:        nil,
		modules:    nil,
	}
}

// Bind chains the current composite with function f and combines state functions and effects.
func (m NovumComposite[T]) Bind(f func(T) NovumComposite[T]) NovumComposite[T] {
	if m.err != nil {
		return m
	}
	if !m.contractFn(m.value) {
		m.err = errors.New("contract violation before Bind: invalid value")
		return m
	}
	next := f(m.value)
	if next.err != nil {
		m.err = fmt.Errorf("error in Bind: %w", next.err)
		return m
	}
	combinedStateFn := func(s state.StateLayer) state.StateLayer {
		intermediate := m.stateFn(s)
		return next.stateFn(intermediate)
	}
	combinedEffects := append(m.effects, next.effects...)
	var combinedModules *module.Container
	if next.modules != nil {
		combinedModules = next.modules
	} else {
		combinedModules = m.modules
	}
	return NovumComposite[T]{
		value:      next.value,
		stateFn:    combinedStateFn,
		effects:    combinedEffects,
		contractFn: next.contractFn,
		err:        m.err,
		modules:    combinedModules,
	}
}

// WithEffect appends a new side effect.
func (m NovumComposite[T]) WithEffect(e effect.Effect) NovumComposite[T] {
	m.effects = append(m.effects, e)
	return m
}

// WithContract sets a new contract function.
func (m NovumComposite[T]) WithContract(fn func(T) bool) NovumComposite[T] {
	m.contractFn = fn
	return m
}

// WithModule sets the module container.
func (m NovumComposite[T]) WithModule(container *module.Container) NovumComposite[T] {
	m.modules = container
	return m
}

// ResolveModule retrieves a module by name from the module container.
func (m NovumComposite[T]) ResolveModule(name string) (interface{}, bool) {
	if m.modules != nil {
		return m.modules.Resolve(name)
	}
	return nil, false
}

// Run executes the composite chain with an initial state and returns the final value, state,
// accumulated effects, module container, and error (error is returned last).
func (m NovumComposite[T]) Run(initialState state.StateLayer) (T, state.StateLayer, []effect.Effect, *module.Container, error) {
	finalState := m.stateFn(initialState)
	if !m.contractFn(m.value) {
		return m.value, finalState, m.effects, m.modules, errors.New("final contract violation")
	}
	return m.value, finalState, m.effects, m.modules, m.err
}
