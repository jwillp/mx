package mx

import (
	"context"
	"fmt"
	"log/slog"
)

type supervisorLoggingPlugin struct{}

func (p supervisorLoggingPlugin) OnHook(ctx context.Context, hook PluginHook) error {
	switch e := hook.(type) {
	case ApplicationSubsystemWillRestartHook:
		Log(ctx).Warn(
			fmt.Sprintf("restarting supervised application subsystem %q", e.ApplicationName),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Duration("restartDelay", e.RestartDelay),
			slog.Time("startedAt", e.StartedAt),
			slog.Int("failureCount", e.FailureCount),
			slog.Bool("circuitBreakerOpen", e.CircuitBreakerOpen),
			slog.Int("circuitBreakerThreshold", e.CircuitBreakerThreshold),
			slog.Any(logKeyError, e.Error),
		)

	case ApplicationSubsystemRestartedHook:
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

	case ApplicationSubsystemMaxRestartReachedHook:
		Log(ctx).Error(
			fmt.Sprintf("giving up restart of supervised application subsystem %q", e.ApplicationName),
			slog.String("reason", e.Reason),
			slog.Int("restartCount", e.RestartCount),
			slog.Int("maxAttempts", e.MaxAttempts),
			slog.Int("failureCount", e.FailureCount),
			slog.Bool("circuitBreakerOpen", e.CircuitBreakerOpen),
			slog.Int("circuitBreakerThreshold", e.CircuitBreakerThreshold),
			slog.Duration("circuitBreakerWindow", e.CircuitBreakerWindow),
			slog.Duration("maxRetryDuration", e.MaxRetryDuration),
			slog.Time("reachedAt", e.ReachedAt),
			slog.Any(logKeyError, e.Error),
		)
	}

	return nil
}

func (supervisorLoggingPlugin) Name() string { return "supervisor.logging" }
