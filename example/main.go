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
	// Build a composite chain with an initial integer value.
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
			// For demonstration, attempt to resolve the "network" module.
			// Note: This uses a workaround since the module container is not directly
			// accessible inside the Bind function.
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

	// Initialize the state (e.g., a counter starting at 0).
	initialState := state.NewStateLayer(0)

	// Execute the composite chain.
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
