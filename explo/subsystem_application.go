package explo

import (
	"context"
	"fmt"
	"log/slog"
)

type ApplicationSubsystem interface {
	Name() string
	Initialize() error
	Run(ctx context.Context) error
	Teardown(ctx context.Context) error
}

// LoggingApplicationSubsystem is an application subsystem decorator that adds logging capabilities to an existing
// ApplicationSubsystem.
type LoggingApplicationSubsystem struct {
	ApplicationSubsystem
	logger *slog.Logger
}

func NewLoggingApplicationSubsystem(
	app ApplicationSubsystem,
	logger *slog.Logger,
) LoggingApplicationSubsystem {
	return LoggingApplicationSubsystem{
		ApplicationSubsystem: app,
		logger:               logger.With(slog.String("subsystem", app.Name())),
	}
}

func (s LoggingApplicationSubsystem) Initialize() error {
	s.logger.Info(fmt.Sprintf("%s: initializing", s.Name()))
	err := s.ApplicationSubsystem.Initialize()
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s: failed to initialize: %s", s.Name(), err.Error()))
		return err
	}

	s.logger.Info(fmt.Sprintf("%s: initialized successfully", s.Name()))
	return nil
}

func (s LoggingApplicationSubsystem) Run(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("%s: running", s.Name()))
	err := s.ApplicationSubsystem.Run(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s: failed: %s", s.Name(), err.Error()))
		return err
	}

	s.logger.Info(fmt.Sprintf("%s: executed successfully", s.Name()))
	return nil
}

func (s LoggingApplicationSubsystem) Teardown(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("%s: tearing down", s.Name()))
	err := s.ApplicationSubsystem.Teardown(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("%s: failed to teardown: %s", s.Name(), err.Error()))
		return err
	}

	s.logger.Info(fmt.Sprintf("%s: teardown completed successfully", s.Name()))
	return nil
}
