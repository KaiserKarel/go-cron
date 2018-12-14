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
}

type ctx struct {
	mu sync.RWMutex
	context.Context
	entry   Entry
	cf      context.CancelFunc
	running bool
}

func (c *ctx) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = true
}

func (c *ctx) Running() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.running
}

func (c *ctx) Cancel() {
	c.cf()
}

func (c *ctx) Entry() Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.entry
}

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
