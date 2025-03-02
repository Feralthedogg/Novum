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
	err := effect.Perform(effect.NewLogEffect("Processing data..."))
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

## Effect Handling

### Overview

- **Explicit Side Effect Separation:**\
  Side effects like logging, network I/O, or file operations are encapsulated via the `Effect` interface. This design separates pure functions from impure ones.
- **Optimized Side Effect Processing:**\
  The side effect functions are simple and can be efficiently executed and inlined by the compiler.

### API

- **`EffectFunc`**\*\* Type\*\*\
  A function type that represents a side effect.
- **`NewLogEffect`**\*\* Function\*\*\
  Returns an `EffectFunc` that logs a given message.
- `Perform(e EffectFunc) error`\
  Executes the given side effect.

---

## Composite

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

type NetworkModule interface {
	Fetch(url string) (string, error)
}

type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

func main() {
	// Create a compile-time dependency container with the network module.
	deps := module.NewContainer[NetworkModule](DefaultNetworkModule{})

	// Create a composite chain with an initial value of 10 and the injected dependencies.
	comp := composite.Return(10, deps).
		WithContract(func(n int) bool {
			return n >= 0
		}).
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
			return composite.Return(n, deps).
				WithEffect(effect.NewLogEffect("Network module is registered"))
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
		for _, eff := range effects {
			_ = eff()
		}
	}
}
```

---

## License

Novum is licensed under the [MIT License](LICENSE).

