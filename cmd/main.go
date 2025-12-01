package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/morebec/misas/mtime"
	"time"

	"github.com/morebec/misas/misas"
	"github.com/morebec/misas/mx"
)

func main() {
	system := mx.NewSystem("MyApp").
		//WithEnvironment(mx.EnvironmentProduction).
		WithEnvironment(mx.EnvironmentDevelopment).
		WithDebug(false).
		WithClock(mtime.NewRealTimeClock(time.UTC))

	system.WithBusinessSubsystem(
		mx.NewBusinessSubsystem("inventory").
			WithCommandHandler(SomeCommand{}, misas.CommandHandlerFunc(func(ctx context.Context, cmd misas.Command) misas.CommandResult {
				return misas.CommandResult{
					Payload: fmt.Errorf("some command failed: inventory is out of stock"),
					//Payload: fmt.Errorf("some command failed: failed to publish event: %w", inventoryEventBus.Publish(ctx, SomeEvent{})),
				}
			})).
			WithEventHandlers("inventory", misas.EventHandlerFunc(func(ctx context.Context, event misas.Event) error {
				return errors.New("inventory event handler failed")
			})),
	)

	system.WithQuerySubsystem(
		mx.NewQuerySubsystem("inventory_reporting").
			WithQueryHandler(GetStockQuery{}, misas.QueryHandlerFunc(func(ctx context.Context, query misas.Query) misas.QueryResult {
				return misas.QueryResult{
					Payload: 100,
				}
			})),
	)

	supervisor := mx.NewSupervisor().
		WithApplicationSubsystem(HelloWorldApplicationSubsystem{
			clock: system.Clock(),
			cb:    system.CommandBus(),
			qb:    system.QueryBus(),
		}, nil)

	system.Run(supervisor)
}

type HelloWorldApplicationSubsystem struct {
	clock mtime.Clock
	cb    misas.CommandBus
	qb    misas.QueryBus
}

func (h HelloWorldApplicationSubsystem) Name() string {
	return "hello_world"
}

func (h HelloWorldApplicationSubsystem) Initialize(context.Context) error { return nil }

func (h HelloWorldApplicationSubsystem) Teardown(context.Context) error { return nil }

func (h HelloWorldApplicationSubsystem) Run(ctx context.Context) error {
	//mx.Log(ctx).Info("Hello, World!")
	//mx.Log(ctx).Info("This is my first application subsystem under the mx framework!")
	//mx.Log(ctx).Info("The current time is " + h.clock.Now().String())
	//
	//mx.Log(ctx).Debug("Some debug information here.")

	result := h.cb.HandleCommand(ctx, SomeCommand{})
	if err, ok := result.Payload.(error); ok && err != nil {
		// Try a query
		queryResult := h.qb.HandleQuery(ctx, GetStockQuery{})
		_ = queryResult
		return err
	}
	return nil
}

type SomeCommand struct{}

func (c SomeCommand) TypeName() misas.CommandTypeName { return "some.command" }

type SomeEvent struct{}

func (e SomeEvent) TypeName() misas.EventTypeName { return "some.event" }

type GetStockQuery struct{}

func (q GetStockQuery) TypeName() misas.QueryTypeName { return "inventory.get_stock" }
