package mx

import (
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
	name          string
	version       string
	environment   Environment
	debug         bool
	loggerHandler slog.Handler
	clock         *HotSwappableClock
	plugins       []SystemPlugin
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
