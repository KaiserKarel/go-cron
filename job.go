package cron

import (
	"context"
	"errors"
)

var (
	// ErrRepeatable indicates the job should be retried immediately.
	ErrRepeatable = errors.New("temporary error")

	// ErrRepeatNextCron indicates the job may be retried when the schedule allows.
	ErrRepeatNextCron = errors.New("attempt at next cron")

	// ErrPermanentFailure indicates the cron should be removed from the schedule.
	ErrPermanentFailure = errors.New("never repeat this cron")

	// ErrCronFailure indicates the cron executor should cease running.
	ErrCronFailure = errors.New("cease all cron operations")
)

// Job is a registerable function. Passed arguments are obtained from the Entry.
type Job func(ctx context.Context, args map[string]interface{}) error
