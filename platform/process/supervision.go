// Package process provides process supervision, signal handling, and graceful
// shutdown lifecycle management for straitd and sg-controller.
package process

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Supervisor manages long-running background tasks with graceful shutdown.
type Supervisor struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdownMu sync.Mutex
	isShutdown bool
	timeout    time.Duration
}

// NewSupervisor creates a new process supervisor with the given shutdown timeout.
func NewSupervisor(timeout time.Duration) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Supervisor{
		ctx:     ctx,
		cancel:  cancel,
		timeout: timeout,
	}
}

// Context returns the supervisor's lifecycle context.
func (s *Supervisor) Context() context.Context {
	return s.ctx
}

// Go launches a supervised background goroutine.
func (s *Supervisor) Go(fn func(ctx context.Context)) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		fn(s.ctx)
	}()
}

// Stop initiates graceful shutdown and waits for all supervised goroutines.
func (s *Supervisor) Stop() {
	s.shutdownMu.Lock()
	if s.isShutdown {
		s.shutdownMu.Unlock()
		return
	}
	s.isShutdown = true
	s.cancel()
	s.shutdownMu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(s.timeout):
	}
}

// WaitForSignals blocks until SIGINT or SIGTERM is received, then triggers Stop.
func (s *Supervisor) WaitForSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
		s.Stop()
	case <-s.ctx.Done():
	}
}
