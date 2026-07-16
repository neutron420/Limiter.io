package middleware

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type State int32

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

type CircuitBreaker struct {
	mu              sync.RWMutex
	name            string
	state           State
	failureCount    int64
	successCount    int64
	failureThreshold int64
	successThreshold int64
	timeout         time.Duration
	lastStateChange time.Time
}

type CircuitBreakerRegistry struct {
	mu     sync.RWMutex
	breakers map[string]*CircuitBreaker
}

var GlobalCircuitBreakers = &CircuitBreakerRegistry{
	breakers: make(map[string]*CircuitBreaker),
}

func NewCircuitBreaker(name string, failureThreshold, successThreshold int64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		lastStateChange:  time.Now(),
	}
}

func (r *CircuitBreakerRegistry) GetOrCreate(name string, failureThreshold, successThreshold int64, timeout time.Duration) *CircuitBreaker {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cb, ok := r.breakers[name]; ok {
		return cb
	}
	cb := NewCircuitBreaker(name, failureThreshold, successThreshold, timeout)
	r.breakers[name] = cb
	return cb
}

func (cb *CircuitBreaker) State() State {
	return State(atomic.LoadInt32((*int32)(&cb.state)))
}

func (cb *CircuitBreaker) setState(s State) {
	atomic.StoreInt32((*int32)(&cb.state), int32(s))
	cb.lastStateChange = time.Now()
}

func (cb *CircuitBreaker) Allow() bool {
	state := cb.State()
	switch state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return true
}

func (cb *CircuitBreaker) Success() {
	state := cb.State()
	switch state {
	case StateHalfOpen:
		count := atomic.AddInt64(&cb.successCount, 1)
		if count >= cb.successThreshold {
			atomic.StoreInt64(&cb.successCount, 0)
			atomic.StoreInt64(&cb.failureCount, 0)
			cb.setState(StateClosed)
		}
	case StateClosed:
		atomic.StoreInt64(&cb.failureCount, 0)
	}
}

func (cb *CircuitBreaker) Failure() {
	count := atomic.AddInt64(&cb.failureCount, 1)
	if count >= cb.failureThreshold {
		atomic.StoreInt64(&cb.failureCount, 0)
		atomic.StoreInt64(&cb.successCount, 0)
		cb.setState(StateOpen)
	}
}

func CircuitBreakerMiddleware(name string, failureThreshold, successThreshold int64, timeout time.Duration) gin.HandlerFunc {
	cb := GlobalCircuitBreakers.GetOrCreate(name, failureThreshold, successThreshold, timeout)
	return func(c *gin.Context) {
		if !cb.Allow() {
			c.AbortWithStatusJSON(503, gin.H{
				"error":   "service temporarily unavailable",
				"breaker": name,
				"state":   "open",
			})
			return
		}
		c.Next()
		if c.Writer.Status() >= 500 {
			cb.Failure()
		} else {
			cb.Success()
		}
	}
}
