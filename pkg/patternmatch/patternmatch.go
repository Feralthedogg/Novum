// pkg/patternmatch/patternmatch.go

package patternmatch

import (
	"fmt"
	"reflect"
)

type Env struct {
	parent   *Env
	bindings map[string]interface{}
}

func NewEnv() *Env {
	return &Env{
		parent:   nil,
		bindings: make(map[string]interface{}),
	}
}

func (e *Env) Extend(key string, value interface{}) *Env {
	return &Env{
		parent:   e,
		bindings: map[string]interface{}{key: value},
	}
}

func (e *Env) Lookup(key string) (interface{}, bool) {
	if val, ok := e.bindings[key]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Lookup(key)
	}
	return nil, false
}

type Pattern interface {
	Match(value interface{}, env *Env) (bool, *Env, error)
}

// --------------------
// Literal Pattern
// --------------------

type Literal[T comparable] struct {
	Value T
}

func (l Literal[T]) Match(value interface{}, env *Env) (bool, *Env, error) {
	v, ok := value.(T)
	if !ok {
		return false, env, fmt.Errorf("Literal: type mismatch, expected %T", l.Value)
	}
	if v == l.Value {
		return true, env, nil
	}
	return false, env, fmt.Errorf("Literal: %v does not equal %v", v, l.Value)
}

// --------------------
// Var Pattern
// --------------------

type Var struct {
	Name string
}

func (v Var) Match(value interface{}, env *Env) (bool, *Env, error) {
	if oldVal, ok := env.Lookup(v.Name); ok {
		if reflect.DeepEqual(oldVal, value) {
			return true, env, nil
		}
		return false, env, fmt.Errorf("Var: variable '%s' already bound to %v, cannot rebind to %v", v.Name, oldVal, value)
	}
	newEnv := env.Extend(v.Name, value)
	return true, newEnv, nil
}

// --------------------
// Pin Pattern
// --------------------

type Pin struct {
	Name string
}

func (p Pin) Match(value interface{}, env *Env) (bool, *Env, error) {
	oldVal, ok := env.Lookup(p.Name)
	if !ok {
		return false, env, fmt.Errorf("Pin: variable '%s' not bound", p.Name)
	}
	if reflect.DeepEqual(oldVal, value) {
		return true, env, nil
	}
	return false, env, fmt.Errorf("Pin: variable '%s' bound to %v, does not match %v", p.Name, oldVal, value)
}

// --------------------
// Cons Pattern
// --------------------

type Cons struct {
	Head Pattern
	Tail Pattern
}

func (c Cons) Match(value interface{}, env *Env) (bool, *Env, error) {
	rv := reflect.ValueOf(value)
	kind := rv.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return false, env, fmt.Errorf("Cons: value is not a slice or array")
	}
	if rv.Len() == 0 {
		return false, env, fmt.Errorf("Cons: empty list cannot match")
	}
	ok, newEnv, err := c.Head.Match(rv.Index(0).Interface(), env)
	if !ok {
		return false, env, fmt.Errorf("Cons: head match failed: %w", err)
	}
	tail := make([]interface{}, 0, rv.Len()-1)
	for i := 1; i < rv.Len(); i++ {
		tail = append(tail, rv.Index(i).Interface())
	}
	ok, finalEnv, err := c.Tail.Match(tail, newEnv)
	if !ok {
		return false, env, fmt.Errorf("Cons: tail match failed: %w", err)
	}
	return true, finalEnv, nil
}

// --------------------
// Wildcard Pattern
// --------------------

type Wildcard struct{}

func (Wildcard) Match(value interface{}, env *Env) (bool, *Env, error) {
	return true, env, nil
}
