package mx

import (
	"time"
)

const (
	SubsystemWillRestartPluginHookName       PluginHookName = "subsystem.will.restart"
	SubsystemRestartedPluginHookName         PluginHookName = "subsystem.restarted"
	SubsystemMaxRestartReachedPluginHookName PluginHookName = "subsystem.max.restart.reached"
)

type SubsystemWillRestartHook struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	RestartDelay    time.Duration
	Error           error
	StartedAt       time.Time
}

func (e SubsystemWillRestartHook) HookName() PluginHookName {
	return SubsystemWillRestartPluginHookName
}

type SubsystemRestartedHook struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	Error           error
	StartedAt       time.Time
	EndedAt         time.Time
}

func (e SubsystemRestartedHook) HookName() PluginHookName {
	return SubsystemRestartedPluginHookName
}

type SubsystemMaxRestartReachedHook struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	Reason          string
	Error           error
	ReachedAt       time.Time
}

func (e SubsystemMaxRestartReachedHook) HookName() PluginHookName {
	return SubsystemMaxRestartReachedPluginHookName
}
