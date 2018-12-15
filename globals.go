package cron

import "time"

var executor, _ = New()

// Configure the global executor.
func Configure(opts ...Option) error {
	for _, opt := range opts {
		err := opt(executor)
		if err != nil {
			return err
		}
	}
	return nil
}

// Register a routine with the global executor.
func Register(ID string, routine Routine) error {
	return executor.Register(ID, routine)
}

// Start the global executor.
func Start() error {
	return executor.Start()
}

// Add a new entry to the global crontab
func Add(entry Entry) error {
	return executor.Add(entry)
}

// Stop the global cron executor.
func Stop() {
	executor.Stop()
}

// StopAll stops the global cron executor and all running cronjobs.
func StopAll() {
	executor.StopAll()
}

// CancelAll stops all running cronjobs
func CancelAll() {
	executor.CancelAll()
}

// Remove an entry from the global executor tab. Does not cancel running job nor removes it from the current queue.
func Remove(entry Entry) error {
	return executor.Remove(entry)
}

// GetLog returns the global executor log channel
func GetLog() chan Log {
	return executor.Log()
}

// Location returns the global exectutor timezone locale
func Location() time.Location {
	return executor.Location()
}

// IsRunning returns true if the global executor is running.
func IsRunning() bool {
	return executor.IsRunning()
}

// Err returns the global executor error channel
func Err() chan error {
	return executor.errors
}

// Cancel a specific entry in the global executor while running. If it is not running, Cancel is a noop.
func Cancel(ID string) {
	executor.Cancel(ID)
}
