package circuitbreaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type state int

const (
	open state = iota
	halfOpen
	closed
)

type CircuitBreaker struct {
	mu              sync.Mutex
	state           state
	failureCount    int
	maxFailureCount int
	resetTimeout    time.Duration
	openChannel     chan struct{}
}

func (cb *CircuitBreaker) openWatcher() {
	for range cb.openChannel {
		time.Sleep(cb.resetTimeout)
		cb.mu.Lock()
		cb.state = halfOpen
		cb.failureCount = 0
		cb.mu.Unlock()
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == open {
		return errors.New("circuit breaker is open")
	}

	cb.mu.Lock()
	err := fn()
	defer cb.mu.Unlock()

	// here, the state can be either half-open or closed
	if err == nil {
		cb.state = closed
		cb.failureCount = 0
		return nil
	}

	// at this point, there is an error occured in the function
	if cb.state == halfOpen {
		cb.state = open
		cb.openChannel <- struct{}{}
	}

	cb.failureCount++
	if cb.failureCount >= cb.maxFailureCount {
		cb.state = open
		cb.openChannel <- struct{}{}
	}

	return err
}

func callDownstream(requestId int, cb *CircuitBreaker, wg *sync.WaitGroup) {
	defer wg.Done()

	err := cb.Execute(func() error {
		if requestId%3 == 2 {
			return errors.New("error :: request failed")
		}

		fmt.Println("success :: requestId:", requestId)
		return nil
	})

	if err != nil {
		fmt.Println("error :: ", err, " for requestId:", requestId)
	}
	time.Sleep(300 * time.Millisecond)
}

func NewCircuitBreaker(maxFailureCount int, resetTimeout time.Duration) *CircuitBreaker {
	circuitBreaker := &CircuitBreaker{
		state:           closed,
		maxFailureCount: maxFailureCount,
		resetTimeout:    resetTimeout,
		openChannel:     make(chan struct{}),
	}

	go circuitBreaker.openWatcher()
	return circuitBreaker
}
