package main

import (
	"fmt"

	"github.com/Feralthedogg/novum/pkg/contract"
	"github.com/Feralthedogg/novum/pkg/effect"
	"github.com/Feralthedogg/novum/pkg/module"
	"github.com/Feralthedogg/novum/pkg/state"
)

type NetworkModule interface {
	Fetch(url string) (string, error)
}

type DefaultNetworkModule struct{}

func (n DefaultNetworkModule) Fetch(url string) (string, error) {
	return "data from " + url, nil
}

func main() {
	s := state.NewStateLayer(0)
	s = s.Increment()
	fmt.Println("Counter:", s.Counter)

	err := effect.Perform(effect.LogEffect{Message: "Processing data..."})
	if err != nil {
		fmt.Println("Effect error:", err)
	}

	container := module.NewContainer()
	container.Register("network", DefaultNetworkModule{})

	if mod, ok := container.Resolve("network"); ok {
		network := mod.(NetworkModule)
		data, _ := network.Fetch("https://api.example.com/data")
		fmt.Println(data)
	}

	result := Add(10, 20)
	fmt.Println("Result:", result)
}

func Add(a, b int) int {
	contract.Assert(a >= 0 && b >= 0, "입력 값은 음수가 아니어야 합니다")
	return a + b
}
