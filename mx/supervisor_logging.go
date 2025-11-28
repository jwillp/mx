package mx

import (
	"context"
	"fmt"
	"log/slog"
)

type supervisorLoggingPlugin struct{}

func (p supervisorLoggingPlugin) OnHook(ctx context.Context, hook PluginHook) error {
	switch e := hook.(type) {
	case SubsystemWillRestartHook:
		Log(ctx).Warn(
			fmt.Sprintf("restarting supervised application subsystem %q", e.ApplicationName),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Duration("restartDelay", e.RestartDelay),
			slog.Any(logKeyError, e.Error),
		)

	case SubsystemRestartedHook:
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

	case SubsystemMaxRestartReachedHook:
		Log(ctx).Error(
			fmt.Sprintf("giving up restart of supervised application subsystem %q: %s", e.ApplicationName, e.Reason),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Any(logKeyError, e.Error),
		)
	}

	return nil
}

func (supervisorLoggingPlugin) Name() string { return "supervisor.logging" }
