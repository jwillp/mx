package mx

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/morebec/misas/misas"
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
	version       string
	environment   Environment
	debug         bool
	loggerHandler slog.Handler
	clock         *HotSwappableClock
}

func NewSystem(name string) *SystemConf {
	return &SystemConf{
		name:        name,
		version:     "0.0.1",
		environment: defaultEnvironment,
		debug:       true,
		clock:       NewHotSwappableClock(misas.NewRealTimeClock(time.UTC)),
	}
}

func (sc *SystemConf) Run(as ApplicationSubsystem) error {
	if sc.loggerHandler == nil {
		sc.loggerHandler = sc.newDefaultLoggerHandler()
	}

	sys := newSystem(*sc)
	return sys.run(as)
}

func (sc *SystemConf) WithEnvironment(env Environment) *SystemConf {
	sc.environment = env

	return sc
}

func (sc *SystemConf) WithVersion(version string) *SystemConf {
	sc.version = version

	return sc
}

func (sc *SystemConf) WithDebug(debug bool) *SystemConf {
	sc.debug = debug

	return sc
}

func (sc *SystemConf) WithClock(c misas.Clock) *SystemConf {
	sc.clock.Swap(c)

	return sc
}

func (sc *SystemConf) Clock() misas.Clock {
	return sc.clock
}

func (sc *SystemConf) newDefaultLoggerHandler() slog.Handler {
	switch sc.environment {
	case EnvironmentDevelopment:
		return NewHumanReadableLogHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	default:
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
}

type system struct {
	info     SystemInfo
	logger   *slog.Logger
	clock    misas.Clock
	eventBus systemEventBus
}

func newSystem(sc SystemConf) *system {
	return &system{
		info: SystemInfo{
			Name:        sc.name,
			Version:     sc.version,
			Environment: sc.environment,
			Debug:       sc.debug,
		},
		clock:    sc.clock,
		logger:   slog.New(sc.loggerHandler),
		eventBus: newSystemEventBus(),
	}
}

func (s *system) run(app ApplicationSubsystem) (err error) {
	ctx := newSystemContext(*s)

	app = newManagedApplicationSubsystem(app, s.eventBus, s.clock)
	appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})

	// INIT
	initializationStartedAt := s.clock.Now()
	s.eventBus.Publish(ctx, SystemInitializationStartedEvent{
		Name:        s.info.Name,
		Version:     s.info.Version,
		Environment: s.info.Environment,
		Debug:       s.info.Debug,
		StartedAt:   s.clock.Now(),
	})
	err = app.Init(appCtx)
	s.eventBus.Publish(ctx, SystemInitializationEndedEvent{
		StartedAt: initializationStartedAt,
		EndedAt:   s.clock.Now(),
		Error:     err,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize system: %w", err)
	}

	// RUN
	runStartedAt := s.clock.Now()
	s.eventBus.Publish(ctx, SystemRunStartedEvent{StartedAt: s.clock.Now()})
	err = app.Run(appCtx)
	s.eventBus.Publish(ctx, SystemRunEndedEvent{StartedAt: runStartedAt, EndedAt: s.clock.Now(), Error: err})

	return err
}

type SubsystemInfo struct {
	Name string
}
