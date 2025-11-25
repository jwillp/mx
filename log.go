package mx

import (
	"context"
	"github.com/lmittmann/tint"
	"hash/fnv"
	"log/slog"
	"os"
	"time"
)

const logKeySubsystem = "subsystem"

type systemLoggerKey struct{}
type applicationSubsystemLoggerKey struct{}

func Log(ctx context.Context) *slog.Logger {

	if ctxLogger := ctx.Value(applicationSubsystemLoggerKey{}); ctxLogger != nil {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}

	if ctxLogger := ctx.Value(systemLoggerKey{}); ctxLogger != nil {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}

	return slog.Default()
}

func (sc *SystemConf) newDefaultLoggerHandler() slog.Handler {
	switch sc.environment {
	case EnvironmentDevelopment:
		return tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Highlight error values in red.
				if a.Value.Kind() == slog.KindAny {
					if _, ok := a.Value.Any().(error); ok {
						return tint.Attr(9, a)
					}
				}
				// Highlight Subsystem key value
				if a.Key == logKeySubsystem {
					subsystemColor := colorFromSystemName(sc.name)
					return tint.Attr(subsystemColor, a)
				}
				return a
			},
		})
	default:
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
}

func colorFromSystemName(name string) uint8 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	hash := h.Sum32()

	colorIndex := int(hash%216) + 16
	return uint8(colorIndex)
}
