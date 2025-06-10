package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// Meta manages node-local metadata (like incarnation) under a directory.
type Meta struct {
	path        string
	mu          sync.Mutex
	incarnation uint64
}

// NewMeta initializes a Meta instance, loading the current incarnation from disk.
func NewMeta(dir string) (*Meta, error) {
	m := &Meta{path: filepath.Join(dir, "incarnation")}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

// load reads from file (or sets to 1 if missing/invalid).
func (m *Meta) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.path)
	if err != nil {
		m.incarnation = 1
		return nil // treat as fresh
	}
	n, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil || n == 0 {
		m.incarnation = 1
		return nil
	}
	m.incarnation = n
	return nil
}

// NextIncarnation increments, persists, and returns the new value.
func (m *Meta) NextIncarnation() (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.incarnation++
	if err := os.WriteFile(m.path, []byte(fmt.Sprintf("%d", m.incarnation)), 0644); err != nil {
		return 0, err
	}
	return m.incarnation, nil
}

// GetIncarnation returns the current value.
func (m *Meta) GetIncarnation() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.incarnation
}
