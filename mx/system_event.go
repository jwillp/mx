package mx

import (
	"context"
	"log/slog"
	"time"

	"github.com/morebec/misas/misas"
)

const (
	EventTypeSystemInitializationStarted misas.EventTypeName = "system.initialization.started"
	EventTypeSystemInitializationEnded   misas.EventTypeName = "system.initialization.ended"
	EventTypeSystemRunStarted            misas.EventTypeName = "system.run.started"
	EventTypeSystemRunEnded              misas.EventTypeName = "system.run.ended"
)

type SystemInitializationStartedEvent struct {
	Name        string
	Version     string
	StartedAt   time.Time
	Environment Environment
	Debug       bool
}

func (e SystemInitializationStartedEvent) TypeName() misas.EventTypeName {
	return EventTypeSystemInitializationStarted
}

type SystemInitializationEndedEvent struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemInitializationEndedEvent) TypeName() misas.EventTypeName {
	return EventTypeSystemInitializationEnded
}

type SystemRunStartedEvent struct {
	StartedAt time.Time
}

func (e SystemRunStartedEvent) TypeName() misas.EventTypeName { return EventTypeSystemRunStarted }

type SystemRunEndedEvent struct {
	StartedAt time.Time
	EndedAt   time.Time
	Error     error
}

func (e SystemRunEndedEvent) TypeName() misas.EventTypeName { return EventTypeSystemRunEnded }

type SubsystemInitializationStartedEvent struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemInitializationStartedEvent) TypeName() misas.EventTypeName {
	return misas.EventTypeInitializationStarted
}

type SubsystemInitializationEndedEvent struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemInitializationEndedEvent) TypeName() misas.EventTypeName {
	return misas.EventTypeInitializationEnded
}

type SubsystemRunStartedEvent struct {
	SubsystemName string
	StartedAt     time.Time
}

func (e SubsystemRunStartedEvent) TypeName() misas.EventTypeName { return misas.EventTypeRunStarted }

type SubsystemRunEndedEvent struct {
	SubsystemName string
	StartedAt     time.Time
	EndedAt       time.Time
	Error         error
}

func (e SubsystemRunEndedEvent) TypeName() misas.EventTypeName {
	return misas.EventTypeRunEnded
}

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
