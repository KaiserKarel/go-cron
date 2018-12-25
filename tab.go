package cron

import "sync"

type UpdateType int

const (
	NewAltered UpdateType = 0
	Deleted    UpdateType = 1
)

// Tab (crontab is short for cron table) defines the storage backend for the scheduler.
type Tab interface {
	// Stores a new entry
	Put(Entry) error

	// Returns all entries
	All() (map[string]*Entry, error)

	// Removes a single entry
	Remove(ID string) error

	// Update a single entry.
	//
	// ID is the ID of the current entry. If the ID does not match the provided entry.ID, the ID will be updated as well.
	Update(ID string, entry Entry) error

	// Clears all entries
	Clear() error

	// Refresh provides any new or updated entries. The executor selects on Refresh to update it's own internal queue.
	//
	// Do not read of Tab.Refresh as package user; instead listen on the Refresh method of executor if you wish to obtain
	// updated/new entries.
	Refresh() <-chan Update
}

// NewMemoryTab returns an in-memory Tab. This is a non-persistent storage.
func NewMemoryTab() *MemoryTab {
	return &MemoryTab{
		mu:      sync.RWMutex{},
		entries: make(map[string]*Entry),
		refresh: make(chan Update, 1000),
	}
}

// MemoryTab is a simple storage backend.
type MemoryTab struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	refresh chan Update
}

// Put overrides an entry in the tab.
func (m *MemoryTab) Put(e Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[e.ID] = &e

	m.refresh <- Update{NewAltered, e}
	return nil
}

// Remove deletes an entry from the tab.
func (m *MemoryTab) Remove(ID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, ID)
	m.refresh <- Update{Deleted, Entry{ID: ID}}
	return nil
}

// Clear deletes all entries from the tab.
func (m *MemoryTab) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[string]*Entry)
	return nil
}

// All returns all entries from the tab.
func (m *MemoryTab) All() (map[string]*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.entries, nil
}

// Refresh returns update on the tab contents
func (m *MemoryTab) Refresh() <-chan Update {
	return m.refresh
}

// Update upserts an entry
func (m *MemoryTab) Update(ID string, e Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[ID] = &e
	m.refresh <- Update{NewAltered, Entry{ID: ID}}
	return nil
}
