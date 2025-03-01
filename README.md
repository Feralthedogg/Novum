# Novum

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [State Management](#state-management)
4. [Effect Handling](#effect-handling)
5. [Module System](#module-system)
6. [Contract Checking](#contract-checking)
7. [License](#license)

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
    Novum creates a new state object rather than modifying the existing state when changes occur. For Example, incrementing a counter produces a new state, ensuring the original state remains unchanged.
- **Context-Based Management:**
    Each request or session can have its own state context, allowing independent and predictable state handling in concurrent environments

### API
- `NewStateLayer(initial int) StateLayer`

    Creates a new State object with the given initial value.
- `(s StateLayer) Increment() StateLayer`

    Returns a new State object with the counter incremented.

---

## Effect Handling
### Overview
- **Explicit Side Effect Separation:**

    Side effects like logging, network I/O, or file operations are encapsulated via the `Effect` interface. This design separates pure functions from impure ones.

- **Optimized Side Effect Processing:**

    The side effect functions are written simply, allowing them to be efficiently executed and easily inlined by the compiler.

### API
- `Effect` **Interface**

    All side effects must implement the `Handle() error` method.
- `LogEffect` **Structure**

    Represents a simple logging effect with a message.

- `Perform(e Effect) error`

    Executes the given side effect.

---

## Module System
### Overview
- **Static Type-Based Dependency Injection:**

    Novum uses static typing for its module system, ensuring dependencies are resolved at compile time.

- **Simple Container Implementation:**

    A straightforward container allows modules to be registered and looked up using a map, keeping the design simple adn efficient.

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

## License
Novum is licensed under the [MIT License](LICENSE).