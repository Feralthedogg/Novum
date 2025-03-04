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
8. [License](#license)

---

## Installation

Novum is managed as a Go module. You can install it using the following command:

```bash
go get github.com/Feralthedogg/Novum
```

---

## Quick Start

Below is a brief example demonstrating how to use Novum’s key features in a composite chain. This example shows both a synchronous chain and an asynchronous Future chain integrated with the Composite API.

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

func main() {
	// Create a dependency container with the network module.
	deps := module.NewContainer[NetworkModule](DefaultNetworkModule{})

	// --- Synchronous Composite Example ---
	comp := composite.Return(10, deps).
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
	initialState := state.NewStateLayer(0)
	finalValue, finalState, effects, err := comp.Run(initialState)
	if err != nil {
		fmt.Println("Synchronous Composite Error:", err)
	} else {
		fmt.Printf("Synchronous Composite - Final Value: %d\n", finalValue)
		fmt.Printf("Synchronous Composite - Final State: %+v\n", finalState)
		fmt.Println("Synchronous Composite - Executing Effects:")
		for _, eff := range effects {
			_ = eff() // Execute side effects.
		}
	}

	// --- Future Composite Example ---
	// Create a Future that asynchronously returns 42 after 100ms.
	fut := future.NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})
	// Convert the Future into a Composite chain using FromFuture.
	compFuture := FromFuture(fut, deps).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			// Multiply the Future result by 3.
			newValue := n * 3
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future result multiplied by 3: %d", newValue)))
		})
	futureResult, _, _, err := compFuture.Run(initialState)
	if err != nil {
		fmt.Println("Future Composite Error:", err)
	} else {
		fmt.Printf("Future Composite - Final Result: %d\n", futureResult)
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

- `NewStateLayer(initial int) StateLayer`  
  Creates a new state object with the specified initial value.
- `(s StateLayer) Increment() StateLayer`  
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

- `NewContainer[NetworkT any](network NetworkT) Container[NetworkT]`  
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

Novum Composite integrates state transitions, effect handling, and contract checks into a unified, chainable abstraction. This allows you to build complex operations as a chain of expressions.

### API

- `Return[T any, Deps any](value T, deps Deps) NovumComposite[T, Deps]`  
  Wraps a value and its dependencies into a composite.
- `Bind(func(T, Deps) NovumComposite[T, Deps]) NovumComposite[T, Deps]`  
  Chains operations together.
- `WithEffect(e EffectFunc) NovumComposite[T, Deps]`  
  Adds a side effect to the chain.
- `WithContract(func(T) bool) NovumComposite[T, Deps]`  
  Sets a contract for validation.
- `Run(initialState StateLayer) (T, StateLayer, []EffectFunc, error)`  
  Executes the chain and returns the final value, state, effects, and any error.

### Future Integration

To integrate asynchronous operations, the following function converts a `future.Future[T]` into a composite chain:

```go
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
```

### Example Usage (see Quick Start above)

The example above demonstrates using Future in a composite chain by:
- Creating a Future with `future.NewFuture`.
- Converting it to a Composite using `FromFuture`.
- Chaining additional operations with `Bind`.

---

## License

Novum is licensed under the [MIT License](LICENSE).