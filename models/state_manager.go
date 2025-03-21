package models

import (
	"sync"
)

type ConversationStage string

const (
	ConversationStageNull            ConversationStage = "stage_null"
	ConversationStageAddUserFood     ConversationStage = "stage_add_user_food"
	ConversationStageStart           ConversationStage = "stage_start"
	ConversationStageAddPlanSchedule ConversationStage = "stage_add_plan_schedule"
	ConversationStageAddMacroTargets ConversationStage = "stage_add_macro_targets"
)

type UserState struct {
	Stage ConversationStage
}

// StateManager handles thread-safe state management
type StateManager struct {
	mu         sync.RWMutex
	userStates map[uint]*UserState
}

// NewStateManager creates a new StateManager.
//
// The function initializes the map of user IDs to UserState structs and
// returns a pointer to the new StateManager.
func NewStateManager() *StateManager {
	return &StateManager{
		userStates: make(map[uint]*UserState),
	}
}

// GetUserState retrieves the UserState for a given user ID.
//
// The function takes a StateManager and a user ID as parameters. It locks the
// map of user IDs to UserState structs, retrieves the UserState for the given
// user ID and then unlocks the map.
func (sm *StateManager) GetUserState(userID uint) *UserState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	userState, exists := sm.userStates[userID]

	if !exists {
		return nil
	}

	return userState
}

// SetUserState sets the conversation stage for a given user ID.
//
// The function takes a StateManager, a user ID and a conversation stage as
// parameters. It locks the map of user IDs to UserState structs, sets the
// conversation stage for the given user ID and then unlocks the map.
func (sm *StateManager) SetUserState(userID uint, stage ConversationStage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.userStates[userID] = &UserState{Stage: stage}
}
