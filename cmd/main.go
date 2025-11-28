package main

import (
	"context"
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

	supervisor := mx.NewSupervisor().
		WithApplicationSubsystem(HelloWorldApplicationSubsystem{
			clock: system.Clock(),
		}, nil)

	system.Run(supervisor)
}

type HelloWorldApplicationSubsystem struct {
	clock misas.Clock
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

	return fmt.Errorf("not implemented")
}
