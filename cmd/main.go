package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/morebec/misas"
)

func main() {
	system := mx.NewSystem("MyApp").
		//WithEnvironment(mx.EnvironmentProduction)
		WithEnvironment(mx.EnvironmentDevelopment)

	if err := system.Run(HelloWorldApplicationSubsystem{}); err != nil {
		panic(err)
	}
}

type HelloWorldApplicationSubsystem struct{}

func (h HelloWorldApplicationSubsystem) Name() string {
	return "hello_world"
}

func (h HelloWorldApplicationSubsystem) Init(context.Context) error { return errors.New("FAILED !") }

func (h HelloWorldApplicationSubsystem) Run(ctx context.Context) error {
	mx.Log(ctx).Info("Hello, World!")

	return fmt.Errorf("OH NO!")
}
