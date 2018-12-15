//go:generate mockgen -destination=tab_mock.go -package=cron github.com/kaiserkarel/go-cron Tab

package cron

import "sync"

// Tab (crontab is short for cron table) defines the storage backend for the scheduler.
type Tab interface {
	// Stores a new entry
	Put(Entry) error

	// Returns all entries
	All() ([]*Entry, error)

	// Removes a single entry
	Remove(Entry) error

	// Clears all entries
	Clear() error
}

// NewMemoryTab returns an in-memory Tab. This is a non-persistent storage.
func NewMemoryTab() *MemoryTab {
	return &MemoryTab{
		mu:      sync.RWMutex{},
		entries: make(map[string]*Entry),
	}
}

// MemoryTab is a simple storage backend.
type MemoryTab struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// Put overrides an entry in the tab.
func (m *MemoryTab) Put(e Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[e.ID] = &e
	return nil
}

// Remove deletes an entry from the tab.
func (m *MemoryTab) Remove(e Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, e.ID)
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
func (m *MemoryTab) All() ([]*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*Entry
	for _, entry := range m.entries {
		res = append(res, entry)
	}
	return res, nil
}
