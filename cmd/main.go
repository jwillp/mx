package main

import (
	"context"
	"fmt"
	"github.com/morebec/misas/misas"
	"github.com/morebec/misas/mx"
	"time"
)

func main() {
	system := mx.NewSystem("MyApp").
		//WithEnvironment(mx.EnvironmentProduction).
		WithEnvironment(mx.EnvironmentDevelopment).
		WithClock(misas.NewRealTimeClock(time.UTC))

	supervisor := mx.NewSupervisor().
		WithApplicationSubsystem(HelloWorldApplicationSubsystem{
			clock: system.Clock(),
		}, &mx.SupervisionOptions{
			RestartPolicy: mx.DefaultRestartPolicy,
		})

	if err := system.Run(supervisor); err != nil {
		panic(err)
	}
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
	mx.Log(ctx).Info("Hello, World!")
	mx.Log(ctx).Info("This is my first application subsystem under the mx framework!")
	mx.Log(ctx).Info("The current time is " + h.clock.Now().String())

	mx.Log(ctx).Debug("Some debug information here.")

	return fmt.Errorf("not implemented")
}
