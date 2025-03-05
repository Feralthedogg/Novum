# Novum

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
8. [License](#license)

---

## Installation

Novum is managed as a Go module. To install, run:

```bash
go get github.com/Feralthedogg/Novum
```

---

## Quick Start

The following example demonstrates Novum's key features using composite chains. It shows both a synchronous chain and an asynchronous Future chain integrated with the Composite API.

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

type NetworkModule interface {
	Fetch(url string) (string, error)
}

type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// If the Future returns an error, it returns a composite with a failing contract and logs the error.
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
	deps := module.NewContainer[NetworkModule](DefaultNetworkModule{})

	// --- Synchronous Composite Example ---
	syncComp := composite.Return(10, deps).
		WithContract(func(n int) bool { return n >= 0 }).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n + 10
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect("Added 10 to the value"))
		}).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n * 2
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect("Multiplied the value by 2"))
		}).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			data, err := deps.GetNetwork().Fetch("https://api.example.com/data")
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
	// Create a Future that asynchronously returns 42 after 100ms.
	fut := future.NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})
	futureComp := FromFuture(fut, deps).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
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
	comps := []composite.NovumComposite[int, module.Container[NetworkModule]]{
		composite.Return(1, deps).WithEffect(effect.NewLogEffect("Composite 1")),
		composite.Return(2, deps).WithEffect(effect.NewLogEffect("Composite 2")),
		composite.Return(3, deps).WithEffect(effect.NewLogEffect("Composite 3")),
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
  Instead of modifying state in place, Novum creates new state objects to ensure consistency and safe concurrency.
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
  Side effects like logging, network I/O, or file operations are encapsulated via the `EffectFunc` type, separating pure functions from impure ones.
- **Optimized Execution:**  
  The side effect functions are simple and can be efficiently executed and inlined by the compiler.

### API

- **`EffectFunc` Type:**  
  A function type that represents a side effect.
- **`NewLogEffect` Function:**  
  Returns an `EffectFunc` that logs a given message.
- `Perform(e EffectFunc) error`  
  Executes the provided side effect.

---

## Module System

### Overview

- **Static Dependency Injection:**  
  Novum uses static typing for modules, ensuring that dependencies are resolved at compile time.
- **Simple Container:**  
  A lightweight container allows modules (e.g., network modules) to be registered and retrieved.

### API

- `NewContainer[NetworkT NetworkModule](network NetworkT) Container[NetworkT]`  
  Creates a new module container with the provided network module.
- `GetNetwork() NetworkT`  
  Retrieves the stored network module.

---

## Contract Checking

### Overview

- **Design by Contract:**  
  Novum includes a lightweight mechanism for asserting preconditions and postconditions. If conditions are not met, errors are raised.
- **Optional Enforcement:**  
  Contracts can be enabled during development for safety and selectively disabled in production.

### API

- `Assert(condition bool, msg string)`  
  Triggers a runtime error with the provided message if the condition is false.

---

## Composite

### Overview

Novum Composite integrates state transitions, effect handling, and contract checks into a unified, chainable abstraction. It enables you to build complex operations as a chain of expressions.

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
  Executes the chain and returns the final value, state, effects, and any error.

### Future Integration

To integrate asynchronous operations, the following function converts a `future.Future[T]` into a composite chain:

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

## License

Novum is licensed under the [MIT License](LICENSE).