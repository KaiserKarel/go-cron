package cron

import "time"

// WithPeriod sets a minimum period for the executor.
func WithPeriod(period time.Duration) Option {
	return func(e *Executor) error {
		e.period = period
		return nil
	}
}

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
