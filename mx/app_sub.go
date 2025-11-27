package mx

import (
	"context"

	"github.com/morebec/misas/misas"
)

type ApplicationSubsystem interface {
	Name() string
	Initialize(context.Context) error
	Run(context.Context) error
	Teardown(context.Context) error
}

type managedApplicationSubsystem struct {
	ApplicationSubsystem
	eventBus systemEventBus
	clock    misas.Clock
}

func newManagedApplicationSubsystem(as ApplicationSubsystem, eventBus systemEventBus, clock misas.Clock) *managedApplicationSubsystem {
	return &managedApplicationSubsystem{
		ApplicationSubsystem: as,
		eventBus:             eventBus,
		clock:                clock,
	}
}

func (s *managedApplicationSubsystem) Initialize(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.eventBus.Publish(ctx, SubsystemInitializationStartedEvent{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.eventBus.Publish(ctx, SubsystemInitializationEndedEvent{
			SubsystemName: s.Name(),
			StartedAt:     startedAt,
			EndedAt:       s.clock.Now(),
			Error:         err,
		})
	}()

	return s.ApplicationSubsystem.Initialize(ctx)
}

func (s *managedApplicationSubsystem) Run(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.eventBus.Publish(ctx, SubsystemRunStartedEvent{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.eventBus.Publish(ctx, SubsystemRunEndedEvent{
			SubsystemName: s.Name(),
			StartedAt:     startedAt,
			EndedAt:       s.clock.Now(),
			Error:         err,
		})
	}()

	return s.ApplicationSubsystem.Run(ctx)
}

func (s *managedApplicationSubsystem) Teardown(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.eventBus.Publish(ctx, SubsystemTeardownStartedEvent{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.eventBus.Publish(ctx, SubsystemTeardownEndedEvent{
			SubsystemName: s.Name(),
			StartedAt:     startedAt,
			EndedAt:       s.clock.Now(),
			Error:         err,
		})
	}()

	return s.ApplicationSubsystem.Teardown(ctx)
}
