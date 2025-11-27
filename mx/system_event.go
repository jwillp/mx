package mx

import (
	"context"
	"log/slog"
	"time"

	"github.com/morebec/misas/misas"
)

const (
	SystemInitializationStartedEventTypeName misas.EventTypeName = "system.initialization.started"
	SystemInitializationEndedEventTypeName   misas.EventTypeName = "system.initialization.ended"
	SystemRunStartedEventTypeName            misas.EventTypeName = "system.run.started"
	SystemRunEndedTypeName                   misas.EventTypeName = "system.run.ended"
	SystemTeardownStartedEventTypeName       misas.EventTypeName = "system.teardown.started"
	SystemTeardownEndedEventTypeName         misas.EventTypeName = "system.teardown.ended"

	SubsystemInitializationStartedEventTypeName misas.EventTypeName = "subsystem.initialization.started"
	SubsystemInitializationEndedEventTypeName   misas.EventTypeName = "subsystem.initialization.ended"
	SubsystemRunStartedEventTypeName            misas.EventTypeName = "subsystem.run.started"
	SubsystemRunEndedEventTypeName              misas.EventTypeName = "subsystem.run.ended"
	SubsystemTeardownStartedEventTypeName       misas.EventTypeName = "subsystem.teardown.started"
	SubsystemTeardownEndedEventTypeName         misas.EventTypeName = "subsystem.teardown.ended"
)

type SystemInitializationStartedEvent struct {
	Name        string
	Version     string
	StartedAt   time.Time
	Environment Environment
	Debug       bool
}

func (e SystemInitializationStartedEvent) TypeName() misas.EventTypeName {
	return SystemInitializationStartedEventTypeName
}

type SystemInitializationEndedEvent struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemInitializationEndedEvent) TypeName() misas.EventTypeName {
	return SystemInitializationEndedEventTypeName
}

type SystemRunStartedEvent struct {
	StartedAt time.Time
}

func (e SystemRunStartedEvent) TypeName() misas.EventTypeName { return SystemRunStartedEventTypeName }

type SystemRunEndedEvent struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemRunEndedEvent) TypeName() misas.EventTypeName { return SystemRunEndedTypeName }

type SystemTeardownStartedEvent struct {
	StartedAt time.Time
}

func (e SystemTeardownStartedEvent) TypeName() misas.EventTypeName {
	return SystemTeardownStartedEventTypeName
}

type SystemTeardownEndedEvent struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemTeardownEndedEvent) TypeName() misas.EventTypeName {
	return SystemTeardownEndedEventTypeName
}

type SubsystemInitializationStartedEvent struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemInitializationStartedEvent) TypeName() misas.EventTypeName {
	return SubsystemInitializationStartedEventTypeName
}

type SubsystemInitializationEndedEvent struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemInitializationEndedEvent) TypeName() misas.EventTypeName {
	return SubsystemInitializationEndedEventTypeName
}

type SubsystemRunStartedEvent struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemRunStartedEvent) TypeName() misas.EventTypeName {
	return SubsystemRunStartedEventTypeName
}

type SubsystemRunEndedEvent struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemRunEndedEvent) TypeName() misas.EventTypeName { return SubsystemRunEndedEventTypeName }

type SubsystemTeardownStartedEvent struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemTeardownStartedEvent) TypeName() misas.EventTypeName {
	return SubsystemTeardownStartedEventTypeName
}

type SubsystemTeardownEndedEvent struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemTeardownEndedEvent) TypeName() misas.EventTypeName {
	return SubsystemTeardownEndedEventTypeName
}

// systemEventBus is an internal event bus used to publish system events.
// It differs from regular event buses in that it handles errors internally by logging them
// instead of returning them to the caller.
type systemEventBus struct{ eb misas.EventBus }

func newSystemEventBus() systemEventBus {
	eventBus := misas.NewInMemoryEventBus()
	eventBus.RegisterHandler(loggingSystemEventHandler{})
	return systemEventBus{
		eb: eventBus,
	}
}

func (s systemEventBus) Publish(ctx context.Context, event misas.Event) {
	err := s.eb.Publish(ctx, event)
	if err != nil {
		Log(ctx).Error("an internal system event handler failed", slog.Any(logKeyError, err))
	}
}
