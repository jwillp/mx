package mx

import (
	"context"
	"log/slog"
	"time"
)

type PluginHookName string

const (
	SystemInitializationStartedPluginHookName PluginHookName = "system.initialization.started"
	SystemInitializationEndedPluginHookName   PluginHookName = "system.initialization.ended"
	SystemRunStartedPluginHookName            PluginHookName = "system.run.started"
	SystemRunEndedTypeName                    PluginHookName = "system.run.ended"
	SystemTeardownStartedPluginHookName       PluginHookName = "system.teardown.started"
	SystemTeardownEndedPluginHookName         PluginHookName = "system.teardown.ended"

	ApplicationSubsystemInitializationStartedPluginHookName PluginHookName = "application_subsystem.initialization.started"
	ApplicationSubsystemInitializationEndedPluginHookName   PluginHookName = "application_subsystem.initialization.ended"
	ApplicationSubsystemRunStartedPluginHookName            PluginHookName = "application_subsystem.run.started"
	ApplicationSubsystemRunEndedPluginHookName              PluginHookName = "application_subsystem.run.ended"
	ApplicationSubsystemTeardownStartedPluginHookName       PluginHookName = "application_subsystem.teardown.started"
	ApplicationSubsystemTeardownEndedPluginHookName         PluginHookName = "application_subsystem.teardown.ended"

	PluginAddedHookName PluginHookName = "plugin.added"
)

type SystemInitializationStartedHook struct {
	Name        string
	Version     string
	StartedAt   time.Time
	Environment Environment
	Debug       bool
	System      *System
}

func (e SystemInitializationStartedHook) HookName() PluginHookName {
	return SystemInitializationStartedPluginHookName
}

type SystemInitializationEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemInitializationEndedHook) HookName() PluginHookName {
	return SystemInitializationEndedPluginHookName
}

type SystemRunStartedHook struct {
	StartedAt time.Time
}

func (e SystemRunStartedHook) HookName() PluginHookName { return SystemRunStartedPluginHookName }

type SystemRunEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemRunEndedHook) HookName() PluginHookName { return SystemRunEndedTypeName }

type SystemTeardownStartedHook struct {
	StartedAt time.Time
}

func (e SystemTeardownStartedHook) HookName() PluginHookName {
	return SystemTeardownStartedPluginHookName
}

type SystemTeardownEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemTeardownEndedHook) HookName() PluginHookName {
	return SystemTeardownEndedPluginHookName
}

type ApplicationSubsystemInitializationStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemInitializationStartedHook) HookName() PluginHookName {
	return ApplicationSubsystemInitializationStartedPluginHookName
}

type ApplicationSubsystemInitializationEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

func (e ApplicationSubsystemInitializationEndedHook) HookName() PluginHookName {
	return ApplicationSubsystemInitializationEndedPluginHookName
}

type ApplicationSubsystemRunStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemRunStartedHook) HookName() PluginHookName {
	return ApplicationSubsystemRunStartedPluginHookName
}

type ApplicationSubsystemRunEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

func (e ApplicationSubsystemRunEndedHook) HookName() PluginHookName {
	return ApplicationSubsystemRunEndedPluginHookName
}

type ApplicationSubsystemTeardownStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemTeardownStartedHook) HookName() PluginHookName {
	return ApplicationSubsystemTeardownStartedPluginHookName
}

type ApplicationSubsystemTeardownEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

type PluginAddedHook struct {
	PluginName string
}

func (e PluginAddedHook) HookName() PluginHookName { return PluginAddedHookName }

func (e ApplicationSubsystemTeardownEndedHook) HookName() PluginHookName {
	return ApplicationSubsystemTeardownEndedPluginHookName
}

type PluginHook interface {
	HookName() PluginHookName
}

type Plugin interface {
	OnHook(context.Context, PluginHook) error
	Name() string
}

type PluginManager interface {
	DispatchHook(context.Context, PluginHook)
	AddPlugin(context.Context, Plugin)
}

type pluginManager struct {
	plugins []Plugin
}

func newPluginManager() *pluginManager {
	return &pluginManager{}
}

func (pm *pluginManager) DispatchHook(ctx context.Context, hook PluginHook) {
	// used an indexed for loop to allow plugins to add more plugins during execution
	for i := 0; i < len(pm.plugins); i++ {
		if err := pm.plugins[i].OnHook(ctx, hook); err != nil {
			// Log the error but continue executing other plugins
			Log(ctx).Error(
				"Plugin hook execution failed",
				slog.String("plugin", pm.plugins[i].Name()),
				slog.String("hook", string(hook.HookName())),
				slog.Any(logKeyError, err),
			)
		}
	}
}

func (pm *pluginManager) AddPlugin(ctx context.Context, plugin Plugin) {
	pm.plugins = append(pm.plugins, plugin)
	pm.DispatchHook(ctx, PluginAddedHook{PluginName: plugin.Name()})
}
