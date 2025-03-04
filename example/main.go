// main.go
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

// Fetch returns a dummy data string from the given URL.
func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

func main() {
	// Create a compile-time dependency container with the network module.
	deps := module.NewContainer[NetworkModule](DefaultNetworkModule{})

	// ---------------------------
	// Synchronous Composite Example
	// ---------------------------
	comp := composite.Return[int, module.Container[NetworkModule]](10, deps).
		// Set a contract to ensure the value is non-negative.
		WithContract(func(n int) bool {
			return n >= 0
		}).
		// Bind: add 10 to the value.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n + 10
			return composite.Return[int, module.Container[NetworkModule]](newValue, deps).
				WithEffect(effect.NewLogEffect("Added 10 to the value"))
		}).
		// Bind: multiply the value by 2.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n * 2
			return composite.Return[int, module.Container[NetworkModule]](newValue, deps).
				WithEffect(effect.NewLogEffect("Multiplied the value by 2"))
		}).
		// Bind: log that the network module is registered.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			return composite.Return[int, module.Container[NetworkModule]](n, deps).
				WithEffect(effect.NewLogEffect("Network module is registered"))
		}).
		// Bind: use the network module to fetch data.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			data, err := deps.GetNetwork().Fetch("https://api.example.com/data")
			if err != nil {
				return composite.Return[int, module.Container[NetworkModule]](n, deps).
					WithEffect(effect.NewLogEffect("Error fetching data from network"))
			}
			return composite.Return[int, module.Container[NetworkModule]](n, deps).
				WithEffect(effect.NewLogEffect("Fetched data: " + data))
		})

	// Initialize the state (e.g., a counter starting at 0).
	initialState := state.NewStateLayer(0)

	// Execute the synchronous composite chain.
	finalValue, finalState, effects, err := comp.Run(initialState)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Synchronous Composite - Final Value: %d\n", finalValue)
		fmt.Printf("Synchronous Composite - Final State: %+v\n", finalState)
		fmt.Println("Synchronous Composite - Accumulated Effects:")
		for _, eff := range effects {
			_ = eff() // Execute each side effect.
		}
	}

	// ---------------------------
	// Future Composite Example
	// ---------------------------
	// Create a Future that returns an int asynchronously.
	fut := future.NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond) // simulate delay
		return 100, nil
	})

	// Convert Future to a Composite chain.
	compFuture := FromFuture[int, module.Container[NetworkModule]](fut, deps).
		// Bind: multiply the future result by 3.
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n * 3
			return composite.Return[int, module.Container[NetworkModule]](newValue, deps).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future result multiplied by 3: %d", newValue)))
		})

	// Execute the Future composite chain.
	futureResult, _, _, err := compFuture.Run(initialState)
	if err != nil {
		fmt.Println("Future Composite Error:", err)
	} else {
		fmt.Printf("Future Composite - Final Result: %d\n", futureResult)
	}
}

// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// If the Future returns an error, a composite with a failing contract is returned,
// causing Run() to fail, and the error is logged as a side effect.
func FromFuture[T any, Deps any](f future.Future[T], deps Deps) composite.NovumComposite[T, Deps] {
	return composite.Return[T, Deps](*new(T), deps).Bind(func(_ T, deps Deps) composite.NovumComposite[T, Deps] {
		res, err := f.Await()
		if err != nil {
			// Return a composite that fails its contract (always returns false)
			// and logs the error message.
			return composite.Return[T, Deps](*new(T), deps).
				WithContract(func(val T) bool { return false }).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future error: %v", err)))
		}
		return composite.Return[T, Deps](res, deps)
	})
}
