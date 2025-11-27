package mx

import (
	"context"
	"errors"
	"github.com/morebec/misas/misas"
	"github.com/samber/lo"
	"log/slog"
	"sync"
)

type ApplicationSubsystemState string

const (
	ApplicationSubsystemStateStopped ApplicationSubsystemState = "STOPPED"
	ApplicationSubsystemStateRunning ApplicationSubsystemState = "RUNNING"
	ApplicationSubsystemStateFailed  ApplicationSubsystemState = "FAILED"
)

// SupervisedApplicationSubsystemStatus represents the complete status of a supervised application subsystem
type SupervisedApplicationSubsystemStatus struct {
	Name  string
	State ApplicationSubsystemState
	Error error
}

type Supervisor struct {
	applications map[string]ApplicationSubsystem
	clock        misas.Clock
	eventBus     systemEventBus

	supervisedApplications map[string]*supervisedApplicationSubsystem
}

func NewSupervisor() *Supervisor { return &Supervisor{} }

func (s *Supervisor) Name() string { return "supervisor" }

func (s *Supervisor) WithApplicationSubsystem(app ApplicationSubsystem) *Supervisor {
	if s.applications == nil {
		s.applications = make(map[string]ApplicationSubsystem, 10)
	}
	s.applications[app.Name()] = app
	return s
}

func (s *Supervisor) Initialize(ctx context.Context) error {
	if s.supervisedApplications == nil {
		s.supervisedApplications = make(map[string]*supervisedApplicationSubsystem, len(s.applications))
	}

	for _, app := range s.applications {
		supervisedApp := &supervisedApplicationSubsystem{
			ApplicationSubsystem: newManagedApplicationSubsystem(app, s.eventBus, s.clock),
		}
		s.supervisedApplications[app.Name()] = supervisedApp
		appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})
		if err := supervisedApp.Initialize(appCtx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Supervisor) Run(ctx context.Context) error {
	for _, app := range s.supervisedApplications {
		appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})
		go func(ctx context.Context, a *supervisedApplicationSubsystem) {
			// we can safely ignore the error here, as it can be captured through the system's event bus.
			_ = a.Run(ctx)
		}(appCtx, app)
	}

	errs := lo.FilterMap(lo.Values(s.supervisedApplications), func(app *supervisedApplicationSubsystem, _ int) (error, bool) {
		state := app.Status()
		return state.Error, state.Error != nil
	})

	if len(errs) > 0 {
		return errors.Join(errs...)
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
			a.Stop()
		}(app)
	}
	wg.Wait()

	return err
}

func (s *Supervisor) Teardown(ctx context.Context) error {
	for _, app := range s.supervisedApplications {
		appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: app.Name()})
		if err := app.Teardown(appCtx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Supervisor) Stop(appName string) error {
	app, exists := s.supervisedApplications[appName]
	if !exists {
		return errors.New("application not found")
	}

	app.Stop()

	return nil
}

func (s *Supervisor) Start(ctx context.Context, appName string) error {
	app, exists := s.supervisedApplications[appName]
	if !exists {
		return errors.New("application not found")
	}

	return app.Start(ctx)
}

func (s *Supervisor) Status() []SupervisedApplicationSubsystemStatus {
	return lo.Map(lo.Values(s.supervisedApplications), func(app *supervisedApplicationSubsystem, _ int) SupervisedApplicationSubsystemStatus {
		return app.Status()
	})
}
