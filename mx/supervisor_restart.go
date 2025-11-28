package mx

import (
	"math"
	"sync"
	"time"
)

const (
	RestartPolicyNo            = "no"
	RestartPolicyAlways        = "always"
	RestartPolicyOnFailure     = "on-failure"
	RestartPolicyUnlessStopped = "unless-stopped"
)

var DefaultRestartPolicy = NewRestartPolicy(RestartPolicyOnFailure)

type RestartPolicy struct {
	policy string

	MaxRetries       int
	MaxRetryDuration time.Duration

	circuitBreakerEnabled      bool
	circuitBreakerThreshold    int // Failures in window to trigger circuit break (rapid-fire safety)
	circuitBreakerResetTimeout time.Duration
	circuitBreakerWindow       time.Duration // Time window for circuit breaker counting

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

func NewRestartPolicy(policy string) *RestartPolicy {
	return &RestartPolicy{
		policy:                     policy,
		MaxRetries:                 5,
		MaxRetryDuration:           0,
		circuitBreakerEnabled:      true,
		circuitBreakerThreshold:    3, // Separate threshold for rapid-fire detection
		circuitBreakerResetTimeout: 30 * time.Second,
		circuitBreakerWindow:       10 * time.Second,
		state:                      &restartPolicyState{},
	}
}

func (p *RestartPolicy) WithMaxRetries(max int) *RestartPolicy {
	p.MaxRetries = max
	return p
}

func (p *RestartPolicy) WithMaxRetryDuration(d time.Duration) *RestartPolicy {
	p.MaxRetryDuration = d
	return p
}

func (p *RestartPolicy) WithCircuitBreaker(enabled bool) *RestartPolicy {
	p.circuitBreakerEnabled = enabled
	return p
}

func (p *RestartPolicy) WithCircuitBreakerThreshold(threshold int) *RestartPolicy {
	p.circuitBreakerThreshold = threshold
	return p
}

func (p *RestartPolicy) WithCircuitBreakerWindow(window time.Duration) *RestartPolicy {
	p.circuitBreakerWindow = window
	return p
}

func (p *RestartPolicy) WithCircuitBreakerResetTimeout(timeout time.Duration) *RestartPolicy {
	p.circuitBreakerResetTimeout = timeout
	return p
}

func (p *RestartPolicy) ShouldRestart(err error) bool {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	switch p.policy {
	case RestartPolicyNo:
		return false
	case RestartPolicyAlways:
		return true
	case RestartPolicyUnlessStopped:
		return true
	case RestartPolicyOnFailure:
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

func (p *RestartPolicy) IsCircuitOpen() bool {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()
	return p.state.circuitOpen
}

func (p *RestartPolicy) CanRestart(now time.Time) bool {
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

	if p.MaxRetries > 0 && p.state.attemptCount >= p.MaxRetries {
		return false
	}

	if p.MaxRetryDuration > 0 && len(p.state.failureWindow) > 0 {
		elapsed := now.Sub(p.state.failureWindow[0])
		if elapsed > p.MaxRetryDuration {
			return false
		}
	}

	return true
}

func (p *RestartPolicy) NextRetryDelay() time.Duration {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	exponentialDelay := time.Duration(math.Pow(2, float64(p.state.attemptCount))) * time.Second
	maxDelay := 5 * time.Minute
	if exponentialDelay > maxDelay {
		exponentialDelay = maxDelay
	}

	return exponentialDelay
}

func (p *RestartPolicy) RecordAttempt() {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()
	p.state.attemptCount++
}

func (p *RestartPolicy) ResetState() {
	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	p.state.attemptCount = 0
	p.state.failureWindow = nil
	p.state.circuitOpen = false
	p.state.lastError = nil
}

func (p *RestartPolicy) GetState() RestartPolicyState {
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
