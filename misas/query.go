package misas

import (
	"context"
	"sync"
)

type QueryTypeName string
type Query interface{ TypeName() QueryTypeName }
type QueryResult struct {
	Payload any
}
type QueryHandler interface {
	Handle(context.Context, Query) QueryResult
}

type QueryHandlerFunc func(context.Context, Query) QueryResult

func (f QueryHandlerFunc) Handle(ctx context.Context, query Query) QueryResult {
	return f(ctx, query)
}

type TypedQueryHandler[T Query] interface {
	Handle(context.Context, T) QueryResult
}

type QueryBus interface {
	HandleQuery(context.Context, Query) QueryResult
	RegisterHandler(QueryTypeName, QueryHandler)
}

type InMemoryQueryBus struct {
	handlers map[QueryTypeName]QueryHandler
	mu       sync.Mutex
}

func NewInMemoryQueryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{
		handlers: make(map[QueryTypeName]QueryHandler),
	}
}

func (b *InMemoryQueryBus) HandleQuery(ctx context.Context, query Query) QueryResult {
	if query == nil {
		panic("query cannot be nil")
	}
	b.mu.Lock()
	handler, ok := b.handlers[query.TypeName()]
	b.mu.Unlock()

	if !ok {
		panic("no handler registered for query type: " + query.TypeName())
	}

	return handler.Handle(ctx, query)
}

func (b *InMemoryQueryBus) RegisterHandler(queryType QueryTypeName, handler QueryHandler) {
	if queryType == "" {
		panic("query type name cannot be empty")
	}
	if handler == nil {
		panic("query handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.handlers[queryType]; exists {
		panic("handler already registered for query type: " + queryType)
	}

	b.handlers[queryType] = handler
}
