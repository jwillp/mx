package mx

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultRestartPolicy = NewRestartPolicy().
	OnAnyError().
	WithExponentialBackoff(time.Second, 2.0, lo.ToPtr(time.Minute))

type RestartPolicy struct {
	MaxRestarts  int
	RestartDelay func(RestartPolicyState, error) time.Duration
	ErrorFilter  func(RestartPolicyState, error) bool

	state *RestartPolicyState
}

func NewRestartPolicy() *RestartPolicy {
	return &RestartPolicy{
		MaxRestarts:  0,
		RestartDelay: func(RestartPolicyState, error) time.Duration { return 0 },
		ErrorFilter:  func(RestartPolicyState, error) bool { return false },
		state:        &RestartPolicyState{},
	}
}

func (p *RestartPolicy) Never() *RestartPolicy {
	p.ErrorFilter = func(RestartPolicyState, error) bool { return false }

	return p
}

func (p *RestartPolicy) Always() *RestartPolicy {
	p.ErrorFilter = func(RestartPolicyState, error) bool { return true }

	return p
}

func (p *RestartPolicy) OnAnyError() *RestartPolicy {
	p.ErrorFilter = func(_ RestartPolicyState, err error) bool { return err != nil }

	return p
}

func (p *RestartPolicy) OnError(filter func(error) bool) *RestartPolicy {
	p.ErrorFilter = func(_ RestartPolicyState, err error) bool { return filter(err) }

	return p
}

func (p *RestartPolicy) WithFixedDelay(delay time.Duration) *RestartPolicy {
	p.RestartDelay = func(RestartPolicyState, error) time.Duration { return delay }

	return p
}

func (p *RestartPolicy) WithExponentialBackoff(
	initialDelay time.Duration,
	factor float64,
	maxDelay *time.Duration,
) *RestartPolicy {
	p.RestartDelay = func(s RestartPolicyState, err error) time.Duration {
		d := time.Duration(float64(initialDelay) * math.Pow(factor, float64(s.AttemptCount-1)))
		if maxDelay != nil && d > *maxDelay {
			return *maxDelay
		}

		return d
	}

	return p
}

type RestartPolicyState struct {
	AttemptCount int
	Errors       []error
}

func (p *RestartPolicy) ShouldRestart(err error) bool {
	shouldRestart := p.ErrorFilter(*p.state, err)
	if shouldRestart && p.MaxRestarts > 0 && p.state.AttemptCount+1 > p.MaxRestarts {
		return false
	}

	return shouldRestart
}

func (p *RestartPolicy) onRestart(err error) {
	p.state.AttemptCount++
	p.state.Errors = append(p.state.Errors, err)
}

func (p *RestartPolicy) delay() time.Duration {
	var delay time.Duration
	if p.RestartDelay != nil {
		delay = p.RestartDelay(*p.state, nil)
	}

	if delay == 0 {
		return time.Second
	}

	return delay
}

type SupervisionOptions struct {
	RestartPolicy *RestartPolicy
}

type supervisedApplicationSubsystem struct {
	ApplicationSubsystem
	Options SupervisionOptions

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
	for {
		// if manually stopped, wait until resumed or terminated or parent ctx done
		if atomic.LoadUint32(&s.stopped) == 1 {
			select {
			case <-s.resumeChan:
				atomic.StoreUint32(&s.stopped, 0)
				// continue to start running
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
			err := s.runOnce(appCtx)
			if err == nil {
				continue
			}
			// Ask policy whether to restart. ShouldRestart increments AttemptCount when it decides to allow a restart.
			if s.shouldRestart(err) {
				restartPolicy := s.Options.RestartPolicy
				restartDelay := restartPolicy.delay()
				time.Sleep(restartDelay)
				restartPolicy.onRestart(err)
				Log(ctx).Info(
					fmt.Sprintf("restarting supervised application subsystem %q", s.Name()),
					slog.Int("restartCount", restartPolicy.state.AttemptCount),
					slog.Int("maxRestarts", restartPolicy.MaxRestarts),
					slog.Duration("restartDelay", restartDelay),
				)
				continue
			}
			return err
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

func (s *supervisedApplicationSubsystem) shouldRestart(err error) bool {
	if s.Options.RestartPolicy == nil {
		s.Options.RestartPolicy = DefaultRestartPolicy
	}
	return s.Options.RestartPolicy.ShouldRestart(err)
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
