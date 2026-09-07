package store

import (
	"errors"
	"sort"
	"sync"
)

// ErrFlagExists is returned by Add when a flag with the same key already exists.
var ErrFlagExists = errors.New("flag already exists")

// Flag is a single feature flag.
type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// Store is a thread-safe in-memory feature-flag store.
type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{flags: make(map[string]Flag)}
}

// Add inserts a flag. It returns ErrFlagExists if the key already exists.
func (s *Store) Add(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrFlagExists
	}
	s.flags[f.Key] = f
	return nil
}

// List returns all flags sorted by key.
func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// Get returns the flag for key and whether it was found.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

// Update replaces enabled, description and rollout for key. It returns the
// updated flag and whether the key existed.
func (s *Store) Update(key string, enabled bool, description string, rollout int) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.flags[key]
	if !ok {
		return Flag{}, false
	}
	f.Enabled = enabled
	f.Description = description
	f.RolloutPercent = rollout
	s.flags[key] = f
	return f, true
}

// Delete removes the flag for key and reports whether it existed.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
