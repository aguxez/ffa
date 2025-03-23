package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
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

	// Appending macros
	macrosStr := appendMacros(user)

	// Append dietary restrictions if any
	var dietaryRestrictionsStr string
	if user.DietaryRestrictions != nil {
		dietaryRestrictionsStr += "Dietary Restrictions:\n"
		for _, restriction := range *user.DietaryRestrictions {
			dietaryRestrictionsStr += fmt.Sprintf("- %s\n", restriction.Name)
		}
	}

	// Append user's body weight if set
	var weightStr string
	if user.Weight != nil {
		weightStr += fmt.Sprintf("Body Weight (kg): %d\n", *user.Weight)
	}

	combinedInput := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		foodsStr,
		macrosStr,
		dietaryRestrictionsStr,
		weightStr,
	)

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
		log.Err(err).Msg("Error calling LLM for plan generation")
		return models.MealPlanResponse{}, err
	}

	jsonStr := ExtractJSONFromResponse(output)

	var parsedResponse models.MealPlanResponse
	err = json.Unmarshal([]byte(jsonStr), &parsedResponse)
	if err != nil {
		log.Debug().Msgf("Raw JSON: %s", jsonStr)
		return models.MealPlanResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return parsedResponse, nil
}

// appendMacros formats and returns a string representing the user's macronutrient targets.
// It includes protein, carbs, fat, and calorie information if they are set.
// The function will append a title "Macro Targets:" if any of these values are present,
// followed by each individual macronutrient and calorie count formatted as a list.
func appendMacros(user *database.User) string {
	var macrosStr string
	if user.Protein != nil ||
		user.Fat != nil ||
		user.Carbs != nil ||
		user.Calories == nil {

		// Only append title if there is a value set
		macrosStr = "Macro Targets:\n"
	}
	if user.Protein != nil {
		macrosStr += fmt.Sprintf("- Protein (grams): %d\n", *user.Protein)
	}
	if user.Carbs != nil {
		macrosStr += fmt.Sprintf("- Carbs (grams): %d\n", *user.Carbs)
	}
	if user.Fat != nil {
		macrosStr += fmt.Sprintf("- Fat (grams): %d\n", *user.Fat)
	}
	if user.Calories != nil {
		macrosStr += fmt.Sprintf("- Calories: %d\n", *user.Calories)
	}
	return macrosStr
}
