package mx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"
)

type loggingPlugin struct{}

func (hl loggingPlugin) OnHook(ctx context.Context, hook SystemPluginHook) error {
	logger := Log(ctx)

	// Log hook dispatch at debug level for tracing
	logger.Debug(fmt.Sprintf("hook dispatched: %q", hook.HookName()))

	switch h := hook.(type) {
	case SystemInitializationStartedHook:
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
	case ApplicationSubsystemInitializationStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q initializing...", h.ApplicationSubsystemName))
	case ApplicationSubsystemInitializationEndedHook:
		if h.Error != nil {
			hl.logApplicationSubsystemError(ctx, h.ApplicationSubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q initialized successfully", h.ApplicationSubsystemName))
	case ApplicationSubsystemRunStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q running...", h.ApplicationSubsystemName))
	case ApplicationSubsystemRunEndedHook:
		if h.Error != nil {
			hl.logApplicationSubsystemError(ctx, h.ApplicationSubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q executed successfully", h.ApplicationSubsystemName))
	case SystemTeardownStartedHook:
		logger.Info("System tearing down...")
	case SystemTeardownEndedHook:
		if h.Error != nil {
			hl.logSystemError(ctx, h.Error)
			return nil
		}
		logger.Info("System teardown completed successfully")
	case ApplicationSubsystemTeardownStartedHook:
		logger.Info(fmt.Sprintf("application subsystem %q tearing down...", h.ApplicationSubsystemName))
	case ApplicationSubsystemTeardownEndedHook:
		if h.Error != nil {
			hl.logApplicationSubsystemError(ctx, h.ApplicationSubsystemName, h.Error)
			return nil
		}
		logger.Info(fmt.Sprintf("application subsystem %q teardown completed successfully", h.ApplicationSubsystemName))

	case PluginAddedHook:
		// Display banner when logging plugin is added (it's always first)
		if h.PluginName == "system.logging" {
			hl.logSystemInfoBanner(ctx)
		}

		logger.Debug(fmt.Sprintf("plugin %q added", h.PluginName))
	}

	return nil
}

func (loggingPlugin) Name() string { return "system.logging" }

func (hl loggingPlugin) logSystemInfoBanner(ctx context.Context) {
	logger := Log(ctx)
	systemInfo := GetSystemInfoFromContext(ctx)
	currentUser, _ := user.Current()
	userID := "unknown"
	if currentUser != nil {
		userID = currentUser.Username
	}
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()

	logger.Info(systemInfo.Name+" v"+systemInfo.Version,
		slog.String("environment", string(systemInfo.Environment)),
		slog.String("host", hostname),
		slog.String("user", userID),
		slog.String("pid", fmt.Sprintf("%d", os.Getpid())),
		slog.String("cwd", cwd),
	)
	logger.Info("Mx Framework v" + Version)
	if systemInfo.Debug {
		logger.Debug("Debug mode is enabled")
		logger.Warn("SYSTEM IS IN DEBUG MODE, TURN OFF FOR PRODUCTION")
	}
}

func (hl loggingPlugin) logSystemError(ctx context.Context, err error) {
	Log(ctx).Error("System failed", slog.Any("error", err))
}

func (hl loggingPlugin) logApplicationSubsystemError(ctx context.Context, applicationSubsystemName string, err error) {
	Log(ctx).Error(fmt.Sprintf("application subsystem %q failed", applicationSubsystemName), slog.Any("error", err))
}
