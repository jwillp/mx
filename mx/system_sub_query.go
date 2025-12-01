package mx

import (
	"context"
	"fmt"
	"github.com/morebec/misas/misas"
	"github.com/samber/lo"
)

type QuerySubsystemConf struct {
	name          string
	queryHandlers map[misas.QueryTypeName]misas.QueryHandler
	eventHandlers map[EventBusName][]misas.EventHandler
}

func NewQuerySubsystem(name string) *QuerySubsystemConf {
	if name == "" {
		panic("name of query subsystem cannot be empty")

	}
	return &QuerySubsystemConf{
		name:          name,
		queryHandlers: make(map[misas.QueryTypeName]misas.QueryHandler),
	}
}

// WithQueryHandler registers a query handler for the given query type with the system's query bus.
// It also automatically the query type in the global query registry for serialization purposes.
func (qc *QuerySubsystemConf) WithQueryHandler(qt misas.Query, h misas.QueryHandler) *QuerySubsystemConf {
	if qt == nil {
		panic(fmt.Sprintf("query subsystem %s: query cannot be empty", qc.name))
	}
	if h == nil {
		panic(fmt.Sprintf("query subsystem %s: handler cannot be nil", qc.name))
	}
	h = withQueryLogging(h)
	h = withQueryContextPropagation(qc.name, h)
	qc.queryHandlers[qt.TypeName()] = h
	QueryRegistry.Register(qt.TypeName(), qt)

	return qc
}

// WithEventHandlers registers event handlers for the given event bus name with the system's event buses.
func (qc *QuerySubsystemConf) WithEventHandlers(eventBusName EventBusName, handlers ...misas.EventHandler) *QuerySubsystemConf {
	if eventBusName == "" {
		panic(fmt.Sprintf("query subsystem %s: event bus name cannot be empty", qc.name))
	}

	handlers = lo.Map(handlers, func(h misas.EventHandler, _ int) misas.EventHandler {
		h = withEventLogging(h)
		h = withEventContextPropagation(qc.name, h)
		return h
	})
	qc.eventHandlers[eventBusName] = append(qc.eventHandlers[eventBusName], handlers...)

	return qc
}

type DynamicBindingQueryBus struct {
	*DynamicBinding[misas.QueryBus]
}

func NewDynamicBindingQueryBus() *DynamicBindingQueryBus {
	return &DynamicBindingQueryBus{DynamicBinding: NewDynamicBinding[misas.QueryBus]()}
}

func (d *DynamicBindingQueryBus) HandleQuery(ctx context.Context, query misas.Query) misas.QueryResult {
	return d.Get().HandleQuery(ctx, query)
}

func (d *DynamicBindingQueryBus) RegisterHandler(queryType misas.QueryTypeName, handler misas.QueryHandler) {
	d.Get().RegisterHandler(queryType, handler)
}
