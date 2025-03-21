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
func (n *NutritionAgent) GenerateMealPlan(ctx context.Context, user *database.User) (models.MealPlanResponse, error) {
	// More structured formatting of input to help the LLM
	foodsStr := "Available Foods:\n"
	for _, food := range user.Foods {
		foodsStr += fmt.Sprintf("- %s\n", food.Name)
	}

	var targetsStr string
	targetsStr = "Macro Targets:\n"

	// Using 'any' here because we have different types for the values
	appendIfExists := func(name string, value any) {
		if value != nil {
			targetsStr += fmt.Sprintf("- %s: %d g\n", name, value)
		}
	}

	appendIfExists("Protein", user.Protein)
	appendIfExists("Carbs", user.Carbs)
	appendIfExists("Fat", user.Fat)
	appendIfExists("Calories", user.Calories)

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
