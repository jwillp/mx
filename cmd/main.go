package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/morebec/misas/misas"
	"github.com/morebec/misas/mx"
)

func main() {
	system := mx.NewSystem("MyApp").
		//WithEnvironment(mx.EnvironmentProduction).
		WithEnvironment(mx.EnvironmentDevelopment).
		WithDebug(false).
		WithClock(misas.NewRealTimeClock(time.UTC))

	system.WithBusinessSubsystem(
		mx.NewBusinessSubsystem("inventory").
			WithCommandHandler("some.command", misas.CommandHandlerFunc(func(ctx context.Context, cmd misas.Command) misas.CommandResult {
				return misas.CommandResult{
					Payload: fmt.Errorf("some command failed: inventory is out of stock"),
					//Payload: fmt.Errorf("some command failed: failed to publish event: %w", inventoryEventBus.Publish(ctx, SomeEvent{})),
				}
			})).
			WithEventHandlers("inventory", misas.EventHandlerFunc(func(ctx context.Context, event misas.Event) error {
				return errors.New("inventory event handler failed")
			})),
	)

	supervisor := mx.NewSupervisor().
		WithApplicationSubsystem(HelloWorldApplicationSubsystem{
			clock: system.Clock(),
			cb:    system.CommandBus(),
		}, nil)

	system.Run(supervisor)
}

type HelloWorldApplicationSubsystem struct {
	clock misas.Clock
	cb    misas.CommandBus
	eb    misas.EventBus
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
	return result.Payload.(error)
}

type SomeCommand struct{}

func (c SomeCommand) TypeName() misas.CommandTypeName { return "some.command" }

type SomeEvent struct{}

func (e SomeEvent) TypeName() misas.EventTypeName { return "some.event" }
