// origin: https://github.com/robfig/cron/blob/master/schedule.go

package cron

import "time"

// The Schedule describes a job's duty cycle.
type Schedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}
