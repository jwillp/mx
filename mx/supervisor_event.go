package mx

import (
	"time"

	"github.com/morebec/misas/misas"
)

const (
	SubsystemWillRestartEventTypeName       misas.EventTypeName = "subsystem.will.restart"
	SubsystemRestartedEventTypeName         misas.EventTypeName = "subsystem.restarted"
	SubsystemMaxRestartReachedEventTypeName misas.EventTypeName = "subsystem.max.restart.reached"
)

type SubsystemWillRestartEvent struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	RestartDelay    time.Duration
	Error           error
	StartedAt       time.Time
}

func (e SubsystemWillRestartEvent) TypeName() misas.EventTypeName {
	return SubsystemWillRestartEventTypeName
}

type SubsystemRestartedEvent struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	Error           error
	StartedAt       time.Time
	EndedAt         time.Time
}

func (e SubsystemRestartedEvent) TypeName() misas.EventTypeName {
	return SubsystemRestartedEventTypeName
}

type SubsystemMaxRestartReachedEvent struct {
	ApplicationName string
	RestartCount    int
	MaxAttempts     int
	Reason          string
	Error           error
	ReachedAt       time.Time
}

func (e SubsystemMaxRestartReachedEvent) TypeName() misas.EventTypeName {
	return SubsystemMaxRestartReachedEventTypeName
}
