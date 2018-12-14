package cron

import (
	"context"
	"errors"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	// DefaultPeriod is the default period between executor wakenings
	DefaultPeriod = time.Second
)

var (
	// ErrJobExists is returned by the executor instance if the provided job ID is present in the registry.
	ErrJobExists = errors.New("job already registerd")
	// ErrJobNotExists is returned/used by the executor if a job is not registered but called/queried.
	ErrJobNotExists = errors.New("job not registered")
)

// Option is a constructor function
type Option func(*Executor) error

// New is the constructor for executor
func New(opts ...Option) (*Executor, error) {
	e := &Executor{
		rmu:      sync.RWMutex{},
		registry: make(map[string]Job),
		cmu:      sync.RWMutex{},
		context:  make(map[string]Context),
		entries:  []*Entry{},
		emu:      sync.RWMutex{},
		errors:   make(chan error, 100),
		stop:     make(chan struct{}),
	}

	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return nil, err
		}
	}

	if e.period == 0 {
		e.period = DefaultPeriod
	}

	if e.tab == nil {
		e.tab = NewMemoryTab()
	}
	return e, nil
}

// Executor is the godlevel struct.
type Executor struct {
	rmu      sync.RWMutex
	registry map[string]Job
	tab      Tab

	cmu     sync.RWMutex
	context map[string]Context

	emu     sync.RWMutex
	entries ByTimeAsc

	period time.Duration
	errors chan error
	stop   chan struct{}
}

// Register a job type
func (e *Executor) Register(ID string, job Job) error {
	e.rmu.Lock()
	defer e.rmu.Unlock()

	if _, exist := e.registry[ID]; exist {
		return ErrJobExists
	}

	e.registry[ID] = job
	return nil
}

// Start the executor. Function registry should have completed before calling start.
func (e *Executor) Start() error {
	var itertime time.Duration
	itertime = 200 * time.Microsecond // sensible default for first round

	for {
		// first get all (persistently) stored entries
		e.emu.Lock()
		var err error
		e.entries, err = e.tab.All()
		if err != nil {
			return err
		}
		e.emu.Unlock()

		select {
		case <-e.stop:
			return nil
		default:
			start := time.Now()

			e.emu.RLock()
			e.entries = e.entries.Unique()
			sort.Sort(e.entries)

			for _, entry := range e.entries {
				select {
				case <-e.stop:
					return nil
				default:
					now := time.Now()

					e.rmu.RLock()
					job, ok := e.registry[entry.Job]
					e.rmu.RUnlock()

					if !ok {
						return ErrJobNotExists
					}

					minval := now.Add(-e.period/2 - itertime)
					maxval := now.Add(e.period/2 - itertime)

					if entry.nextRun.IsZero() {
						entry.nextRun = entry.MustNext(now)
					}

					if entry.nextRun.Before(maxval) && entry.nextRun.After(minval) {
						entry.lastRun = now
						entry.nextRun = entry.MustNext(now)
						go e.runJob(job, entry, now)
					}
				}
			}
			itertime = time.Now().Sub(start)
			time.Sleep(e.period - itertime)
			e.emu.RUnlock()
		}
	}
}

// Add a cron entry to the executor
func (e *Executor) Add(entry *Entry) error {
	return e.tab.Put(entry)
}

// Stop the cron executor.
func (e *Executor) Stop() {
	e.stop <- struct{}{}
}

// StopAll stops the cron executor and all running cronjobs.
func (e *Executor) StopAll() {
	e.stop <- struct{}{}

	e.cmu.RLock()
	defer e.cmu.RUnlock()

	for _, ctx := range e.context {
		ctx.Cancel()
	}
}

// CancelAll stops all running cronjobs
func (e *Executor) CancelAll() {
	e.cmu.RLock()
	defer e.cmu.RUnlock()

	for _, ctx := range e.context {
		ctx.Cancel()
	}
}

// Remove an entry from the tab. Does not cancel running job nor removes it from the current queue.
func (e *Executor) Remove(entry *Entry) error {
	return e.tab.Remove(entry)
}

// Cancel a specific entry while running. If it is not running, Cancel is a noop.
func (e *Executor) Cancel(ID string) {
	e.cmu.RLock()
	defer e.cmu.RUnlock()

	var ctx Context
	var exists bool
	if ctx, exists = e.context[ID]; !exists {
		return
	}
	ctx.Cancel()
}

func (e *Executor) runJob(job Job, entry *Entry, now time.Time) {

	switch entry.Policy {
	case SingleInstanceOnly:
		ctx := e.context[entry.ID]
		ctx.Cancel()

		// wait for the job to stop running
		for ctx.Running() {
			runtime.Gosched()
		}
	case CancelRunning:
		ctx := e.context[entry.ID]
		ctx.Cancel()
	case SkipIfRunning:
		ctx := e.context[entry.ID]
		if ctx.Running() {
			return
		}
	default:
		break
	}

	back := backoff.NewExponentialBackOff()
	ctx := FromContext(context.Background(), *entry)
	e.cmu.Lock()
	e.context[entry.ID] = ctx
	e.cmu.Unlock()

	err := backoff.Retry(
		func() error {
			ctx.Start()
			err := job(ctx, entry.Args)
			ctx.Cancel()

			switch err {
			case ErrRepeatable:
				return err
			case ErrRepeatNextCron:
				return backoff.Permanent(err)
			case ErrPermanentFailure:
				e.rmu.Lock()
				defer e.rmu.Unlock()

				tabErr := e.tab.Remove(entry)
				if tabErr != nil {
					e.errors <- tabErr
				}
				return backoff.Permanent(err)
			case ErrCronFailure:
				e.Stop()
				return backoff.Permanent(err)
			}
			return err
		}, back)

	if err != nil {
		e.errors <- err
	}
}
