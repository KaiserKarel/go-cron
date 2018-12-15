package cron

import (
	"context"
	"sync"
)

// Context is used to control running jobs.
type Context interface {
	context.Context
	Start()
	Running() bool
	Cancel()
	Entry() Entry
	Report() string
	WriteReport(string)
}

type ctx struct {
	mu sync.RWMutex
	context.Context
	entry   Entry
	cf      context.CancelFunc
	running bool
	report  string
}

// Start sets the state of the job ctx to running. If already started, Start is a noop.
func (c *ctx) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = true
}

// Running returns the job running state.
func (c *ctx) Running() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.running
}

// Cancel sends a cancel signal to the job.
func (c *ctx) Cancel() {
	c.cf()
}

// Entry returns the job entry.
func (c *ctx) Entry() Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.entry
}

// Report returns the job report.
func (c *ctx) Report() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.report
}

// WriteReport sets the job report.
func (c *ctx) WriteReport(report string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.report = report
}

// FromContext creates a cron Context from existing contexts and the entry. The contexts defaults to a running status of False.
func FromContext(ct context.Context, entry Entry) Context {
	ct, cf := context.WithCancel(ct)

	return &ctx{
		mu:      sync.RWMutex{},
		Context: ct,
		cf:      cf,
		entry:   entry,
		running: false,
	}
}
