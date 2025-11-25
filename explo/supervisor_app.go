package explo

import (
	"context"
	"sync"
)

// supervisedApplicationSubsystem is an internal decorator that adds cancellation capability to an application subsystem
type supervisedApplicationSubsystem struct {
	ApplicationSubsystem
	cancel context.CancelFunc
	err    error
	mu     sync.RWMutex
}

func (s *supervisedApplicationSubsystem) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return nil
	}
	s.err = nil
	s.mu.Unlock()

	s.mu.Lock()
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	err := s.ApplicationSubsystem.Run(ctx)
	cancel()

	s.mu.Lock()
	s.err = err
	s.cancel = nil
	s.mu.Unlock()

	return err
}

func (s *supervisedApplicationSubsystem) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *supervisedApplicationSubsystem) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return nil
	}
	s.err = nil
	s.mu.Unlock()

	appCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	go func() {
		err := s.ApplicationSubsystem.Run(appCtx)
		cancel()
		s.mu.Lock()
		s.err = err
		s.cancel = nil
		s.mu.Unlock()
	}()

	return nil
}

func (s *supervisedApplicationSubsystem) Status() ApplicationSubsystemStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := ApplicationSubsystemStateStopped
	if s.err != nil {
		state = ApplicationSubsystemStateFailed
	} else if s.cancel != nil {
		state = ApplicationSubsystemStateRunning
	}

	return ApplicationSubsystemStatus{
		Name:  s.ApplicationSubsystem.Name(),
		State: state,
		Error: s.err,
	}
}
