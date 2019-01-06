package cron

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

var (
	// ErrJobExists is returned by the executor instance if the provided job ID is present in the registry.
	ErrJobExists = errors.New("job already registered")
	// ErrJobNotExists is returned/used by the executor if a job is not registered but called/queried.
	ErrJobNotExists = errors.New("job not registered")

	// DefaultLocation is the default locale of the executor.
	DefaultLocation = time.UTC
)

type ErrNotFound error

// New is the constructor for executor
func New(opts ...Option) (*Executor, error) {
	e := &Executor{
		registry: make(map[string]Routine),
		context:  make(map[string]context.Context),
		entries:  map[string]*Entry{},
		stop:     make(chan struct{}),
	}

	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return nil, err
		}
	}

	if e.refresh == nil {
		e.refresh = make(chan Update, 100)
	}

	if e.errors == nil {
		e.errors = make(chan error, 100)
	}

	if e.log == nil {
		e.log = make(chan Log, 100)
	}

	if e.tab == nil {
		e.tab = NewMemoryTab()
	}

	if e.location == nil {
		e.location = DefaultLocation
	}

	return e, nil
}

// Executor is the godlevel struct.
type Executor struct {
	mu      sync.RWMutex
	running bool

	rmu      sync.RWMutex
	registry map[string]Routine
	tab      Tab

	cmu     sync.RWMutex
	context map[string]context.Context

	emu     sync.RWMutex
	entries map[string]*Entry
	ordered []string
	refresh chan Update

	errors   chan error
	stop     chan struct{}
	log      chan Log
	location *time.Location
}

// Register a routine with the global executor.
func (e *Executor) Register(ID string, routine Routine) error {
	e.rmu.Lock()
	defer e.rmu.Unlock()

	if _, exist := e.registry[ID]; exist {
		return ErrJobExists
	}

	e.registry[ID] = routine
	return nil
}

// Start the executor. Function registry should have completed before calling start.
func (e *Executor) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return errors.New("executor already running")
	}

	e.running = true
	e.mu.Unlock()
	e.emu.Lock()
	var err error

	e.entries, err = e.tab.All()
	if err != nil {
		return err
	}
	e.emu.Unlock()

	for {
		now := e.now()

		// calculate nextRuns
		for _, entry := range e.entries {
			var err error
			entry.NextRun, err = entry.Next(now)
			if err != nil {
				e.errors <- errors.Wrapf(err, "invalid cron expression for ID: %s", entry.ID)
			}
		}

		// order by nextrun and start a timer
		e.ordered = keys(e.entries)

		var timer *time.Timer
		if len(e.entries) == 0 {
			timer = time.NewTimer(time.Hour)
		} else {
			timer = time.NewTimer(e.entries[e.ordered[0]].NextRun.Sub(now))
		}

		select {
		// when the Stop method is called
		case <-e.stop:
			timer.Stop()
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			return nil

		// when the tab sends on the Refresh channel; either a new/altered item or an update.
		case update := <-e.tab.Refresh():
			switch update.UpdateType {
			case NewAltered:
				e.entries[update.ID] = &update.Entry
			case Deleted:
				e.emu.Lock()
				delete(e.entries, update.ID)
				e.emu.Unlock()
			}
			e.refresh <- update // TODO make this leaky
			continue            // continue the loop to ensure we restart the timer.

		// When the next cronjob is ready
		case <-timer.C:
			e.emu.RLock()

			for _, ID := range e.ordered {
				entry, exists := e.entries[ID]
				if !exists {
					continue
				}

				select {
				case <-e.stop:
					return nil
				default:
					now := time.Now()

					e.rmu.RLock()
					job, ok := e.registry[entry.Routine]
					e.rmu.RUnlock()

					if !ok {
						return ErrJobNotExists
					}

					if entry.NextRun.Before(now) {
						entry.PreviousRun = now
						entry.NextRun = entry.MustNext(now)
						go e.runJob(job, *entry, now)
					}
				}
			}
			e.emu.RUnlock()
		}
	}
}

// Add a cron entry to the executor
func (e *Executor) Add(entry Entry) error {
	err := e.tab.Put(entry)
	if err != nil {
		return err
	}
	return nil
}

// Stop the cron executor.
func (e *Executor) Stop() {
	e.stop <- struct{}{}
}

// StopAll stops the cron executor and all running cronjobs.
func (e *Executor) StopAll() {
	e.stop <- struct{}{}
	e.CancelAll()
}

// CancelAll stops all running cronjobs
func (e *Executor) CancelAll() {
	e.cmu.RLock()
	defer e.cmu.RUnlock()

	for _, ctx := range e.context {
		cf := CancelFromContext(ctx)
		cf()
	}
}

// Remove an entry from the tab. Does not cancel running job nor removes it from the current queue.
func (e *Executor) Remove(ID string) error {
	e.emu.Lock()
	defer e.emu.Unlock()
	delete(e.entries, ID)

	return e.tab.Remove(ID)
}

// Log returns the log channel
func (e *Executor) Log() chan Log {
	return e.log
}

// Location returns the exectutor timezone locale
func (e *Executor) Location() time.Location {
	return *e.location
}

// IsRunning returns true if the executor is running.
func (e *Executor) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Err returns the error channel
func (e *Executor) Err() chan error {
	return e.errors
}

// Refresh blocks on the refresh channel,which returns updates on the internal executor queue
func (e *Executor) Refresh() <-chan Update {
	return e.refresh
}

// Cancel a specific entry while running. If it is not running, Cancel returns ErrNotFound
func (e *Executor) Cancel(ID string) error {
	e.cmu.RLock()
	defer e.cmu.RUnlock()

	var ctx context.Context
	var exists bool
	if ctx, exists = e.context[ID]; !exists {
		return ErrNotFound(errors.New("job not found"))
	}
	cf := CancelFromContext(ctx)
	cf()
	return nil
}

func (e *Executor) runJob(routine Routine, entry Entry, now time.Time) {
	// Handle job panics
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err := errors.Errorf("cron: panic running job: %v\n%s", r, buf)
			e.errors <- err
		}
	}()

	e.cmu.RLock()
	ctx, exists := e.context[entry.ID]
	e.cmu.RUnlock()

	// if the ctx does exist, the job has already ran before and might still be running.
	if exists {
		cf := CancelFromContext(ctx)

		switch entry.Policy {
		case SingleInstanceOnly:
			cf()
			// wait for the job to stop running
			for ContextRunning(ctx) {
				runtime.Gosched()
			}
		case CancelRunning:
			cf()
		case SkipIfRunning:
			if ContextRunning(ctx) {
				return
			}
		// default equates to RunParallel.
		default:
			break
		}
	}

	back := backoff.NewExponentialBackOff()
	ctx, cf := NewContext(context.Background(), entry)
	// Setting the new context
	e.cmu.Lock()
	e.context[entry.ID] = ctx
	e.cmu.Unlock()

	// try the job with exponential backoff
	err := backoff.Retry(
		func() error {
			ContextStart(ctx)
			var err error
			err = routine(ctx, entry.Args)
			defer cf()

			switch err.(type) {
			case ErrRepeatable:
				return err
			case ErrRepeatNextCron:
				return backoff.Permanent(err)
			case ErrPermanentFailure:
				e.rmu.Lock()
				defer e.rmu.Unlock()

				tabErr := e.tab.Remove(entry.ID)
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

	log := newLog(entry, now)
	log.Ended = time.Now()
	log.Err = err
	log.Report = ReportFromContext(ctx)

	select {
	case e.log <- log:
		return

	// if the logchannel is full, we discard the oldest log obj.
	default:
		<-e.log
		e.log <- log
	}
}

func (e *Executor) now() time.Time {
	return time.Now().In(e.location)
}
