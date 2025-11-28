package mx

import (
	"fmt"
	"sync/atomic"
)

// LateBinding is a binding that can be bound only once, and must be bound before
// getting its value.
//
// If a binding is intended to be mutable (i.e. can be rebound multiple times), use
// DynamicBinding instead.
//
// NOTE: A binding allows decoupling references to objects and their
// actual instantiation. This is useful for dependency injection and avoiding
// circular dependencies.
type LateBinding[T any] struct {
	ptr   atomic.Value
	bound atomic.Bool
}

func NewLateBinding[T any]() *LateBinding[T] { return &LateBinding[T]{} }

func (l *LateBinding[T]) Get() T {
	if !l.bound.Load() {
		panic("late binding: value not bound yet")
	}
	return l.ptr.Load().(T)
}

func (l *LateBinding[T]) Bind(value T) {
	if l.bound.Swap(true) {
		panic(fmt.Sprintf("late binding %T: cannot bind a value already bound", value))
	}
	l.ptr.Store(value)
}

func (l *LateBinding[T]) IsBound() bool { return l.bound.Load() }

// DynamicBinding is a binding that can be rebound multiple times, but must be
// bound before getting its value.
//
// If a binding is intended to be immutable (i.e. can be bound only once), use
// LateBinding instead.
//
// NOTE: A binding allows decoupling references to objects and their
// actual instantiation. This is useful for dependency injection and avoiding
// circular dependencies.
type DynamicBinding[T any] struct {
	ptr   atomic.Value
	bound atomic.Bool
}

func NewDynamicBinding[T any]() *DynamicBinding[T] { return &DynamicBinding[T]{} }

func (d *DynamicBinding[T]) Bind(value T) {
	d.ptr.Store(value)
	d.bound.Store(true)
}

func (d *DynamicBinding[T]) Get() T {
	if !d.bound.Load() {
		panic(fmt.Sprintf("dynamic binding %T: value not bound", *new(T)))
	}
	v := d.ptr.Load()
	return v.(T)
}

func (d *DynamicBinding[T]) IsBound() bool { return d.bound.Load() }
