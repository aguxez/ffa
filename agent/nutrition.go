package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"

	"github.com/aguxez/ffa/models"
)

// NutritionAgent represents the main agent with memory capabilities
type NutritionAgent struct {
	state *models.StateManager
	llm   *openai.LLM
}

const (
	Breakfast models.FoodCategory = "breakfast"
	Lunch     models.FoodCategory = "lunch"
	Dinner    models.FoodCategory = "dinner"
	Snack     models.FoodCategory = "snack"
)

// NewNutritionAgent creates a new nutrition agent
func NewNutritionAgent(llm *openai.LLM, state *models.StateManager) *NutritionAgent {
	return &NutritionAgent{
		state: state,
		llm:   llm,
	}
}

// GenerateMealPlan generates a meal plan given current foods and targets.
// It also considers the previous conversations and meal plans.
// The generated plan is a JSON object with pre defined keys.
func (n *NutritionAgent) GenerateMealPlan(ctx context.Context) (models.MealPlanResponse, error) {
	// foods, macrosDays := n.state.GetFoodsAndMacros()
	foods := []string{}
	macrosDays := make([]models.MacroDay, 0)

	// More structured formatting of input to help the LLM
	foodsStr := "Available Foods:\n"
	for _, food := range foods {
		foodsStr += fmt.Sprintf("- %s\n", food)
	}

	var targetsStr string
	if len(macrosDays) > 0 {
		targetsStr := "Macro Targets:\n"
		targetsStr += fmt.Sprintf("- Protein: %d g\n", macrosDays[0].Target.Protein)
		targetsStr += fmt.Sprintf("- Carbs: %d g\n", macrosDays[0].Target.Carbs)
		targetsStr += fmt.Sprintf("- Fat: %d g\n", macrosDays[0].Target.Fat)
		targetsStr += fmt.Sprintf("- Calories: %d\n", macrosDays[0].Target.Calories)
	}

	combinedInput := fmt.Sprintf("%s\n%s", foodsStr, targetsStr)
	promptTemplate := prompts.NewPromptTemplate(NewPromptTemplate(), []string{"CombinedInput"})
	promptInput := map[string]any{
		"CombinedInput": combinedInput,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return models.MealPlanResponse{}, fmt.Errorf("formatting prompt: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Temperature is overridden to promote more diversity in answers for a given input
	output, err := n.llm.Call(ctxWithTimeout, prompt, llms.WithTemperature(1.5))
	if err != nil {
		return models.MealPlanResponse{}, err
	}

	jsonStr := extractJSON(output)

	var parsedResponse models.MealPlanResponse
	err = json.Unmarshal([]byte(jsonStr), &parsedResponse)
	if err != nil {
		return models.MealPlanResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return parsedResponse, nil
}

// extractJSON finds and extracts JSON content from a string by removing the markup
func extractJSON(s string) string {
	var final string

	final = strings.ReplaceAll(s, "\n", "")
	final = strings.ReplaceAll(final, "```json", "")
	final = strings.ReplaceAll(final, "```", "")

	return final
}
