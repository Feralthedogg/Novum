// pkg/composite/composite.go
package composite

import (
	"errors"
	"fmt"
	"sync"

	// For Numeric constraint
	"github.com/Feralthedogg/Novum/pkg/effect"
	"github.com/Feralthedogg/Novum/pkg/future"
	st "github.com/Feralthedogg/Novum/pkg/state"
)

// Numeric is a type constraint for numeric types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// Add is a type-safe function that adds two numeric values.
func Add[T Numeric](a, b T) T {
	return a + b
}

// Multiply is a type-safe function that multiplies two numeric values.
func Multiply[T Numeric](a, b T) T {
	return a * b
}

// NovumComposite chains computations with state transitions, side effects, and contract checks.
// T represents the computed value type, and Deps represents compile‑time dependencies.
type NovumComposite[T any, Deps any] struct {
	value      T
	stateFn    func(st.StateLayer[int]) st.StateLayer[int]
	effects    []effect.EffectFunc
	contractFn func(T) bool
	err        error
	deps       Deps
}

// Return wraps a value and dependencies into a NovumComposite.
func Return[T any, Deps any](value T, deps Deps) NovumComposite[T, Deps] {
	return NovumComposite[T, Deps]{
		value: value,
		stateFn: func(s st.StateLayer[int]) st.StateLayer[int] {
			return s
		},
		effects:    nil,
		contractFn: func(val T) bool { return true },
		err:        nil,
		deps:       deps,
	}
}

// Bind chains the current composite with function f, combining state transitions and effects.
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
	combinedStateFn := func(s st.StateLayer[int]) st.StateLayer[int] {
		return next.stateFn(m.stateFn(s))
	}
	combinedEffects := append(m.effects, next.effects...) // SA4010: append result is used in combinedEffects below.
	return NovumComposite[T, Deps]{
		value:      next.value,
		stateFn:    combinedStateFn,
		effects:    combinedEffects,
		contractFn: next.contractFn,
		err:        m.err,
		deps:       m.deps,
	}
}

// WithEffect appends a side effect to the composite chain.
func (m NovumComposite[T, Deps]) WithEffect(e effect.EffectFunc) NovumComposite[T, Deps] {
	m.effects = append(m.effects, e)
	return m
}

// WithContract sets a contract function to validate the computed value.
func (m NovumComposite[T, Deps]) WithContract(fn func(T) bool) NovumComposite[T, Deps] {
	m.contractFn = fn
	return m
}

// Run executes the composite chain with an initial state and returns the final value,
// the final state, accumulated side effects, and any error.
func (m NovumComposite[T, Deps]) Run(initialState st.StateLayer[int]) (T, st.StateLayer[int], []effect.EffectFunc, error) {
	finalState := m.stateFn(initialState)
	if !m.contractFn(m.value) {
		return m.value, finalState, m.effects, errors.New("final contract violation")
	}
	return m.value, finalState, m.effects, m.err
}

// ------------------------------
// Future Integration
// ------------------------------

// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// This function bridges asynchronous Future results with the composite chain.
// If the Future returns an error, a composite with a failing contract and error logging is returned.
func FromFuture[T any, Deps any](f future.Future[T], deps Deps) NovumComposite[T, Deps] {
	return Return[T, Deps](*new(T), deps).Bind(func(_ T, deps Deps) NovumComposite[T, Deps] {
		res, err := f.Await()
		if err != nil {
			return Return[T, Deps](*new(T), deps).
				WithContract(func(val T) bool { return false }).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future error: %v", err)))
		}
		return Return[T, Deps](res, deps)
	})
}

// ------------------------------
// Parallel Composite Integration
// ------------------------------

// Parallel takes a slice of NovumComposite[T, Deps] and returns a composite whose value is a slice of T.
// It runs all composites concurrently and aggregates their results.
// If any composite returns an error, the first error is logged.
func Parallel[T any, Deps any](comps []NovumComposite[T, Deps]) NovumComposite[[]T, Deps] {
	deps := comps[0].deps
	return Return[[]T, Deps](nil, deps).Bind(func(_ []T, deps Deps) NovumComposite[[]T, Deps] {
		n := len(comps)
		results := make([]T, n)
		errCh := make(chan error, n)
		effCh := make(chan []effect.EffectFunc, n)
		var wg sync.WaitGroup
		wg.Add(n)
		for i, comp := range comps {
			i, comp := i, comp // capture loop variables
			go func() {
				defer wg.Done()
				res, _, effs, err := comp.Run(st.NewStateLayer[int](0))
				errCh <- err
				results[i] = res
				effCh <- effs
			}()
		}
		wg.Wait()
		close(errCh)
		close(effCh)
		var firstErr error
		totalEffects := []effect.EffectFunc{}
		for err := range errCh {
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		for effs := range effCh {
			totalEffects = append(totalEffects, effs...)
		}
		if firstErr != nil {
			return Return[[]T, Deps](results, deps).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Parallel composite error: %v", firstErr)))
		}
		return Return[[]T, Deps](results, deps).
			WithEffect(effect.NewLogEffect("Parallel composite executed successfully"))
	}).WithContract(func(vals []T) bool {
		return vals != nil
	})
}
