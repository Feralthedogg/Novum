// pkg/module/module.go

package module

type Container[NetworkT any] struct {
	Network NetworkT
}

func NewContainer[NetworkT any](network NetworkT) Container[NetworkT] {
	return Container[NetworkT]{Network: network}
}

func (c Container[NetworkT]) GetNetwork() NetworkT {
	return c.Network
}
