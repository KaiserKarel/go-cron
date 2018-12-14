//go:generate mockgen -destination=tab_mock.go -package=cron github.com/kaiserkarel/go-cron Tab

package cron

import "sync"

// Tab (crontab is short for cron table) defines the storage backend for the scheduler.
type Tab interface {
	// Stores a new entry
	Put(*Entry) error

	// Returns all entries
	All() ([]*Entry, error)

	// Removes a single entry
	Remove(*Entry) error

	// Clears all entries
	Clear() error
}

func NewMemoryTab() *MemoryTab {
	return &MemoryTab{
		mu:      sync.RWMutex{},
		entries: make(map[string]*Entry),
	}
}

type MemoryTab struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

func (m *MemoryTab) Put(e *Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[e.ID] = e
	return nil
}

func (m *MemoryTab) Remove(e *Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, e.ID)
	return nil
}

func (m *MemoryTab) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[string]*Entry)
	return nil
}

func (m *MemoryTab) All() ([]*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*Entry
	for _, entry := range m.entries {
		res = append(res, entry)
	}
	return res, nil
}
