package mx

import (
	"context"

	"github.com/morebec/misas/misas"
)

type ApplicationSubsystem interface {
	// Initialize is called to set up the application subsystem before running.
	Initialize(context.Context) error

	// Run is called to start the application subsystem's operations.
	Run(context.Context) error

	// Teardown is called to clean up resources used by the application subsystem when shutting down.
	Teardown(context.Context) error

	// Name returns the unique name of the application subsystem.
	Name() string
}

type managedApplicationSubsystem struct {
	ApplicationSubsystem
	pm    SystemPluginManager
	clock misas.Clock
}

func newManagedApplicationSubsystem(app ApplicationSubsystem, pm SystemPluginManager, clock misas.Clock) *managedApplicationSubsystem {
	return &managedApplicationSubsystem{
		ApplicationSubsystem: app,
		pm:                   pm,
		clock:                clock,
	}
}

func (s *managedApplicationSubsystem) Initialize(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, ApplicationSubsystemInitializationStartedHook{ApplicationSubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, ApplicationSubsystemInitializationEndedHook{
			ApplicationSubsystemName: s.Name(),
			StartedAt:                startedAt,
			EndedAt:                  s.clock.Now(),
			Error:                    err,
		})
	}()

	return s.ApplicationSubsystem.Initialize(ctx)
}

func (s *managedApplicationSubsystem) Run(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, ApplicationSubsystemRunStartedHook{ApplicationSubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, ApplicationSubsystemRunEndedHook{
			ApplicationSubsystemName: s.Name(),
			StartedAt:                startedAt,
			EndedAt:                  s.clock.Now(),
			Error:                    err,
		})
	}()

	return s.ApplicationSubsystem.Run(ctx)
}

func (s *managedApplicationSubsystem) Teardown(ctx context.Context) (err error) {
	startedAt := s.clock.Now()
	s.pm.DispatchHook(ctx, ApplicationSubsystemTeardownStartedHook{ApplicationSubsystemName: s.Name(), StartedAt: startedAt})
	defer func() {
		s.pm.DispatchHook(ctx, ApplicationSubsystemTeardownEndedHook{
			ApplicationSubsystemName: s.Name(),
			StartedAt:                startedAt,
			EndedAt:                  s.clock.Now(),
			Error:                    err,
		})
	}()

	return s.ApplicationSubsystem.Teardown(ctx)
}
