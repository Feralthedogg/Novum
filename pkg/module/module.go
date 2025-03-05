// pkg/module/module.go

package module

// NetworkModule defines a simple network fetching interface.
type NetworkModule interface {
	Fetch(url string) (string, error)
}

// Container restricts the stored module to types that implement NetworkModule.
type Container[NetworkT NetworkModule] struct {
	Network NetworkT
}

// NewContainer creates a new container with the provided network module.
func NewContainer[NetworkT NetworkModule](network NetworkT) Container[NetworkT] {
	return Container[NetworkT]{Network: network}
}

// GetNetwork returns the stored network module.
func (c Container[NetworkT]) GetNetwork() NetworkT {
	return c.Network
}
