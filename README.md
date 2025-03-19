# Novum

Novum is a zero‑cost abstraction framework for Go that integrates state management, effect handling, contract checking, dependency injection, and advanced composite patterns into a single, chainable API. It is designed to bring functional programming patterns (such as pattern matching, asynchronous operations, and parallel composition) into the Go ecosystem without compromising on performance.

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [State Management](#state-management)
4. [Effect Handling](#effect-handling)
5. [Module System](#module-system)
6. [Contract Checking](#contract-checking)
7. [Composite](#composite)
   - [Future Integration](#future-integration)
   - [Parallel Composite Integration](#parallel-composite-integration)
8. [Pattern Matching](#pattern-matching)
9. [License](#license)

---

## Installation

Novum is managed as a Go module. To install, run:

```bash
go get github.com/Feralthedogg/Novum
```

---

## Quick Start

The following example demonstrates Novum’s key features using composite chains. It shows a synchronous chain, an asynchronous Future chain, and a parallel composite chain integrated with the module container.

```go
package main

import (
	"fmt"
	"time"

	"github.com/Feralthedogg/Novum/pkg/composite"
	"github.com/Feralthedogg/Novum/pkg/effect"
	"github.com/Feralthedogg/Novum/pkg/future"
	"github.com/Feralthedogg/Novum/pkg/module"
	"github.com/Feralthedogg/Novum/pkg/state"
)

// NetworkModule defines a simple network fetching interface.
type NetworkModule interface {
	Fetch(url string) (string, error)
}

// DefaultNetworkModule is a basic implementation of NetworkModule.
type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// If the Future returns an error, the composite fails its contract and logs the error.
func FromFuture[T any, Deps any](f future.Future[T], deps Deps) composite.NovumComposite[T, Deps] {
	return composite.Return[T, Deps](*new(T), deps).Bind(func(_ T, deps Deps) composite.NovumComposite[T, Deps] {
		res, err := f.Await()
		if err != nil {
			return composite.Return[T, Deps](*new(T), deps).
				WithContract(func(val T) bool { return false }).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future error: %v", err)))
		}
		return composite.Return[T, Deps](res, deps)
	})
}

func main() {
	// Create a dependency container with the network module.
	modContainer := module.NewContainer()
	modContainer.Register("network", DefaultNetworkModule{})

	// --- Synchronous Composite Example ---
	syncComp := composite.Return(10, modContainer).
		WithContract(func(n int) bool { return n >= 0 }).
		Bind(func(n int, deps *module.Container) composite.NovumComposite[int, *module.Container] {
			newValue := n + 10
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect("Added 10 to the value"))
		}).
		Bind(func(n int, deps *module.Container) composite.NovumComposite[int, *module.Container] {
			newValue := n * 2
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect("Multiplied the value by 2"))
		}).
		Bind(func(n int, deps *module.Container) composite.NovumComposite[int, *module.Container] {
			mod, ok := deps.Resolve("network")
			if !ok {
				return composite.Return(n, deps).
					WithEffect(effect.NewLogEffect("Network module not found"))
			}
			network, ok := mod.(NetworkModule)
			if !ok {
				return composite.Return(n, deps).
					WithEffect(effect.NewLogEffect("Invalid network module type"))
			}
			data, err := network.Fetch("https://api.example.com/data")
			if err != nil {
				return composite.Return(n, deps).
					WithEffect(effect.NewLogEffect("Error fetching data from network"))
			}
			return composite.Return(n, deps).
				WithEffect(effect.NewLogEffect("Fetched data: " + data))
		})
	initialState := state.NewStateLayer[int](0)
	finalValue, finalState, effects, err := syncComp.Run(initialState)
	if err != nil {
		fmt.Println("Synchronous Composite Error:", err)
	} else {
		fmt.Printf("Synchronous Composite - Final Value: %d\n", finalValue)
		fmt.Printf("Synchronous Composite - Final State: %+v\n", finalState)
		fmt.Println("Synchronous Composite - Executing Effects:")
		for _, eff := range effects {
			_ = eff()
		}
	}

	// --- Future Composite Example ---
	fut := future.NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})
	futureComp := FromFuture(fut, modContainer).
		Bind(func(n int, deps *module.Container) composite.NovumComposite[int, *module.Container] {
			newValue := n * 3
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future result multiplied by 3: %d", newValue)))
		})
	futureResult, _, _, err := futureComp.Run(initialState)
	if err != nil {
		fmt.Println("Future Composite Error:", err)
	} else {
		fmt.Printf("Future Composite - Final Result: %d\n", futureResult)
	}

	// --- Parallel Composite Example ---
	comps := []composite.NovumComposite[int, *module.Container]{
		composite.Return(1, modContainer).WithEffect(effect.NewLogEffect("Parallel composite 1")),
		composite.Return(2, modContainer).WithEffect(effect.NewLogEffect("Parallel composite 2")),
		composite.Return(3, modContainer).WithEffect(effect.NewLogEffect("Parallel composite 3")),
	}
	parallelComp := composite.Parallel(comps)
	parallelResult, _, _, err := parallelComp.Run(state.NewStateLayer[int](0))
	if err != nil {
		fmt.Println("Parallel Composite Error:", err)
	} else {
		fmt.Printf("Parallel Composite - Final Results: %+v\n", parallelResult)
	}
}
```

---

## State Management

### Overview

- **Immutable State Transitions:**  
  Instead of modifying state in place, Novum creates new state objects, ensuring consistency and safe concurrency.
- **Context-Based Management:**  
  Each request or session maintains its own state, allowing predictable behavior in concurrent environments.

### API

- `NewStateLayer(initial int) StateLayer[int]`  
  Creates a new state object with the specified initial value.
- `(s StateLayer[int]) Increment() StateLayer[int]`  
  Returns a new state object with the counter incremented.

---

## Effect Handling

### Overview

- **Explicit Side Effect Separation:**  
  Side effects such as logging, network I/O, or file operations are encapsulated via the `EffectFunc` type, keeping pure functions free of side effects.
- **Optimized Execution:**  
  The side effect functions are simple and can be efficiently executed and inlined by the compiler.

### API

- **`EffectFunc` Type:**  
  A function type representing a side effect.
- **`NewLogEffect` Function:**  
  Returns an `EffectFunc` that logs a given message.
- `Perform(e EffectFunc) error`  
  Executes the provided side effect.

---

## Module System

### Overview

- **Static Dependency Injection:**  
  Novum uses a module container that can store and resolve dependencies by key.
- **Flexible and Type-Safe:**  
  The container supports registering multiple modules and resolving them in a type-safe manner.

### API

- `NewContainer() *Container`  
  Creates a new module container.
- `Register(key string, module interface{})`  
  Registers a module under a given key.
- `Resolve(key string) (interface{}, bool)`  
  Retrieves the module associated with a key.
- `ResolveAs[T any](key string) (T, bool)`  
  Retrieves the module as type T in a type-safe manner.

---

## Contract Checking

### Overview

- **Design by Contract:**  
  Novum includes a lightweight mechanism for asserting preconditions and postconditions. If conditions are not met, errors are raised.
- **Optional Enforcement:**  
  Contracts can be enabled during development for additional safety and selectively disabled in production.

### API

- `Assert(condition bool, msg string)`  
  Triggers a runtime error with the provided message if the condition is false.

---

## Composite

### Overview

Novum Composite integrates state transitions, effect handling, and contract checks into a single chainable abstraction. This enables you to build complex operations as a sequence of expressions.

### API

- `Return[T any, Deps any](value T, deps Deps) NovumComposite[T, Deps]`  
  Wraps a value and its dependencies into a composite.
- `Bind(func(T, Deps) NovumComposite[T, Deps]) NovumComposite[T, Deps]`  
  Chains operations together.
- `WithEffect(e EffectFunc) NovumComposite[T, Deps]`  
  Adds a side effect to the chain.
- `WithContract(func(T) bool) NovumComposite[T, Deps]`  
  Sets a contract for validation.
- `Run(initialState StateLayer[int]) (T, StateLayer[int], []EffectFunc, error)`  
  Executes the chain and returns the final value, state, side effects, and error (if any).

### Future Integration

To integrate asynchronous operations, use `FromFuture` to convert a `future.Future[T]` into a composite chain:

```go
// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// If the Future returns an error, the composite fails its contract and logs the error.
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
```

### Parallel Composite Integration

The `Parallel` function runs multiple composite chains concurrently and aggregates their results:

```go
// Parallel takes a slice of NovumComposite[T, Deps] and returns a composite whose value is a slice of T.
// It runs all composites concurrently and aggregates their results. If any composite returns an error,
// the first error is logged.
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
```

---

## Pattern Matching

Novum also includes a pure functional pattern matching module inspired by Elixir. It supports basic patterns:  
- **Literal:** Matches if a value equals a given literal.  
- **Var:** Matches any value and binds it to a variable.  
- **Pin:** Matches only if a variable is already bound and equals the given value.  
- **Cons:** Decomposes a list into head and tail.  
- **Wildcard:** Matches any value without binding.

```go
// pkg/patternmatch/patternmatch.go
package patternmatch

import (
	"fmt"
	"reflect"
)

// Env is a persistent environment for variable bindings.
type Env struct {
	parent   *Env
	bindings map[string]interface{}
}

// NewEnv creates a new empty environment.
func NewEnv() *Env {
	return &Env{
		parent:   nil,
		bindings: make(map[string]interface{}),
	}
}

// Extend creates a new environment node with an additional binding.
func (e *Env) Extend(key string, value interface{}) *Env {
	return &Env{
		parent:   e,
		bindings: map[string]interface{}{key: value},
	}
}

// Lookup searches for a binding in the environment chain.
func (e *Env) Lookup(key string) (interface{}, bool) {
	if val, ok := e.bindings[key]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Lookup(key)
	}
	return nil, false
}

// Pattern is the interface for pattern matching.
type Pattern interface {
	Match(value interface{}, env *Env) (bool, *Env, error)
}

// Literal pattern matches if the value equals the literal.
type Literal[T comparable] struct {
	Value T
}

func (l Literal[T]) Match(value interface{}, env *Env) (bool, *Env, error) {
	v, ok := value.(T)
	if !ok {
		return false, env, fmt.Errorf("Literal: type mismatch, expected %T", l.Value)
	}
	if v == l.Value {
		return true, env, nil
	}
	return false, env, fmt.Errorf("Literal: %v does not equal %v", v, l.Value)
}

// Var pattern matches any value and binds it to a variable.
type Var struct {
	Name string
}

func (v Var) Match(value interface{}, env *Env) (bool, *Env, error) {
	if oldVal, ok := env.Lookup(v.Name); ok {
		if reflect.DeepEqual(oldVal, value) {
			return true, env, nil
		}
		return false, env, fmt.Errorf("Var: variable '%s' already bound to %v, cannot rebind to %v", v.Name, oldVal, value)
	}
	return true, env.Extend(v.Name, value), nil
}

// Pin pattern matches only if the variable is already bound and equals the given value.
type Pin struct {
	Name string
}

func (p Pin) Match(value interface{}, env *Env) (bool, *Env, error) {
	oldVal, ok := env.Lookup(p.Name)
	if !ok {
		return false, env, fmt.Errorf("Pin: variable '%s' not bound", p.Name)
	}
	if reflect.DeepEqual(oldVal, value) {
		return true, env, nil
	}
	return false, env, fmt.Errorf("Pin: variable '%s' bound to %v does not match %v", p.Name, oldVal, value)
}

// Cons pattern matches a non-empty list by decomposing it into head and tail.
type Cons struct {
	Head Pattern
	Tail Pattern
}

func (c Cons) Match(value interface{}, env *Env) (bool, *Env, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return false, env, fmt.Errorf("Cons: value is not a slice or array")
	}
	if rv.Len() == 0 {
		return false, env, fmt.Errorf("Cons: empty list cannot match")
	}
	ok, newEnv, err := c.Head.Match(rv.Index(0).Interface(), env)
	if !ok {
		return false, env, fmt.Errorf("Cons: head match failed: %w", err)
	}
	tail := make([]interface{}, 0, rv.Len()-1)
	for i := 1; i < rv.Len(); i++ {
		tail = append(tail, rv.Index(i).Interface())
	}
	return c.Tail.Match(tail, newEnv)
}

// Wildcard matches any value without binding.
type Wildcard struct{}

func (Wildcard) Match(value interface{}, env *Env) (bool, *Env, error) {
	return true, env, nil
}
```

### Usage Example for Pattern Matching

```go
package main

import (
	"fmt"

	"github.com/Feralthedogg/Novum/pkg/patternmatch"
)

func main() {
	// Literal pattern matching.
	lit := patternmatch.Literal[int]{Value: 42}
	ok, env, err := lit.Match(42, patternmatch.NewEnv())
	fmt.Printf("Literal match: %v, Env: %+v, Error: %v\n", ok, env, err)

	// Variable pattern matching.
	varPat := patternmatch.Var{Name: "x"}
	ok, env, err = varPat.Match(100, patternmatch.NewEnv())
	fmt.Printf("Variable match: %v, Env: %+v, Error: %v\n", ok, env, err)

	// Pin pattern matching.
	baseEnv := patternmatch.NewEnv()
	baseEnv = baseEnv.Extend("y", 200)
	pinPat := patternmatch.Pin{Name: "y"}
	ok, env, err = pinPat.Match(200, baseEnv)
	fmt.Printf("Pin match: %v, Env: %+v, Error: %v\n", ok, env, err)

	// Cons pattern matching.
	consPat := patternmatch.Cons{
		Head: patternmatch.Var{Name: "head"},
		Tail: patternmatch.Wildcard{},
	}
	ok, env, err = consPat.Match([]interface{}{1, 2, 3}, patternmatch.NewEnv())
	fmt.Printf("Cons match: %v, Env: %+v, Error: %v\n", ok, env, err)
}
```

---

## License

Novum is licensed under the [MIT License](LICENSE).