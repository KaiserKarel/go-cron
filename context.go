package cron

import (
	"context"
)

type ctxKey int

const (
	running ctxKey = 1
	entry   ctxKey = 2
	report  ctxKey = 3
	cf      ctxKey = 4
)

// ContextStart sets the state of the job ctx to running. If already started, Start is a noop.
func ContextStart(ctx context.Context) context.Context {
	return context.WithValue(ctx, running, true)
}

// ContextRunning returns the job running state.
func ContextRunning(ctx context.Context) bool {
	running, set := ctx.Value(running).(bool)
	return running && set
}

// EntryFromContext returns the job entry.
func EntryFromContext(ctx context.Context) Entry {
	entry, set := ctx.Value(entry).(Entry)
	if !set {
		return Entry{}
	}
	return entry
}

// ReportFromContext returns the job report.
func ReportFromContext(ctx context.Context) string {
	return ctx.Value(report).(string)
}

// CancelFromContext returns the cancel func.
func CancelFromContext(ctx context.Context) context.CancelFunc {
	cfunc, set := ctx.Value(cf).(context.CancelFunc)
	if !set {
		return nil
	}
	return cfunc
}

// ContextWithCancel sets the cancel func in a context.
func ContextWithCancel(ctx context.Context, cfunc context.CancelFunc) context.Context {
	return context.WithValue(ctx, cf, cfunc)

}

// ContextWithReport sets the job report.
func ContextWithReport(ctx context.Context, rep string) context.Context {
	return context.WithValue(ctx, report, rep)
}

// ContextWithEntry sets the job Entry.
func ContextWithEntry(ctx context.Context, e Entry) context.Context {
	return context.WithValue(ctx, entry, e)
}

// NewContext creates a cron Context from existing contexts and the entry. The contexts defaults to a running status of False.
func NewContext(ctx context.Context, entry Entry) (context.Context, context.CancelFunc) {
	ctx, cf := context.WithCancel(ctx)
	return ContextWithCancel(ContextWithEntry(ctx, entry), cf), cf
}
