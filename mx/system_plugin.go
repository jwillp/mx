package mx

import (
	"context"
	"log/slog"
	"time"
)

type SystemPluginHookName string

const (
	SystemInitializationStartedPluginHookName SystemPluginHookName = "system.initialization.started"
	SystemInitializationEndedPluginHookName   SystemPluginHookName = "system.initialization.ended"
	SystemExecutionStartedPluginHookName      SystemPluginHookName = "system.run.started"
	SystemExecutionEndedTypeName              SystemPluginHookName = "system.run.ended"
	SystemTeardownStartedPluginHookName       SystemPluginHookName = "system.teardown.started"
	SystemTeardownEndedPluginHookName         SystemPluginHookName = "system.teardown.ended"

	ApplicationSubsystemInitializationStartedPluginHookName SystemPluginHookName = "application_subsystem.initialization.started"
	ApplicationSubsystemInitializationEndedPluginHookName   SystemPluginHookName = "application_subsystem.initialization.ended"
	ApplicationSubsystemRunStartedPluginHookName            SystemPluginHookName = "application_subsystem.run.started"
	ApplicationSubsystemRunEndedPluginHookName              SystemPluginHookName = "application_subsystem.run.ended"
	ApplicationSubsystemTeardownStartedPluginHookName       SystemPluginHookName = "application_subsystem.teardown.started"
	ApplicationSubsystemTeardownEndedPluginHookName         SystemPluginHookName = "application_subsystem.teardown.ended"

	BusinessSubsystemInitializationStartedPluginHookName SystemPluginHookName = "business_subsystem.initialization.started"
	BusinessSubsystemInitializationEndedPluginHookName   SystemPluginHookName = "business_subsystem.initialization.ended"

	QuerySubsystemInitializationStartedPluginHookName SystemPluginHookName = "query_subsystem.initialization.started"
	QuerySubsystemInitializationEndedPluginHookName   SystemPluginHookName = "query_subsystem.initialization.ended"

	PluginAddedHookName SystemPluginHookName = "plugin.added"
)

type SystemInitializationStartedHook struct {
	Name        string
	Version     string
	StartedAt   time.Time
	Environment Environment
	Debug       bool
	System      *System
}

func (e SystemInitializationStartedHook) HookName() SystemPluginHookName {
	return SystemInitializationStartedPluginHookName
}

type SystemInitializationEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemInitializationEndedHook) HookName() SystemPluginHookName {
	return SystemInitializationEndedPluginHookName
}

type SystemExecutionStartedHook struct {
	StartedAt time.Time
}

func (e SystemExecutionStartedHook) HookName() SystemPluginHookName {
	return SystemExecutionStartedPluginHookName
}

type SystemExecutionEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemExecutionEndedHook) HookName() SystemPluginHookName {
	return SystemExecutionEndedTypeName
}

type SystemTeardownStartedHook struct {
	StartedAt time.Time
}

func (e SystemTeardownStartedHook) HookName() SystemPluginHookName {
	return SystemTeardownStartedPluginHookName
}

type SystemTeardownEndedHook struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemTeardownEndedHook) HookName() SystemPluginHookName {
	return SystemTeardownEndedPluginHookName
}

type ApplicationSubsystemInitializationStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemInitializationStartedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemInitializationStartedPluginHookName
}

type ApplicationSubsystemInitializationEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

func (e ApplicationSubsystemInitializationEndedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemInitializationEndedPluginHookName
}

type ApplicationSubsystemRunStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemRunStartedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemRunStartedPluginHookName
}

type ApplicationSubsystemRunEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

func (e ApplicationSubsystemRunEndedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemRunEndedPluginHookName
}

type ApplicationSubsystemTeardownStartedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
}

func (e ApplicationSubsystemTeardownStartedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemTeardownStartedPluginHookName
}

type ApplicationSubsystemTeardownEndedHook struct {
	ApplicationSubsystemName string
	StartedAt                time.Time
	EndedAt                  time.Time
	Error                    error
}

func (e ApplicationSubsystemTeardownEndedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemTeardownEndedPluginHookName
}

type BusinessSubsystemInitializationStartedHook struct {
	BusinessSubsystemName string
	StartedAt             time.Time
}

func (e BusinessSubsystemInitializationStartedHook) HookName() SystemPluginHookName {
	return BusinessSubsystemInitializationStartedPluginHookName
}

type BusinessSubsystemInitializationEndedHook struct {
	BusinessSubsystemName string
	StartedAt             time.Time
	EndedAt               time.Time
	Error                 error
}

func (e BusinessSubsystemInitializationEndedHook) HookName() SystemPluginHookName {
	return BusinessSubsystemInitializationEndedPluginHookName
}

type QuerySubsystemInitializationStartedHook struct {
	QuerySubsystemName string
	StartedAt          time.Time
}

func (e QuerySubsystemInitializationStartedHook) HookName() SystemPluginHookName {
	return QuerySubsystemInitializationStartedPluginHookName
}

type QuerySubsystemInitializationEndedHook struct {
	QuerySubsystemName string
	StartedAt          time.Time
	EndedAt            time.Time
	Error              error
}

func (e QuerySubsystemInitializationEndedHook) HookName() SystemPluginHookName {
	return QuerySubsystemInitializationEndedPluginHookName
}

type PluginAddedHook struct {
	PluginName string
}

func (e PluginAddedHook) HookName() SystemPluginHookName { return PluginAddedHookName }

type SystemPluginHook interface {
	HookName() SystemPluginHookName
}

type SystemPlugin interface {
	OnHook(context.Context, SystemPluginHook) error
	Name() string
}

type SystemPluginManager interface {
	DispatchHook(context.Context, SystemPluginHook)
	AddPlugin(context.Context, SystemPlugin)
}

type systemPluginManager struct {
	plugins []SystemPlugin
}

func newPluginManager() *systemPluginManager {
	return &systemPluginManager{}
}

func (pm *systemPluginManager) DispatchHook(ctx context.Context, hook SystemPluginHook) {
	// used an indexed for loop to allow plugins to add more plugins during execution
	for i := 0; i < len(pm.plugins); i++ {
		if err := pm.plugins[i].OnHook(ctx, hook); err != nil {
			// Log the error but continue executing other plugins
			Log(ctx).Error(
				"SystemPlugin hook execution failed",
				slog.String("plugin", pm.plugins[i].Name()),
				slog.String("hook", string(hook.HookName())),
				slog.Any(logKeyError, err),
			)
		}
	}
}

func (pm *systemPluginManager) AddPlugin(ctx context.Context, plugin SystemPlugin) {
	pm.plugins = append(pm.plugins, plugin)
	pm.DispatchHook(ctx, PluginAddedHook{PluginName: plugin.Name()})
}

// lateBindingSystemPluginManager is an implementation of SystemPluginManager that allows
// for late binding of the actual SystemPluginManager implementation. This is useful in scenarios
// where the SystemPluginManager needs to be referenced before it is fully initialized.
type lateBindingSystemPluginManager struct {
	*LateBinding[SystemPluginManager]
}

func newLateBindingSystemPluginManager() *lateBindingSystemPluginManager {
	return &lateBindingSystemPluginManager{
		LateBinding: NewLateBinding[SystemPluginManager](),
	}
}

func (h *lateBindingSystemPluginManager) AddPlugin(ctx context.Context, p SystemPlugin) {
	h.Get().AddPlugin(ctx, p)
}

func (h *lateBindingSystemPluginManager) DispatchHook(ctx context.Context, hook SystemPluginHook) {
	h.Get().DispatchHook(ctx, hook)
}
