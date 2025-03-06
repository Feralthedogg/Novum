// pkg/module/module.go
package module

import "sync"

// Container is a dependency injection container that can store and resolve multiple modules by key.
type Container struct {
	modules map[string]interface{}
	mu      sync.RWMutex
}

// NewContainer creates a new empty dependency container.
func NewContainer() *Container {
	return &Container{
		modules: make(map[string]interface{}),
	}
}

// Register stores the given module under the specified key.
func (c *Container) Register(key string, module interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules[key] = module
}

// Resolve retrieves the module stored under the given key.
// It returns the module as an interface{} and a boolean indicating whether it was found.
func (c *Container) Resolve(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mod, ok := c.modules[key]
	return mod, ok
}

// ResolveAs retrieves the module stored under the given key and asserts it to type T.
// It returns the module of type T and a boolean indicating success.
func ResolveAs[T any](c *Container, key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mod, ok := c.modules[key]
	if !ok {
		var zero T
		return zero, false
	}
	result, ok := mod.(T)
	return result, ok
}
