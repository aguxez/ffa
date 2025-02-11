package models

import "sync"

// StateManager handles thread-safe state management
type StateManager struct {
	mu      sync.RWMutex
	foods   []Food
	targets []MacroDay
}

func (s *StateManager) UpdateFoods(foods []Food) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.foods = foods
}

func (s *StateManager) UpdateTargets(targets []MacroDay) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targets = targets
}

func (s *StateManager) GetCurrentState() ([]Food, []MacroDay) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.foods, s.targets
}
