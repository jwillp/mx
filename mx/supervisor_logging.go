package mx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/morebec/misas/misas"
)

type supervisorLoggingEventHandler struct{}

func (h supervisorLoggingEventHandler) Handle(ctx context.Context, event misas.Event) error {
	switch e := event.(type) {
	case SubsystemWillRestartEvent:
		Log(ctx).Warn(
			fmt.Sprintf("restarting supervised application subsystem %q", e.ApplicationName),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Duration("restartDelay", e.RestartDelay),
			slog.Any(logKeyError, e.Error),
		)

	case SubsystemRestartedEvent:
		if e.Error == nil {
			Log(ctx).Info(
				fmt.Sprintf("restarted supervised application subsystem %q", e.ApplicationName),
				slog.Int("restartCount", e.RestartCount),
				slog.Int("maxAttempts", e.MaxAttempts),
			)
		} else {
			Log(ctx).Error(
				fmt.Sprintf("restart failed for supervised application subsystem %q", e.ApplicationName),
				slog.Int("restartCount", e.RestartCount),
				slog.Int("maxAttempts", e.MaxAttempts),
				slog.Any(logKeyError, e.Error),
			)
		}

	case SubsystemMaxRestartReachedEvent:
		Log(ctx).Error(
			fmt.Sprintf("giving up restart of supervised application subsystem %q: %s", e.ApplicationName, e.Reason),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Any(logKeyError, e.Error),
		)
	}

	return nil
}
