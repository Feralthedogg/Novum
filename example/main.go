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

// FromFuture converts a future.Future[T] into a NovumComposite[T, Deps].
// If the Future returns an error, a composite with a failing contract is returned,
// causing Run() to fail, and the error is logged as a side effect.
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

	// ---------------------------
	// Synchronous Composite Example
	// ---------------------------
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

	// ---------------------------
	// Future Composite Example
	// ---------------------------
	// Create a Future that asynchronously returns 42 after 100ms.
	fut := future.NewFuture(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})
	// Convert the Future into a Composite chain.
	futureComp := FromFuture(fut, deps).
		Bind(func(n int, deps module.Container[NetworkModule]) composite.NovumComposite[int, module.Container[NetworkModule]] {
			newValue := n * 3
			return composite.Return(newValue, deps).
				WithEffect(effect.NewLogEffect(fmt.Sprintf("Future result multiplied by 3: %d", newValue)))
		})
	futureResult, _, _, err := futureComp.Run(state.NewStateLayer[int](0))
	if err != nil {
		fmt.Println("Future Composite Error:", err)
	} else {
		fmt.Printf("Future Composite - Final Result: %d\n", futureResult)
	}

	// ---------------------------
	// Parallel Composite Example
	// ---------------------------
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
