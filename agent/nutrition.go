package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"

	"github.com/aguxez/ffa/database"
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

// GenerateMealPlan generates a meal plan based on the available foods and macro targets
//
// It takes a list of available foods and macro targets, formats them into a structured
// prompt, and calls the LLM to generate a meal plan.
//
// The function will return an error if there was an issue with the LLM call or unmarshalling
// the response.
func (n *NutritionAgent) GenerateMealPlan(ctx context.Context, foods []database.UserFood) (models.MealPlanResponse, error) {
	macrosDays := make([]models.MacroDay, 0)

	// More structured formatting of input to help the LLM
	foodsStr := "Available Foods:\n"
	for _, food := range foods {
		foodsStr += fmt.Sprintf("- %s\n", food.Name)
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
	promptTemplate := prompts.NewPromptTemplate(NewNutritionPromptTemplate(), []string{"CombinedInput"})
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

	jsonStr := ExtractJSONFromResponse(output)

	var parsedResponse models.MealPlanResponse
	err = json.Unmarshal([]byte(jsonStr), &parsedResponse)
	if err != nil {
		return models.MealPlanResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return parsedResponse, nil
}
