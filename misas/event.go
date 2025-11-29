package misas

import (
	"context"
	"sync"
)

type EventTypeName string
type Event interface{ TypeName() EventTypeName }
type EventHandler interface {
	Handle(context.Context, Event) error
}
type EventBus interface {
	RegisterHandler(handler EventHandler)
	Publish(context.Context, Event) error
}

type EventHandlerFunc func(context.Context, Event) error

func (f EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

type InMemoryEventBus struct {
	handlers []EventHandler
	mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: []EventHandler{},
	}
}

func (bus *InMemoryEventBus) RegisterHandler(handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers = append(bus.handlers, handler)
}

func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mu.RLock()
	handlers := make([]EventHandler, len(bus.handlers))
	copy(handlers, bus.handlers)
	bus.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
