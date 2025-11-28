package mx

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const defaultTeardownTimeout = 30 * time.Second

type applicationSubsystemRegistration struct {
	app     ApplicationSubsystem
	options SupervisionOptions
}

type Supervisor struct {
	// Raw application registrations (stored before wrapping)
	rawApplications map[string]applicationSubsystemRegistration
	// Wrapped supervised applications (created during Initialize)
	supervisedApplications map[string]*supervisedApplicationSubsystem
	clock                  *HotSwappableClock
	pm                     *hotSwappableSystemPluginManager
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		rawApplications:        make(map[string]applicationSubsystemRegistration),
		supervisedApplications: make(map[string]*supervisedApplicationSubsystem),
		clock:                  NewHotSwappableClock(nil),
		pm:                     newHotSwappableSystemPluginManager(nil),
	}
}

func (s *Supervisor) Name() string { return "supervisor" }

func (s *Supervisor) WithApplicationSubsystem(app ApplicationSubsystem, options *SupervisionOptions) *Supervisor {
	// Store raw application registration without wrapping
	// Wrapping will happen during Initialize() when pm and clock are initialized
	s.rawApplications[app.Name()] = applicationSubsystemRegistration{
		app:     app,
		options: *options,
	}
	return s
}

func (s *Supervisor) OnHook(ctx context.Context, hook SystemPluginHook) error {
	if h, ok := hook.(SystemInitializationStartedHook); ok {
		s.pm.Swap(h.System.PluginManager())
		s.clock.Swap(h.System.Clock())
		s.pm.AddPlugin(ctx, supervisorLoggingPlugin{})
	}

	return nil
}

func (s *Supervisor) Initialize(ctx context.Context) error {
	// Wrap raw application subsystems with managed application subsystems now that pm and clock are initialized
	for name, reg := range s.rawApplications {
		supervisedApp := &supervisedApplicationSubsystem{
			ApplicationSubsystem: newManagedApplicationSubsystem(reg.app, s.pm, s.clock),
			Options:              reg.options,
			pm:                   s.pm,
		}
		s.supervisedApplications[name] = supervisedApp
	}

	Log(ctx).Debug("Initializing supervised applications...", slog.Int("nbApplications", len(s.supervisedApplications)))
	for _, app := range s.supervisedApplications {
		appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})
		if err := app.Initialize(appCtx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Supervisor) Run(ctx context.Context) error {
	for _, app := range s.supervisedApplications {
		go func(a *supervisedApplicationSubsystem) {
			// we can safely ignore the error here, as it can be captured through the system's event bus.
			_ = a.Run(ctx)
		}(app)
	}

	<-ctx.Done()
	err := ctx.Err()
	if err != nil {
		Log(ctx).Info("Supervisor received shutdown signal, stopping supervised applications...", slog.String("signal", err.Error()))
	}

	wg := sync.WaitGroup{}
	for _, app := range s.supervisedApplications {
		wg.Add(1)
		go func(a *supervisedApplicationSubsystem) {
			defer wg.Done()
			a.Terminate()
		}(app)
	}
	wg.Wait()

	return err
}

func (s *Supervisor) Teardown(ctx context.Context) error {
	// per-request timeout to avoid hangs
	ctx, cancel := context.WithTimeout(ctx, defaultTeardownTimeout)
	defer cancel()

	for _, app := range s.supervisedApplications {
		appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})

		// run teardown in a goroutine and wait for either completion or timeout
		done := make(chan error, 1)
		go func(a *supervisedApplicationSubsystem) {
			done <- a.Teardown(appCtx)
		}(app)

		select {
		case err := <-done:
			if err != nil {
				return err
			}
		case <-appCtx.Done():
			// If an application blocks teardown and the appCtx times out, panic to avoid infinite loops
			panic(fmt.Sprintf("teardown timeout for application %q: %v", app.Name(), appCtx.Err()))
		}
	}

	return nil
}
