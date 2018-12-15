package cron

import (
	"context"
)

// ErrRepeatable indicates the job should be retried immediately.
type ErrRepeatable error

// ErrRepeatNextCron indicates the job may be retried when the schedule allows.
type ErrRepeatNextCron error

// ErrPermanentFailure indicates the cron should be removed from the schedule.
type ErrPermanentFailure error

// ErrCronFailure indicates the cron executor should cease running.
type ErrCronFailure error

// Routine is a registerable function. Passed arguments are obtained from the Entry.
type Routine func(ctx context.Context, args map[string]interface{}) error
