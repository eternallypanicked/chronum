package engine

import (
    "sync"
)

type StateManager struct {
    mu     sync.RWMutex
    states map[string]StepStatus
}

func NewStateManager() *StateManager {
    return &StateManager{states: make(map[string]StepStatus)}
}

func (s *StateManager) Set(id string, status StepStatus) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.states[id] = status
}

func (s *StateManager) Get(id string) StepStatus {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.states[id]
}

func (s *StateManager) Snapshot() map[string]StepStatus {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make(map[string]StepStatus)
    for k, v := range s.states {
        out[k] = v
    }
    return out
}
