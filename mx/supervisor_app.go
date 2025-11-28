package mx

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type SupervisionOptions struct {
	RestartPolicy *RestartPolicy
}

// SupervisedApp is a control interface for a supervised application subsystem
type SupervisedApp interface {
	Stop()
	Start()
	Terminate()
}

type supervisedApplicationSubsystem struct {
	ApplicationSubsystem
	Options SupervisionOptions
	pm      PluginManager

	// lazy init for channels
	initOnce sync.Once

	// runtime state using atomics
	isRunning uint32 // 0 == false, 1 == true
	stopped   uint32 // 1 == manually stopped

	// control channels
	stopTrigger   chan struct{} // triggers cancellation of the current run (for Stop)
	resumeChan    chan struct{} // used to resume after Stop
	terminateChan chan struct{} // terminates the supervised subsystem
}

func (s *supervisedApplicationSubsystem) Run(ctx context.Context) error {
	appCtx := newSubsystemContext(ctx, SubsystemInfo{Name: s.Name()})
	s.ensureInit()

	if s.Options.RestartPolicy == nil {
		s.Options.RestartPolicy = DefaultRestartPolicy
	}

	for {
		if atomic.LoadUint32(&s.stopped) == 1 {
			select {
			case <-s.resumeChan:
				atomic.StoreUint32(&s.stopped, 0)
			case <-s.terminateChan:
				Log(ctx).Info(fmt.Sprintf("terminating supervised application subsystem %q", s.Name()))
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.terminateChan:
			Log(ctx).Info(fmt.Sprintf("terminating supervised application subsystem %q", s.Name()))
			return nil
		default:
			now := time.Now()
			err := s.runOnce(appCtx)

			if err == nil {
				s.Options.RestartPolicy.ResetState()
				continue
			}

			policy := s.Options.RestartPolicy
			if !policy.ShouldRestart(err) {
				return err
			}

			policy.recordFailure(err, now)

			if !policy.CanRestart(now) {
				reason := "max retries exceeded"
				if policy.IsCircuitOpen() {
					reason = "circuit breaker open (too many failures)"
				} else if policy.MaxRetryDuration > 0 {
					reason = "max retry duration exceeded"
				}

				state := policy.GetState()
				s.pm.DispatchHook(ctx, SubsystemMaxRestartReachedHook{
					ApplicationName: s.Name(),
					RestartCount:    state.AttemptCount,
					MaxAttempts:     policy.MaxRetries,
					Reason:          reason,
					Error:           err,
					ReachedAt:       now,
				})
				return err
			}

			delay := policy.NextRetryDelay()
			policy.RecordAttempt()
			state := policy.GetState()

			s.pm.DispatchHook(ctx, SubsystemWillRestartHook{
				ApplicationName: s.Name(),
				RestartCount:    state.AttemptCount,
				MaxAttempts:     policy.MaxRetries,
				RestartDelay:    delay,
				Error:           err,
				StartedAt:       now,
			})

			select {
			case <-time.After(delay):
			case <-s.terminateChan:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}

			s.pm.DispatchHook(ctx, SubsystemRestartedHook{
				ApplicationName: s.Name(),
				RestartCount:    state.AttemptCount,
				MaxAttempts:     policy.MaxRetries,
				Error:           nil,
				StartedAt:       now,
				EndedAt:         time.Now(),
			})
		}
	}
}

func (s *supervisedApplicationSubsystem) runOnce(ctx context.Context) error {
	// create a child cancellable context so we can cancel the running app on Stop/Terminate
	ctxRun, cancel := context.WithCancel(ctx)

	// exit signal to wake up watcher when Run completes
	exit := make(chan struct{})
	defer close(exit)

	// watcher watches for stopTrigger or terminate to cancel ctxRun
	go func() {
		select {
		case <-s.stopTrigger:
			cancel()
		case <-s.terminateChan:
			cancel()
		case <-exit:
		}
	}()

	atomic.StoreUint32(&s.isRunning, 1)
	defer atomic.StoreUint32(&s.isRunning, 0)

	err := s.ApplicationSubsystem.Run(ctxRun)
	cancel()

	return err
}

func (s *supervisedApplicationSubsystem) Stop() {
	atomic.StoreUint32(&s.stopped, 1)
	select {
	case s.stopTrigger <- struct{}{}:
	default:
	}
}

func (s *supervisedApplicationSubsystem) Start() {
	atomic.StoreUint32(&s.stopped, 0)
	select {
	case s.resumeChan <- struct{}{}:
	default:
	}
}

func (s *supervisedApplicationSubsystem) Terminate() {
	s.ensureInit()
	select {
	case s.terminateChan <- struct{}{}:
	default:
	}
}

// ensureInit lazily initializes the control channels
func (s *supervisedApplicationSubsystem) ensureInit() {
	s.initOnce.Do(func() {
		s.stopTrigger = make(chan struct{}, 1)
		s.resumeChan = make(chan struct{}, 1)
		s.terminateChan = make(chan struct{}, 1)
	})
}
