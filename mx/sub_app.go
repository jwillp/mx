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
	pm    PluginManager
	clock misas.Clock
}

func newManagedApplicationSubsystem(app ApplicationSubsystem, pm PluginManager, clock misas.Clock) *managedApplicationSubsystem {
	return &managedApplicationSubsystem{
		ApplicationSubsystem: app,
		pm:                   pm,
		clock:                clock,
	}
}

func (s *managedApplicationSubsystem) Initialize(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, SubsystemInitializationStartedHook{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, SubsystemInitializationEndedHook{
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
	s.pm.DispatchHook(ctx, SubsystemRunStartedHook{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, SubsystemRunEndedHook{
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
	s.pm.DispatchHook(ctx, SubsystemTeardownStartedHook{SubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, SubsystemTeardownEndedHook{
			SubsystemName: s.Name(),
			StartedAt:     startedAt,
			EndedAt:       s.clock.Now(),
			Error:         err,
		})
	}()

	return s.ApplicationSubsystem.Teardown(ctx)
}
