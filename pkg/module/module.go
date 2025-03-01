// pkg/module/module.go

package module

import "sync"

type Container struct {
	modules map[string]interface{}
	mu      sync.RWMutex
}

func NewContainer() *Container {
	return &Container{modules: make(map[string]interface{})}
}

func (c *Container) Register(name string, module interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules[name] = module
}

func (c *Container) Resolve(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mod, ok := c.modules[name]
	return mod, ok
}
