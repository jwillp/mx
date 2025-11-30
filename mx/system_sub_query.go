package mx

import (
	"context"

	"github.com/morebec/misas/misas"
)

type QuerySubsystemConf struct {
	name          string
	queryHandlers map[misas.QueryTypeName]misas.QueryHandler
}

func NewQuerySubsystem(name string) *QuerySubsystemConf {
	if name == "" {
		panic("query subsystem name cannot be empty")
	}
	return &QuerySubsystemConf{
		name:          name,
		queryHandlers: make(map[misas.QueryTypeName]misas.QueryHandler),
	}
}

func (qc *QuerySubsystemConf) WithQueryHandler(qt misas.QueryTypeName, h misas.QueryHandler) *QuerySubsystemConf {
	if qt == "" {
		panic("query subsystem: " + qc.name + " query type name cannot be empty")
	}
	if h == nil {
		panic("query subsystem: " + qc.name + " query handler cannot be nil")
	}
	qc.queryHandlers[qt] = h
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
