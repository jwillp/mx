package mx

import (
	"math"
	"sync"
	"time"
)

const (
	ApplicationSubsystemRestartPolicyNo            = "no"
	ApplicationSubsystemRestartPolicyAlways        = "always"
	ApplicationSubsystemRestartPolicyOnFailure     = "on-failure"
	ApplicationSubsystemRestartPolicyUnlessStopped = "unless-stopped"
)

const (
	defaultMaxRestarts               = 5
	defaultCircuitBreakerThreshold   = 3
	defaultCircuitBreakerResetTimout = 30 * time.Second
	defaultCircuitBreakerWindow      = 10 * time.Second
)

var DefaultRestartPolicy = NewApplicationSubsystemRestartPolicy(ApplicationSubsystemRestartPolicyOnFailure)

type RestartPolicy struct {
	policy string

	maxRestarts      int
	maxRetryDuration time.Duration

	circuitBreakerEnabled bool
	// number of failures within the window to trigger circuit breaker and prevent
	// rapid-fire restart because of quick errors.
	circuitBreakerThreshold int

	// maximum time the circuit breaker remains open before allowing restarts again
	circuitBreakerResetTimeout time.Duration
	// time window for counting failures towards the circuit breaker threshold
	circuitBreakerWindow time.Duration

	state *restartPolicyState
}

type restartPolicyState struct {
	mu              sync.Mutex
	attemptCount    int
	lastError       error
	failureWindow   []time.Time
	circuitOpen     bool
	circuitOpenTime time.Time
}

func NewApplicationSubsystemRestartPolicy(policy string) *RestartPolicy {
	return &RestartPolicy{
		policy:                     policy,
		maxRestarts:                defaultMaxRestarts,
		circuitBreakerEnabled:      true,
		circuitBreakerThreshold:    defaultCircuitBreakerThreshold,
		circuitBreakerResetTimeout: defaultCircuitBreakerResetTimout,
		circuitBreakerWindow:       defaultCircuitBreakerWindow,
		state:                      &restartPolicyState{},
	}
}

func (p *RestartPolicy) WithMaxRestarts(max int) *RestartPolicy {
	p.maxRestarts = max
	return p
}

func (p *RestartPolicy) shouldRestart(err error) bool {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	switch p.policy {
	case ApplicationSubsystemRestartPolicyNo:
		return false
	case ApplicationSubsystemRestartPolicyAlways:
		return true
	case ApplicationSubsystemRestartPolicyUnlessStopped:
		return true
	case ApplicationSubsystemRestartPolicyOnFailure:
		return err != nil
	default:
		return false
	}
}

func (p *RestartPolicy) recordFailure(err error, now time.Time) {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	p.state.lastError = err
	p.state.failureWindow = append(p.state.failureWindow, now)

	p.prunFailureWindow(now)
	p.updateCircuitBreaker(now)
}

func (p *RestartPolicy) prunFailureWindow(now time.Time) {
	cutoff := now.Add(-p.circuitBreakerWindow)
	i := 0
	for i < len(p.state.failureWindow) && p.state.failureWindow[i].Before(cutoff) {
		i++
	}
	p.state.failureWindow = p.state.failureWindow[i:]
}

func (p *RestartPolicy) updateCircuitBreaker(now time.Time) {
	if !p.circuitBreakerEnabled {
		return
	}

	if p.state.circuitOpen {
		if now.Sub(p.state.circuitOpenTime) > p.circuitBreakerResetTimeout {
			p.state.circuitOpen = false
			p.state.failureWindow = nil
		}
		return
	}

	// Circuit breaker opens when failures reach threshold within the window (rapid-fire safety)
	if len(p.state.failureWindow) >= p.circuitBreakerThreshold {
		p.state.circuitOpen = true
		p.state.circuitOpenTime = now
	}
}

func (p *RestartPolicy) isCircuitOpen() bool {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()
	return p.state.circuitOpen
}

func (p *RestartPolicy) canRestart(now time.Time) bool {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	if p.state.circuitOpen {
		if now.Sub(p.state.circuitOpenTime) > p.circuitBreakerResetTimeout {
			p.state.circuitOpen = false
			p.state.failureWindow = nil
			return true
		}
		return false
	}

	if p.maxRestarts > 0 && p.state.attemptCount >= p.maxRestarts {
		return false
	}

	return true
}

func (p *RestartPolicy) nextRetryDelay() time.Duration {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	exponentialDelay := time.Duration(math.Pow(2, float64(p.state.attemptCount))) * time.Second
	maxDelay := 5 * time.Minute
	if exponentialDelay > maxDelay {
		exponentialDelay = maxDelay
	}

	return exponentialDelay
}

func (p *RestartPolicy) recordAttempt() {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()
	p.state.attemptCount++
}

func (p *RestartPolicy) resetState() {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	p.state.attemptCount = 0
	p.state.failureWindow = nil
	p.state.circuitOpen = false
	p.state.lastError = nil
}

func (p *RestartPolicy) getState() RestartPolicyState {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	return RestartPolicyState{
		AttemptCount: p.state.attemptCount,
		FailureCount: len(p.state.failureWindow),
		CircuitOpen:  p.state.circuitOpen,
		LastError:    p.state.lastError,
	}
}

type RestartPolicyState struct {
	AttemptCount int
	FailureCount int
	CircuitOpen  bool
	LastError    error
}
