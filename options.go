package cron

import "time"

// Option is a constructor function
type Option func(*Executor) error

// WithTab sets a storage backend for the executor
func WithTab(tab Tab) Option {
	return func(e *Executor) error {
		e.tab = tab
		return nil
	}
}

// WithLog sets a log channel
func WithLog(log chan Log) Option {
	return func(e *Executor) error {
		e.log = log
		return nil
	}
}

// WithError sets an error channel
func WithError(errors chan error) Option {
	return func(e *Executor) error {
		e.errors = errors
		return nil
	}
}

// WithLocation sets the location.
//
// Location defaults to time.Local.
func WithLocation(location time.Location) Option {
	return func(e *Executor) error {
		e.location = &location
		return nil
	}
}
