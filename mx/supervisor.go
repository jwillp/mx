package mx

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const defaultTeardownTimeout = 5 * time.Second

type Supervisor struct {
	supervisedApplications map[string]*supervisedApplicationSubsystem
	clock                  *HotSwappableClock
	eventBus               *hotSwappableSystemEventBus
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		supervisedApplications: make(map[string]*supervisedApplicationSubsystem),
		clock:                  NewHotSwappableClock(nil),
		eventBus:               newHotSwappableSystemEventBus(nil),
	}
}

func (s *Supervisor) Name() string { return "supervisor" }

func (s *Supervisor) WithApplicationSubsystem(app ApplicationSubsystem, options *SupervisionOptions) *Supervisor {
	supervisedApp := &supervisedApplicationSubsystem{
		ApplicationSubsystem: newManagedApplicationSubsystem(app, systemEventBus(s.eventBus), s.clock),
		Options:              *options,
		eventBus:             systemEventBus(s.eventBus),
	}
	s.supervisedApplications[app.Name()] = supervisedApp
	return s
}

func (s *Supervisor) Initialize(ctx context.Context) error {
	s.eventBus.RegisterHandler(supervisorLoggingEventHandler{})

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
