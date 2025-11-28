package mx

import (
	"time"
)

const (
	ApplicationSubsystemWillRestartPluginHookName       SystemPluginHookName = "application_subsystem.will.restart"
	ApplicationSubsystemRestartedPluginHookName         SystemPluginHookName = "application_subsystem.restarted"
	ApplicationSubsystemMaxRestartReachedPluginHookName SystemPluginHookName = "application_subsystem.max.restart.reached"
)

type ApplicationSubsystemWillRestartHook struct {
	ApplicationName         string
	RestartCount            int
	MaxAttempts             int
	RestartDelay            time.Duration
	Error                   error
	StartedAt               time.Time
	FailureCount            int  // Failures in current window
	CircuitBreakerOpen      bool // Is circuit breaker currently open
	CircuitBreakerThreshold int  // Threshold for rapid-fire detection
}

func (e ApplicationSubsystemWillRestartHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemWillRestartPluginHookName
}

type ApplicationSubsystemRestartedHook struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	Error           error
	StartedAt       time.Time
	EndedAt         time.Time
}

func (e ApplicationSubsystemRestartedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemRestartedPluginHookName
}

type ApplicationSubsystemMaxRestartReachedHook struct {
	ApplicationName         string
	RestartCount            int    // Actual restart attempts made
	MaxAttempts             int    // Configured max retry limit
	Reason                  string // Why restart stopped
	Error                   error
	ReachedAt               time.Time
	FailureCount            int           // Failures when limit was reached
	CircuitBreakerOpen      bool          // Is circuit breaker the cause
	CircuitBreakerThreshold int           // Rapid-fire failure threshold
	CircuitBreakerWindow    time.Duration // Time window for circuit breaker
}

func (e ApplicationSubsystemMaxRestartReachedHook) HookName() SystemPluginHookName {
	return ApplicationSubsystemMaxRestartReachedPluginHookName
}
