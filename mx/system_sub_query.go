package mx

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"log/slog"

	"github.com/morebec/misas/misas"
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

func (qc *QuerySubsystemConf) WithQueryHandler(qt misas.QueryTypeName, h misas.QueryHandler) *QuerySubsystemConf {
	if qt == "" {
		panic(fmt.Sprintf("query subsystem %s: query type name cannot be empty", qc.name))
	}
	if h == nil {
		panic(fmt.Sprintf("query subsystem %s: handler cannot be nil", qc.name))
	}
	qc.queryHandlers[qt] = newSystemQueryHandler(qc.name, h)

	return qc
}

func (qc *QuerySubsystemConf) WithEventHandlers(eventBusName EventBusName, handlers ...misas.EventHandler) *QuerySubsystemConf {
	if eventBusName == "" {
		panic(fmt.Sprintf("query subsystem %s: event bus name cannot be empty", qc.name))
	}

	handlers = lo.Map(handlers, func(h misas.EventHandler, _ int) misas.EventHandler {
		return newSystemEventHandler(qc.name, h)
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

func newSystemQueryHandler(subsystemName string, h misas.QueryHandler) misas.QueryHandler {
	return misas.QueryHandlerFunc(func(ctx context.Context, q misas.Query) misas.QueryResult {
		origin := Ctx(ctx).SubsystemInfo().Name
		ctx = newSubsystemContext(ctx, SubsystemInfo{Name: subsystemName})
		logger := Log(ctx).With(slog.String("originSubsystem", origin))

		logger.Info(fmt.Sprintf("handling query %q", q.TypeName()), slog.String("query", string(q.TypeName())))
		result := h.Handle(ctx, q)

		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to handle query %q", q.TypeName()),
				slog.Any(logKeyError, result.Error),
				slog.String(
					"query",
					string(q.TypeName()),
				),
			)
		} else {
			logger.Info(fmt.Sprintf("successfully handled query %q", q.TypeName()), slog.String("query", string(q.TypeName())))
		}

		return result
	})
}
