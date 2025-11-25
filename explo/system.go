package explo

import (
	"context"
	"log/slog"
	"os"
)

const (
	EnvironmentProduction Environment = "production"
	Staging               Environment = "staging"
	Development           Environment = "development"
)

type Environment string

type System struct {
	name               string
	environment        Environment
	businessSubsystems map[string]*BusinessSubsystem
	logger             *slog.Logger
}

func NewSystem(name string) *System {
	return &System{
		name:        name,
		environment: Development,
		logger:      slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (s *System) WithEnvironment(env Environment) *System {
	s.environment = env

	return s
}

func (s *System) WithBusinessSubsystem(subsystem *BusinessSubsystem) *System {
	if s.businessSubsystems == nil {
		s.businessSubsystems = make(map[string]*BusinessSubsystem)
	}

	s.businessSubsystems[subsystem.Name] = subsystem

	return s
}

func (s *System) Run(app ApplicationSubsystem) error {
	app = NewLoggingApplicationSubsystem(app, s.logger)

	if err := app.Init(); err != nil {
		return err
	}

	return app.Run(context.Background())
}

func (s *System) Logger() *slog.Logger { return s.logger }
