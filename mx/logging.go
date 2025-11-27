package mx

import (
	"context"
	"fmt"
	"github.com/morebec/misas/misas"
	"log/slog"
	"os"
	"os/user"
)

const logKeySubsystem = "subsystem"
const logKeyError = "error"

// SystemInfo holds system-level metadata
type SystemInfo struct {
	Name        string
	Version     string
	Environment Environment
	Debug       bool
}

func Log(ctx context.Context) ContextualLogger {
	logger := getLoggerFromContext(ctx)

	subsystemInfo := GetSubsystemInfoFromContext(ctx)
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

func (c ContextualLogger) unwrap() *slog.Logger { return c.logger }

type loggingSystemEventHandler struct{}

func (h loggingSystemEventHandler) Handle(ctx context.Context, event misas.Event) error {
	logger := Log(ctx)

	switch e := event.(type) {
	case SystemInitializationStartedEvent:
		currentUser, _ := user.Current()
		userID := "unknown"
		if currentUser != nil {
			userID = currentUser.Username
		}
		hostname, _ := os.Hostname()
		cwd, _ := os.Getwd()

		logger.Info(e.Name+" v"+e.Version,
			slog.String("environment", "development"),
			slog.String("host", hostname),
			slog.String("user", userID),
			slog.String("pid", fmt.Sprintf("%d", os.Getpid())),
			slog.String("cwd", cwd),
		)
		logger.Info("Mx Framework v" + Version)
		if e.Debug {
			logger.Warn("SYSTEM IS IN DEBUG MODE, TURN OFF FOR PRODUCTION")
		}
		logger.Info("System initializing ...")
	case SystemInitializationEndedEvent:
		if e.Error != nil {
			h.logSystemError(ctx, e.Error)
			return nil
		}
		logger.Info("System initialized successfully")
	case SystemRunStartedEvent:
		logger.Info("System running...")
	case SystemRunEndedEvent:
		if e.Error != nil {
			h.logSystemError(ctx, e.Error)
			return nil
		}
		logger.Info("System executed successfully")
	case SubsystemInitializationStartedEvent:
		logger.Info(fmt.Sprintf("application subsystem %q initializing...", e.SubsystemName))
	case SubsystemInitializationEndedEvent:
		if e.Error != nil {
			h.logSubsystemError(ctx, e.SubsystemName, e.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q initialized successfully", e.SubsystemName))
	case SubsystemRunStartedEvent:
		logger.Info(fmt.Sprintf("application subsystem %q running...", e.SubsystemName))
	case SubsystemRunEndedEvent:
		if e.Error != nil {
			h.logSubsystemError(ctx, e.SubsystemName, e.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q executed successfully", e.SubsystemName))
	}

	return nil
}

func (h loggingSystemEventHandler) logSystemError(ctx context.Context, err error) {
	Log(ctx).Error("System failed", slog.Any("error", err))
}

func (h loggingSystemEventHandler) logSubsystemError(ctx context.Context, subsystemName string, err error) {
	Log(ctx).Error(fmt.Sprintf("application subsystem %q failed", subsystemName), slog.Any("error", err))
}
