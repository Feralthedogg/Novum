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
Here is a brief example demonstrating the main features of Novum.

```go
package main

import (
	"fmt"

	"github.com/Feralthedogg/Novum/contract"
	"github.com/Feralthedogg/Novum/effect"
	"github.com/Feralthedogg/Novum/module"
	"github.com/Feralthedogg/Novum/state"
)

func main() {
	// 1. State Management: Create an initial state and increment the counter.
	s := state.NewStateLayer(0)
	s = s.Increment()
	fmt.Println("Counter:", s.Counter) // Output: Counter: 1

	// 2. Effect Handling: Execute a log effect.
	err := effect.Perform(effect.LogEffect{Message: "Processing data..."})
	if err != nil {
		fmt.Println("Effect error:", err)
	}

	// 3. Module System: Register and resolve a network module.
	container := module.NewContainer()
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
  The side effect functions are written simply, allowing them to be efficiently executed and easily inlined by the compiler.

### API
- **`Effect` Interface**  
  All side effects must implement the `Handle() error` method.
- **`LogEffect` Structure**  
  Represents a simple logging effect with a message.
- `Perform(e Effect) error`  
  Executes the given side effect.

---

## Module System

### Overview
- **Static Type-Based Dependency Injection:**  
  Novum uses static typing for its module system, ensuring dependencies are resolved at compile time.
- **Simple Container Implementation:**  
  A straightforward container allows modules to be registered and looked up using a map, keeping the design simple and efficient.

### API
- `NewContainer() *Container`  
  Creates a new module container.
- `(c *Container) Register(name string, module interface{})`  
  Registers a module under the specified name.
- `(c *Container) Resolve(name string) (interface{}, bool)`  
  Retrieves a registered module by name.

---

## Contract Checking

### Overview
- **Design by Contract:**  
  Novum includes a lightweight system to assert preconditions and postconditions in your functions. If conditions fail, the system will alert you immediately.
- **Optional Enforcement:**  
  Contracts can be enabled during development for safety and debugging, and selectively disabled in production for performance.

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
- `Return[T any](value T) NovumComposite[T]`  
  Wraps a value into a composite with default settings.
- `Bind(func(T) NovumComposite[T]) NovumComposite[T]`  
  Chains operations while combining state transitions and side effects.
- `WithEffect(e Effect) NovumComposite[T]`  
  Appends a side effect to the composite.
- `WithContract(func(T) bool) NovumComposite[T]`  
  Sets the contract function for value validation.
- `WithModule(*Container) NovumComposite[T]`  
  Sets the module container for dependency injection.
- `ResolveModule(name string) (interface{}, bool)`  
  Retrieves a module by name from the attached module container.
- `Run(initialState StateLayer) (T, StateLayer, []Effect, *Container, error)`  
  Executes the composite chain, returning the final value, state, accumulated side effects, module container, and error (error is the last returned value).

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

func main() {
	// Build a composite chain starting with an initial integer value.
	comp := composite.Return(10).
		// Set a contract to ensure the value is non-negative.
		WithContract(func(n int) bool {
			return n >= 0
		}).
		// Bind operation: add 10 to the value.
		Bind(func(n int) composite.NovumComposite[int] {
			newValue := n + 10
			return composite.Return(newValue).
				WithEffect(effect.LogEffect{Message: "Added 10 to the value"})
		}).
		// Bind operation: multiply the value by 2.
		Bind(func(n int) composite.NovumComposite[int] {
			newValue := n * 2
			return composite.Return(newValue).
				WithEffect(effect.LogEffect{Message: "Multiplied the value by 2"})
		}).
		// Bind operation: register a network module.
		Bind(func(n int) composite.NovumComposite[int] {
			cont := module.NewContainer()
			cont.Register("network", DefaultNetworkModule{})
			return composite.Return(n).
				WithModule(cont).
				WithEffect(effect.LogEffect{Message: "Registered network module"})
		}).
		// Bind operation: simulate network module usage.
		Bind(func(n int) composite.NovumComposite[int] {
			mod, ok := composite.Return(n).ResolveModule("network")
			if ok {
				network := mod.(NetworkModule)
				data, err := network.Fetch("https://api.example.com/data")
				if err != nil {
					return composite.Return(n).
						WithEffect(effect.LogEffect{Message: "Error fetching data from network"})
				}
				return composite.Return(n).
					WithEffect(effect.LogEffect{Message: "Fetched data: " + data})
			}
			return composite.Return(n).
				WithEffect(effect.LogEffect{Message: "Network module not found"})
		})

	// Create an initial state.
	initialState := state.NewStateLayer(0)

	// Run the composite chain.
	finalValue, finalState, effects, modules, err := comp.Run(initialState)

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
		if modules != nil {
			if mod, ok := modules.Resolve("network"); ok {
				fmt.Println("Network module is registered:", mod)
			}
		}
	}
}

// Dummy NetworkModule implementation for composite example.
type NetworkModule interface {
	Fetch(url string) (string, error)
}

type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}
```

---

## License
Novum is licensed under the [MIT License](LICENSE).

---