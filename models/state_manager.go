package models

import (
	"sync"
)

// StateManager handles thread-safe state management
type StateManager struct {
	mu      sync.RWMutex
	foods   []Food
	targets []MacroDay
	plans   []MealPlanFood
}

// SaveMealPlan saves a meal plan by appending it to the stored plans.
// This method is thread-safe and ensures that concurrent accesses to
// the plans slice do not cause data races.

func (s *StateManager) SaveMealPlan(plan MealPlanFood) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans = append(s.plans, plan)
}

// UpdateFoods updates the list of foods in the state manager.
// This method is thread-safe and ensures that concurrent accesses to
// the foods slice do not cause data races.
func (s *StateManager) UpdateFoods(foods []Food) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.foods = foods
}

// UpdateTargets updates the list of targets in the state manager.
// This method is thread-safe and ensures that concurrent accesses to
// the targets slice do not cause data races.
func (s *StateManager) UpdateTargets(targets []MacroDay) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targets = targets
}

// GetFoodsAndMacros retrieves the current state (foods and targets) from the manager.
// This method is thread-safe and returns a copy of the current state.
func (s *StateManager) GetFoodsAndMacros() ([]Food, []MacroDay) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.foods, s.targets
}

// GetPlans retrieves the current list of meal plans stored in the state manager.
// This method is thread-safe and returns a copy of the plans slice, ensuring
// that concurrent accesses do not cause data races.
func (s *StateManager) GetPlans() []MealPlanFood {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plans
}
