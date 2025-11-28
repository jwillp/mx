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

	SubsystemInitializationStartedPluginHookName PluginHookName = "subsystem.initialization.started"
	SubsystemInitializationEndedPluginHookName   PluginHookName = "subsystem.initialization.ended"
	SubsystemRunStartedPluginHookName            PluginHookName = "subsystem.run.started"
	SubsystemRunEndedPluginHookName              PluginHookName = "subsystem.run.ended"
	SubsystemTeardownStartedPluginHookName       PluginHookName = "subsystem.teardown.started"
	SubsystemTeardownEndedPluginHookName         PluginHookName = "subsystem.teardown.ended"

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

type SubsystemInitializationStartedHook struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemInitializationStartedHook) HookName() PluginHookName {
	return SubsystemInitializationStartedPluginHookName
}

type SubsystemInitializationEndedHook struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemInitializationEndedHook) HookName() PluginHookName {
	return SubsystemInitializationEndedPluginHookName
}

type SubsystemRunStartedHook struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemRunStartedHook) HookName() PluginHookName {
	return SubsystemRunStartedPluginHookName
}

type SubsystemRunEndedHook struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemRunEndedHook) HookName() PluginHookName { return SubsystemRunEndedPluginHookName }

type SubsystemTeardownStartedHook struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemTeardownStartedHook) HookName() PluginHookName {
	return SubsystemTeardownStartedPluginHookName
}

type SubsystemTeardownEndedHook struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

type PluginAddedHook struct {
	PluginName string
}

func (e PluginAddedHook) HookName() PluginHookName { return PluginAddedHookName }

func (e SubsystemTeardownEndedHook) HookName() PluginHookName {
	return SubsystemTeardownEndedPluginHookName
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
