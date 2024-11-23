package circuit_breaker

import (
	"errors"
	"go.uber.org/atomic"
	"time"
)

const (
	stateClosed   = "closed"
	stateOpen     = "open"
	stateHalfOpen = "half_open"
)

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

type circuitBreaker struct {
	state         string
	failureCount  *atomic.Uint64
	maxFailures   uint64
	resetTimeout  time.Duration
	halfOpenTimer *time.Timer
}

func New(maxFailures uint64, resetTimeout time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:        stateClosed,
		failureCount: atomic.NewUint64(0),
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *circuitBreaker) Call(operation func() error) error {
	switch cb.state {
	case stateOpen:
		return ErrCircuitBreakerOpen
	case stateHalfOpen, stateClosed:
		err := operation()
		if err != nil {
			cb.recordFailure()
			return err
		}
		cb.reset()
		return nil
	}
	return nil
}

func (cb *circuitBreaker) recordFailure() {
	fCount := cb.failureCount.Inc()
	if fCount >= cb.maxFailures {
		cb.transitionTo(stateOpen)
		cb.halfOpenTimer = time.AfterFunc(cb.resetTimeout, func() {
			cb.transitionTo(stateHalfOpen)
		})
	}
}

func (cb *circuitBreaker) reset() {
	cb.failureCount = atomic.NewUint64(0)
	cb.transitionTo(stateClosed)
}

func (cb *circuitBreaker) transitionTo(state string) {
	cb.state = state
}
