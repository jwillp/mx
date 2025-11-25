package mx

import (
	"context"
	"fmt"
)

type ApplicationSubsystem interface {
	Name() string
	Init(context.Context) error
	Run(context.Context) error
}

// LoggingApplicationSubsystem is an application subsystem decorator that adds logging capabilities to an existing
// ApplicationSubsystem.
type loggingApplicationSubsystem struct {
	ApplicationSubsystem
}

func (s loggingApplicationSubsystem) Init(ctx context.Context) error {
	Log(ctx).Info(fmt.Sprintf("%s: initializing", s.Name()))
	err := s.ApplicationSubsystem.Init(ctx)
	if err != nil {
		Log(ctx).Error(fmt.Sprintf("%s: failed to initialize: %s", s.Name(), err.Error()))
		return err
	}

	Log(ctx).Info(fmt.Sprintf("%s: initialized successfully", s.Name()))
	return nil
}

func (s loggingApplicationSubsystem) Run(ctx context.Context) error {
	Log(ctx).Info(fmt.Sprintf("%s: running", s.Name()))
	err := s.ApplicationSubsystem.Run(ctx)
	if err != nil {
		Log(ctx).Error(fmt.Sprintf("%s: failed: %s", s.Name(), err.Error()))
		return err
	}

	Log(ctx).Info(fmt.Sprintf("%s: executed successfully", s.Name()))
	return nil
}
