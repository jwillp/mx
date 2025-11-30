package mx

import (
	"context"
	"log/slog"
)

const logKeySubsystem = "subsystem"
const logKeyError = "error"

func Log(ctx context.Context) ContextualLogger {
	logger := getLoggerFromContext(ctx)

	subsystemInfo := Ctx(ctx).SubsystemInfo()
	if (subsystemInfo != SubsystemInfo{}) {
		logger = logger.With(slog.String(logKeySubsystem, subsystemInfo.Name))
	}

	return ContextualLogger{ctx: ctx, logger: logger}
}

func getLoggerFromContext(ctx context.Context) *slog.Logger {
	if ctxLogger := ctx.Value(systemLoggerContextKey{}); ctxLogger != nil {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}

	return slog.Default()
}

type ContextualLogger struct {
	ctx    context.Context
	logger *slog.Logger
}

func (c ContextualLogger) Debug(msg string, args ...any) {
	c.logger.DebugContext(c.ctx, msg, args...)
}

func (c ContextualLogger) Info(msg string, args ...any) {
	c.logger.InfoContext(c.ctx, msg, args...)
}

func (c ContextualLogger) Warn(msg string, args ...any) {
	c.logger.WarnContext(c.ctx, msg, args...)
}

func (c ContextualLogger) Error(msg string, args ...any) {
	c.logger.ErrorContext(c.ctx, msg, args...)
}

func (c ContextualLogger) With(args ...any) ContextualLogger {
	return ContextualLogger{
		ctx:    c.ctx,
		logger: c.logger.With(args...),
	}
}
