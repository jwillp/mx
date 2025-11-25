package mx

import (
	"context"
	"fmt"
	"log/slog"
)

type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentDevelopment Environment = "development"
	defaultEnvironment                 = EnvironmentDevelopment
)

type SystemConf struct {
	name          string
	environment   Environment
	loggerHandler slog.Handler
}

func NewSystem(name string) *SystemConf {
	return &SystemConf{
		name:        name,
		environment: defaultEnvironment,
	}
}

func (sc *SystemConf) Run(as ApplicationSubsystem) error {
	if sc.loggerHandler == nil {
		sc.loggerHandler = sc.newDefaultLoggerHandler()
	}

	s := system{
		applicationSubsystem: as,
		logger:               slog.New(sc.loggerHandler),
	}

	return s.run()
}

func (sc *SystemConf) WithEnvironment(env Environment) *SystemConf {
	sc.environment = env

	return sc
}

type system struct {
	applicationSubsystem ApplicationSubsystem
	logger               *slog.Logger
}

func (s *system) run() error {
	ctx := s.newSystemContext()

	Log(ctx).Info("Initializing system...")
	s.applicationSubsystem = loggingApplicationSubsystem{
		ApplicationSubsystem: s.applicationSubsystem,
	}

	appCtx := newApplicationContext(ctx, s.applicationSubsystem)
	if err := s.applicationSubsystem.Init(appCtx); err != nil {
		Log(appCtx).Error("failed to initialize system: "+err.Error(), slog.Any("error", err))
		return fmt.Errorf("failed to initialize system: %w", err)
	}
	Log(ctx).Info("System initialized successfully")

	Log(ctx).Info("Running system")

	return s.applicationSubsystem.Run(ctx)
}

func (s *system) newSystemContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, systemLoggerKey{}, s.logger)

	return ctx
}

func newApplicationContext(ctx context.Context, as ApplicationSubsystem) context.Context {
	ctx = context.WithValue(ctx,
		applicationSubsystemLoggerKey{},
		Log(ctx).With(slog.String(logKeySubsystem, as.Name())),
	)

	return ctx
}
