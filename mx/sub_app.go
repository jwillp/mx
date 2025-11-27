package mx

import (
	"context"

	"github.com/morebec/misas/misas"
)

type ApplicationSubsystem interface {
	// Initialize is called to set up the subsystem before running.
	Initialize(context.Context) error

	// Run is called to start the subsystem's operations.
	Run(context.Context) error

	// Teardown is called to clean up resources used by the subsystem when shutting down.
	Teardown(context.Context) error

	// Name returns the unique name of the subsystem.
	Name() string
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
