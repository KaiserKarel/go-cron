package cron

import (
	"time"
)

// RunPolicy defines running behaviour of an entry.
type RunPolicy int

const (
	// RunParallel runs an entry while ignoring other running instances.
	RunParallel RunPolicy = 0

	// CancelRunning cancels previously running entries before starting.
	CancelRunning RunPolicy = 1

	// SingleInstanceOnly cancels previously running entries, awaits cancellation confirmation before starting.
	SingleInstanceOnly RunPolicy = 2

	// SkipIfRunning ignores the entry if another instance is currently running
	SkipIfRunning RunPolicy = 3
)

// Entry specifies a single (crontab) entry, a function which is executed periodically.
type Entry struct {
	// Globally unique ID
	ID string

	// cron expression
	Expression string

	// Job-identifier to be executed
	Job string

	// Arguments passed to the Job. Arguments should be JSON serializeable if a persistent store is used.
	Args map[string]interface{}

	Policy RunPolicy

	nextRun time.Time
	lastRun time.Time
}

// Next returns the time when the entry should be scheduled next.
func (e Entry) Next(t time.Time) (time.Time, error) {
	schedule, err := defaultParser.Parse(e.Expression)

	if err != nil {
		return time.Time{}, err

	}

	return schedule.Next(t), nil
}

// MustNext calls Next and panics if an error is returned.
func (e Entry) MustNext(t time.Time) time.Time {
	time, err := e.Next(t)
	if err != nil {
		panic(err)
	}
	return time
}

// ByTimeAsc defines ordering for []*Entry
type ByTimeAsc []*Entry

func (b ByTimeAsc) Len() int { return len(b) }
func (b ByTimeAsc) Less(i, j int) bool {
	now := time.Now()
	return b[i].MustNext(now).Before(b[j].MustNext(now))
}
func (b ByTimeAsc) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Unique ensures all items are unique by removing entries with duplicate IDs, this destorys ordering.
func (b ByTimeAsc) Unique() ByTimeAsc {
	set := make(map[string]struct{})
	var uniques = b

	for i, entry := range b {
		if _, exists := set[entry.ID]; exists {
			uniques = b.RemoveCons(i)
			continue
		}
		set[entry.ID] = struct{}{}
	}
	return uniques
}

// RemoveCons removes the ith element in constant time, does not preserves order
func (b ByTimeAsc) RemoveCons(i int) ByTimeAsc {
	tmp := b[:0]
	b[i] = b[len(b)-1] // Copy last element to index i.
	b[len(b)-1] = nil  // Erase last element (write zero value).
	tmp = b[:len(b)-1] // Truncate slice.
	return tmp
}

// RemoveLin removes the ith element in linear time, preserves order
func (b ByTimeAsc) RemoveLin(i int) {
	copy(b[i:], b[i+1:]) // Shift a[i+1:] left one index.
	b[len(b)-1] = nil    // Erase last element (write zero value).
	b = b[:len(b)-1]     // Truncate slice.
}
