package mx

import (
	"github.com/morebec/misas/mtime"
	"github.com/samber/lo"
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
	name               string
	version            string
	environment        Environment
	debug              bool
	loggerHandler      slog.Handler
	clock              *DynamicBindingClock
	plugins            []SystemPlugin
	businessSubsystems map[string]BusinessSubsystemConf
	commandBus         *DynamicBindingCommandBus
	eventBuses         map[EventBusName]*DynamicBindingEventBus
	querySubsystems    map[string]QuerySubsystemConf
	queryBus           *DynamicBindingQueryBus
}

func NewSystem(name string) *SystemConf {
	return &SystemConf{
		name:        name,
		version:     "0.0.1",
		environment: defaultEnvironment,
		debug:       true,
		clock: func() *DynamicBindingClock {
			b := NewDynamicBindingClock()
			b.Bind(mtime.NewRealTimeClock(time.UTC))
			return b
		}(),
		businessSubsystems: make(map[string]BusinessSubsystemConf, 10),
		commandBus:         NewDynamicBindingCommandBus(),
		eventBuses:         make(map[EventBusName]*DynamicBindingEventBus, 10),
		querySubsystems:    make(map[string]QuerySubsystemConf, 10),
		queryBus:           NewDynamicBindingQueryBus(),
	}
}

func (sc *SystemConf) RunE(app ApplicationSubsystem) error {
	if sc.loggerHandler == nil {
		sc.loggerHandler = sc.newDefaultLoggerHandler()
	}

	sys := newSystem(sc)

	return sys.run(app)
}

func (sc *SystemConf) Run(as ApplicationSubsystem) {
	if err := sc.RunE(as); err != nil {
		panic(err)
	}
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

func (sc *SystemConf) WithClock(c mtime.Clock) *SystemConf {
	sc.clock.Bind(c)

	return sc
}

func (sc *SystemConf) Clock() mtime.Clock {
	return sc.clock
}

func (sc *SystemConf) WithBusinessSubsystem(bc *BusinessSubsystemConf) *SystemConf {
	sc.businessSubsystems[bc.name] = *bc

	return sc
}

func (sc *SystemConf) WithCommandBus(b misas.CommandBus) *SystemConf {
	sc.commandBus.Bind(b)

	return sc
}

func (sc *SystemConf) CommandBus() misas.CommandBus { return sc.commandBus }

func (sc *SystemConf) EventBus(s EventBusName) misas.EventBus {
	if _, exists := sc.eventBuses[s]; !exists {
		sc.eventBuses[s] = NewDynamicBindingEventBus()
	}

	return sc.eventBuses[s]
}

func (sc *SystemConf) WithQuerySubsystem(qc *QuerySubsystemConf) *SystemConf {
	sc.querySubsystems[qc.name] = *qc

	return sc
}

func (sc *SystemConf) QueryBus() misas.QueryBus { return sc.queryBus }

func (sc *SystemConf) WithPlugin(p SystemPlugin) *SystemConf {
	sc.plugins = append(sc.plugins, p)

	return sc
}

func (sc *SystemConf) newDefaultLoggerHandler() slog.Handler {
	switch sc.environment {
	case EnvironmentDevelopment:
		return NewHumanReadableLogHandler(os.Stdout, &slog.HandlerOptions{
			Level: lo.Ternary(!sc.debug, slog.LevelInfo, slog.LevelDebug),
		})
	default:
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lo.Ternary(sc.debug, slog.LevelDebug, slog.LevelInfo),
		})
	}
}
