package engine

import (
	"sync"
	"time"
)

type StepRecord struct {
	Status    StepStatus
	UpdatedAt time.Time
}

type StateManager struct {
	mu     sync.RWMutex
	states map[string]StepRecord
}

func NewStateManager() *StateManager {
	return &StateManager{states: make(map[string]StepRecord)}
}

func (s *StateManager) Set(id string, status StepStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[id] = StepRecord{
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	}
}

func (s *StateManager) Get(id string) StepStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if rec, ok := s.states[id]; ok {
		return rec.Status
	}
	return StepPending
}

// Snapshot returns a copy of the current state map (statuses only).
func (s *StateManager) Snapshot() map[string]StepStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]StepStatus, len(s.states))
	for k, v := range s.states {
		out[k] = v.Status
	}
	return out
}
