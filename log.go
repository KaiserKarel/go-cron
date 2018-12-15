package cron

import (
	"time"
)

// Log is emitted after a job has run
type Log struct {
	ID      string
	Entry   Entry
	Started time.Time
	Ended   time.Time
	Report  string
	Err     error
}

func newLog(entry Entry, started time.Time) Log {
	return Log{
		ID:      entry.ID + started.String(),
		Entry:   entry,
		Started: started,
	}
}
