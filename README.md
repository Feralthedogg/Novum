# Novum

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [State Management](#state-management)
4. [Effect Handling](#effect-handling)
5. [Module System](#module-system)
6. [Contract Checking](#contract-checking)
7. [Composite](#composite)
8. [License](#license)

---

## Installation
Novum is managed as a Go module. You can install it using the following command:
```bash
go get github.com/Feralthedogg/Novum
```

---

## Quick Start
Below is a brief example demonstrating the main features of Novum.

```go
package main

import (
	"fmt"

	"github.com/Feralthedogg/Novum/pkg/contract"
	"github.com/Feralthedogg/Novum/pkg/effect"
	"github.com/Feralthedogg/Novum/pkg/module"
	"github.com/Feralthedogg/Novum/pkg/state"
)

func main() {
	// 1. State Management: Create an initial state and increment the counter.
	s := state.NewStateLayer(0)
	s = s.Increment()
	fmt.Println("Counter:", s.Counter) // Output: Counter: 1

	// 2. Effect Handling: Execute a log effect.
	err := effect.Perform(effect.LogEffect("Processing data..."))
	if err != nil {
		fmt.Println("Effect error:", err)
	}

	// 3. Module System: Register and resolve a network module.
	container := module.NewContainer(DefaultNetworkModule{})
	container.Register("network", DefaultNetworkModule{})
	if mod, ok := container.Resolve("network"); ok {
		network := mod.(NetworkModule)
		data, _ := network.Fetch("https://api.example.com/data")
		fmt.Println(data)
	}

	// 4. Contract Checking: Validate function inputs.
	result := Add(10, 20)
	fmt.Println("Result:", result)
}

type NetworkModule interface {
	Fetch(url string) (string, error)
}

type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

// Add validates that the inputs are non-negative before performing addition.
func Add(a, b int) int {
	contract.Assert(a >= 0 && b >= 0, "Input values must be non-negative")
	return a + b
}
```

---

## State Management

### Overview
- **Immutable State Transitions:**  
  Novum creates a new state object rather than modifying the existing state when changes occur. For example, incrementing a counter produces a new state, ensuring the original state remains unchanged.
- **Context-Based Management:**  
  Each request or session can have its own state context, allowing independent and predictable state handling in concurrent environments.

### API
- `NewStateLayer(initial int) StateLayer`  
  Creates a new state object with the given initial value.
- `(s StateLayer) Increment() StateLayer`  
  Returns a new state object with the counter incremented.

---

## Effect Handling

### Overview
- **Explicit Side Effect Separation:**  
  Side effects like logging, network I/O, or file operations are encapsulated via the `Effect` interface. This design separates pure functions from impure ones.
- **Optimized Side Effect Processing:**  
  The side effect functions are simple and can be efficiently executed and inlined by the compiler.

### API
- **`EffectFunc` Type**  
  A function type that represents a side effect.
- **`LogEffect` Function**  
  Returns an `EffectFunc` that logs a given message.
- `Perform(e EffectFunc) error`  
  Executes the given side effect.

---

## Module System

### Overview
- **Static Type-Based Dependency Injection:**  
  Novum uses static typing for its module system, ensuring dependencies are resolved at compile time.
- **Simple Container Implementation:**  
  A straightforward container allows modules to be registered and looked up using a map.

### API
- `NewContainer(network NetworkT) Container[NetworkT]`  
  Creates a new module container with a provided network module.
- `(c Container[NetworkT]) Register(name string, module interface{})`  
  Registers a module under the specified name.
- `(c Container[NetworkT]) Resolve(name string) (interface{}, bool)`  
  Retrieves a registered module by name.
- `(c Container[NetworkT]) GetNetwork() NetworkT`  
  Retrieves the network module.

---

## Contract Checking

### Overview
- **Design by Contract:**  
  Novum includes a lightweight system to assert preconditions and postconditions in your functions. If conditions fail, the system will alert you immediately.
- **Optional Enforcement:**  
  Contracts can be enabled during development for safety and debugging, and selectively disabled in production.

### API
- `Assert(condition bool, msg string)`  
  Triggers a runtime error with a message if the condition is false.

---

## Composite

### Overview
Novum Composite integrates state management, effect handling, module system, and contract checking into a single unified abstraction. This composite allows you to chain operations while accumulating state transitions, side effects, and contract validations in one continuous flow.

### Key Features
- **Unified Chaining:**  
  Use `Bind` to chain operations that combine value transformations, state updates, and side effect accumulation.
- **Integrated Contract Checking:**  
  Set a contract function with `WithContract` that validates each step in the chain.
- **Module Integration:**  
  Attach a module container with `WithModule` and resolve dependencies using `ResolveModule`.
- **Error Propagation:**  
  Any error in the chain will be captured and prevent further processing.

### API
- `Return[T any, Deps any](value T, deps Deps) NovumComposite[T, Deps]`  
  Wraps a value and its dependencies into a composite with default settings.
- `Bind(func(T, Deps) NovumComposite[T, Deps]) NovumComposite[T, Deps]`  
  Chains operations while combining state transitions and side effects.
- `WithEffect(e EffectFunc) NovumComposite[T, Deps]`  
  Appends a side effect to the composite.
- `WithContract(func(T) bool) NovumComposite[T, Deps]`  
  Sets the contract function for value validation.
- `WithModule(deps Deps) NovumComposite[T, Deps]`  
  Sets the module container for dependency injection.
- `ResolveModule(name string) (interface{}, bool)`  
  Retrieves a module by name from the attached module container.
- `Run(initialState StateLayer) (T, StateLayer, []EffectFunc, error)`  
  Executes the composite chain, returning the final value, state, accumulated side effects, and error (with error as the last return value).

### Example Usage
```go
package main

import (
	"fmt"
	"github.com/Feralthedogg/Novum/pkg/composite"
	"github.com/Feralthedogg/Novum/pkg/effect"
	"github.com/Feralthedogg/Novum/pkg/module"
	"github.com/Feralthedogg/Novum/pkg/state"
)

// NetworkModule defines a simple network fetching interface.
type NetworkModule interface {
	Fetch(url string) (string, error)
}

// DefaultNetworkModule is a basic implementation of NetworkModule.
type DefaultNetworkModule struct{}

// Fetch returns a dummy data string from the given URL.
func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

func main() {
	// Create a compile-time dependency container with the network module.
	deps := module.NewContainer[NetworkModule](DefaultNetworkModule{})

	// Create a composite chain with an initial value of 10 and the injected dependencies.
	comp := composite.Return(10, deps).
		// Set a contract to ensure the value is non-negative.
		WithContract(func(n int) bool {
			return n >= 0
		}).
		// Bind: add 10 to the value.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n + 10
			return composite.Return(newValue, deps).
				WithEffect(effect.LogEffect("Added 10 to the value"))
		}).
		// Bind: multiply the value by 2.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n * 2
			return composite.Return(newValue, deps).
				WithEffect(effect.LogEffect("Multiplied the value by 2"))
		}).
		// Bind: log that the network module is registered.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			return composite.Return(n, deps).
				WithEffect(effect.LogEffect("Network module is registered"))
		}).
		// Bind: use the network module to fetch data.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			data, err := deps.GetNetwork().Fetch("https://api.example.com/data")
			if err != nil {
				return composite.Return(n, deps).
					WithEffect(effect.LogEffect("Error fetching data from network"))
			}
			return composite.Return(n, deps).
				WithEffect(effect.LogEffect("Fetched data: " + data))
		})

	// Create an initial state.
	initialState := state.NewStateLayer(0)

	// Run the composite chain.
	finalValue, finalState, effects, err := comp.Run(initialState)

	// Print the results.
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Final Value: %d\n", finalValue)
		fmt.Printf("Final State: %+v\n", finalState)
		fmt.Println("Accumulated Effects:")
		for _, e := range effects {
			if logEff, ok := e.(effect.LogEffect); ok {
				fmt.Println(" -", logEff.Message)
			}
		}
	}
}
```

---

## License
Novum is licensed under the [MIT License](LICENSE).

---