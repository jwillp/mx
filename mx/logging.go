package mx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"
)

const logKeySubsystem = "subsystem"
const logKeyError = "error"

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

type loggingPlugin struct{}

func (hl loggingPlugin) OnHook(ctx context.Context, hook PluginHook) error {
	logger := Log(ctx)

	// Log hook dispatch at debug level for tracing
	logger.Debug(fmt.Sprintf("hook dispatched: %q", hook.HookName()))

	switch h := hook.(type) {
	case SystemInitializationStartedHook:
		currentUser, _ := user.Current()
		userID := "unknown"
		if currentUser != nil {
			userID = currentUser.Username
		}
		hostname, _ := os.Hostname()
		cwd, _ := os.Getwd()

		logger.Info(h.Name+" v"+h.Version,
			slog.String("environment", "development"),
			slog.String("host", hostname),
			slog.String("user", userID),
			slog.String("pid", fmt.Sprintf("%d", os.Getpid())),
			slog.String("cwd", cwd),
		)
		logger.Info("Mx Framework v" + Version)
		if h.Debug {
			logger.Debug("Debug mode is enabled")
			logger.Warn("SYSTEM IS IN DEBUG MODE, TURN OFF FOR PRODUCTION")
		}
		logger.Info("System initializing ...")
	case SystemInitializationEndedHook:
		if h.Error != nil {
			hl.logSystemError(ctx, h.Error)
			return nil
		}
		logger.Info("System initialized successfully")
	case SystemRunStartedHook:
		logger.Info("System running...")
	case SystemRunEndedHook:
		if h.Error != nil {
			hl.logSystemError(ctx, h.Error)
			return nil
		}
		logger.Info("System executed successfully")
	case SubsystemInitializationStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q initializing...", h.SubsystemName))
	case SubsystemInitializationEndedHook:
		if h.Error != nil {
			hl.logSubsystemError(ctx, h.SubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q initialized successfully", h.SubsystemName))
	case SubsystemRunStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q running...", h.SubsystemName))
	case SubsystemRunEndedHook:
		if h.Error != nil {
			hl.logSubsystemError(ctx, h.SubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q executed successfully", h.SubsystemName))
	case SystemTeardownStartedHook:
		logger.Info("System tearing down...")
	case SystemTeardownEndedHook:
		if h.Error != nil {
			hl.logSystemError(ctx, h.Error)
			return nil
		}
		logger.Info("System teardown completed successfully")
	case SubsystemTeardownStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q tearing down...", h.SubsystemName))
	case SubsystemTeardownEndedHook:
		if h.Error != nil {
			hl.logSubsystemError(ctx, h.SubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q teardown completed successfully", h.SubsystemName))

	case PluginAddedHook:
		logger.Debug(fmt.Sprintf("plugin %q added", h.PluginName))
	}

	return nil
}

func (loggingPlugin) Name() string { return "system.logging" }

func (hl loggingPlugin) logSystemError(ctx context.Context, err error) {
	Log(ctx).Error("System failed", slog.Any("error", err))
}

func (hl loggingPlugin) logSubsystemError(ctx context.Context, subsystemName string, err error) {
	Log(ctx).Error(fmt.Sprintf("application subsystem %q failed", subsystemName), slog.Any("error", err))
}
