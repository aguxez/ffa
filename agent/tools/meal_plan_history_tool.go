package agent

import (
	"context"
	"fmt"

	"github.com/aguxez/ffa/models"
)

type MealPlanHistoryTool struct {
	state *models.StateManager
}

// Name returns the name of the tool.
func (t *MealPlanHistoryTool) Name() string {
	return "meal_plan_history"
}

// Description returns a description of the tool.
func (t *MealPlanHistoryTool) Description() string {
	return `Retrieves historical meal plans with macros and their outcomes.
	This tool is only called once during a conversation.`
}

// Call executes the meal plan history tool.
//
// It retrieves the historical meal plans from the state manager and prints them
// in a human-readable format.
func (t *MealPlanHistoryTool) Call(ctx context.Context, _ string) (string, error) {
	plans := t.state.GetPlans()

	if len(plans) == 0 {
		// Explicit response to prevent confusion
		return `IMPORTANT: No historical meal plans found.
		DO NOT CALL THIS TOOL AGAIN. Create a new meal plan based
		on the available foods and macro targets without historical
		data.`, nil
	}

	result := "Here are the historical meal plans:\n\n"
	// we do not need to check length here as we have already checked
	// above. This won't do anything
	for i, plan := range plans {
		result += fmt.Sprintf("=== PLAN %d ===\n", i+1)

		// Itemize food items
		result += fmt.Sprintf("Item %s\n", plan.Food)
		result += fmt.Sprintf("  Category: %s\n", plan.FoodCategory)
		result += fmt.Sprintf("  Weight: %s\n", plan.Weight)
		result += fmt.Sprintf("  Macros: %s\n", plan.Macros)
		if plan.FoodExplanation != "" {
			result += fmt.Sprintf("  Explanation: %s\n", plan.FoodExplanation)
		}
		result += "\n"

		result += "------------------------\n\n"
	}

	// Add explicit instruction
	result += `IMPORTANT: You now have information about previous meal plans.
	DO NOT call this tool again. Use this to create a new meal plan that
	considers both historical data and the current available foods and
	macro targets.`

	return result, nil
}
