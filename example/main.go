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

	// Initialize the state (e.g., a counter starting at 0).
	initialState := state.NewStateLayer(0)

	// Execute the composite chain.
	finalValue, finalState, effects, err := comp.Run(initialState)

	// Print the results.
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Final Value: %d\n", finalValue)
		fmt.Printf("Final State: %+v\n", finalState)
		fmt.Println("Accumulated Effects:")
		// Execute each side effect.
		for _, eff := range effects {
			_ = eff()
		}
	}
}
