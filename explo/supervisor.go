package explo

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/samber/lo"
)

type ApplicationSubsystemState string

const (
	ApplicationSubsystemStateStopped ApplicationSubsystemState = "STOPPED"
	ApplicationSubsystemStateRunning ApplicationSubsystemState = "RUNNING"
	ApplicationSubsystemStateFailed  ApplicationSubsystemState = "FAILED"
)

// ApplicationSubsystemStatus represents the complete status of a supervised application subsystem
type ApplicationSubsystemStatus struct {
	Name  string
	State ApplicationSubsystemState
	Error error
}

type SupervisorApplicationSubsystem struct {
	applications map[string]*supervisedApplicationSubsystem
	logger       *slog.Logger
}

func NewSupervisorSubsystem(logger *slog.Logger) *SupervisorApplicationSubsystem {
	return &SupervisorApplicationSubsystem{
		applications: make(map[string]*supervisedApplicationSubsystem, 10),
		logger:       logger,
	}
}

func (s *SupervisorApplicationSubsystem) Name() string { return "supervisor" }

func (s *SupervisorApplicationSubsystem) WithApplicationSubsystem(as ApplicationSubsystem) *SupervisorApplicationSubsystem {
	if s.applications == nil {
		s.applications = make(map[string]*supervisedApplicationSubsystem)
	}
	s.applications[as.Name()] = &supervisedApplicationSubsystem{
		ApplicationSubsystem: NewLoggingApplicationSubsystem(as, s.logger),
	}

	return s
}

func (s *SupervisorApplicationSubsystem) Initialize() error {
	for _, app := range s.applications {
		if err := app.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SupervisorApplicationSubsystem) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}

	for _, app := range s.applications {
		wg.Add(1)
		go func(a *supervisedApplicationSubsystem) {
			defer wg.Done()
			a.Run(ctx)
		}(app)
	}

	wg.Wait()

	errs := lo.FilterMap(lo.Values(s.applications), func(app *supervisedApplicationSubsystem, _ int) (error, bool) {
		state := app.Status()
		return state.Error, state.Error != nil
	})

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *SupervisorApplicationSubsystem) Teardown(ctx context.Context) error {
	for _, app := range s.applications {
		if err := app.Teardown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *SupervisorApplicationSubsystem) Stop(appName string) error {
	app, exists := s.applications[appName]
	if !exists {
		return errors.New("application not found")
	}

	app.Stop()

	return nil
}

func (s *SupervisorApplicationSubsystem) Start(appName string, ctx context.Context) error {
	app, exists := s.applications[appName]
	if !exists {
		return errors.New("application not found")
	}

	return app.Start(ctx)
}

func (s *SupervisorApplicationSubsystem) Status() []ApplicationSubsystemStatus {
	return lo.Map(lo.Values(s.applications), func(app *supervisedApplicationSubsystem, _ int) ApplicationSubsystemStatus {
		return app.Status()
	})
}
