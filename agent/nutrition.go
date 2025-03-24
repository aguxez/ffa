package agent

import (
	"context"
	"errors"
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

// GenerateMealPlan generates a complete meal plan for a given user. It first asks the
// Large Language Model to generate a meal plan based on the user's preferences and
// body weight. Then it creates a shopping list for this plan. The user is expected
// to have a non-empty list of preferred foods and a non-nil weight.
func (n *NutritionAgent) GenerateMealPlan(ctx context.Context, user *database.User) (models.CompleteMealPlan, error) {
	// First ask for a meal plan
	mealResponse, err := n.generateMealPlanFromLLM(ctx, user)
	if err != nil {
		return models.CompleteMealPlan{}, err
	}

	// Then ask for a meal plan explanation
	explanationResponse, err := n.generateMealPlanExplanationFromLLM(ctx, mealResponse)
	if err != nil {
		return models.CompleteMealPlan{}, err
	}

	// Then ask for a meal plan preparation
	preparationResponse, err := n.generateMealPlanPreparationFromLLM(ctx, mealResponse)
	if err != nil {
		return models.CompleteMealPlan{}, err
	}

	// Then create the shopping list for this plan
	shoppingList, err := n.generateMealPlanShoppingListfromLLM(ctx, mealResponse.Plan)
	if err != nil {
		return models.CompleteMealPlan{}, err
	}

	return models.CompleteMealPlan{
		MealPlan:            mealResponse,
		MealPlanExplanation: explanationResponse,
		MealPlanPreparation: preparationResponse,
		ShoppingList:        shoppingList,
	}, nil
}

// generateMealPlanFromLLM generates a meal plan for a given user based on their available foods,
// macro targets, dietary restrictions, and body weight.
//
// It formats the input into a structured prompt and calls the LLM to generate a meal plan.
//
// The function will return an error if there was an issue with the LLM call or unmarshalling the response.
func (n *NutritionAgent) generateMealPlanFromLLM(ctx context.Context, user *database.User) (models.MealPlan, error) {
	log.Debug().Msg("Generating meal plan from LLM")

	// More structured formatting of input to help the LLM
	foodsStr := "Foods the user likes:\n"
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

	// Append the number of days the plan should be for
	var planDurationDays string
	if user.PlanSchedule != nil {
		planDurationDays += fmt.Sprintf("Plan Duration (days): %d\n", user.PlanSchedule.IntervalDays)
	}

	combinedInput := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		foodsStr,
		macrosStr,
		dietaryRestrictionsStr,
		weightStr,
		planDurationDays,
	)

	promptTemplate := prompts.NewPromptTemplate(NewNutritionPromptTemplate(), []string{"CombinedInput"})
	promptInput := map[string]any{
		"CombinedInput": combinedInput,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return models.MealPlan{}, fmt.Errorf("formatting prompt: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Temperature is overridden to promote more diversity in answers for a given input
	output, err := n.llm.Call(ctxWithTimeout, prompt, llms.WithTemperature(1.5))
	if err != nil {
		log.Err(err).Msg("Error calling LLM for plan generation")
		return models.MealPlan{}, err
	}

	return ParseJSONFromResponse[models.MealPlan](output)
}

// generateMealPlanExplanationFromLLM generates a comprehensive explanation for a given meal plan.
// It formats the input meal plan into a structured prompt and calls the LLM to generate the explanation.
// The function will return an error if there was an issue with the LLM call or unmarshalling the response.
func (n *NutritionAgent) generateMealPlanExplanationFromLLM(ctx context.Context, mealPlan models.MealPlan) (models.MealPlanExplanation, error) {
	log.Debug().Msg("Generating meal plan explanation from LLM")

	promptTemplate := prompts.NewPromptTemplate(NewMealPlanExplanationTemplate(), []string{"PlanFromChain1"})
	promptInput := map[string]any{
		"PlanFromChain1": mealPlan.Plan,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return models.MealPlanExplanation{}, fmt.Errorf("formatting prompt: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Temperature is overridden to promote more diversity in answers for a given input
	output, err := n.llm.Call(ctxWithTimeout, prompt, llms.WithModel(string(LLMModelGemini)))
	if err != nil {
		log.Err(err).Msg("Error calling LLM for plan explanation")
		return models.MealPlanExplanation{}, err
	}

	return ParseJSONFromResponse[models.MealPlanExplanation](output)
}

func (n *NutritionAgent) generateMealPlanPreparationFromLLM(ctx context.Context, mealPlan models.MealPlan) (models.MealPlanPreparation, error) {
	log.Debug().Msg("Generating meal plan preparation from LLM")

	promptTemplate := prompts.NewPromptTemplate(NewMealPlanPreparationTemplate(), []string{"PlanFromChain1"})
	promptInput := map[string]any{
		"PlanFromChain1": mealPlan.Plan,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return models.MealPlanPreparation{}, fmt.Errorf("formatting prompt: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Temperature is overridden to promote more diversity in answers for a given input
	output, err := n.llm.Call(ctxWithTimeout, prompt, llms.WithModel(string(LLMModelGemini)))
	if err != nil {
		log.Err(err).Msg("Error calling LLM for plan explanation")
		return models.MealPlanPreparation{}, err
	}

	return ParseJSONFromResponse[models.MealPlanPreparation](output)
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

// generateMealPlanShoppingListfromLLM generates a shopping list for a given meal plan.
// It formats the foods from the plan and prompts the LLM to generate a shopping list.
// The function will return an error if there was an issue with the LLM call or unmarshalling the response.
func (n *NutritionAgent) generateMealPlanShoppingListfromLLM(ctx context.Context, plan []models.MealPlanFood) ([]models.MealPlanShoppingList, error) {
	log.Debug().Msg("Creating meal plan shopping list")
	if len(plan) == 0 {
		return []models.MealPlanShoppingList{}, errors.New("plan is empty")
	}

	// Format the foods from the plan
	foodsStr := "Foods:\n"
	for _, food := range plan {
		foodsStr += fmt.Sprintf("- %s\n", food.Food)
	}

	combinedInput := fmt.Sprintf("%s\n", foodsStr)

	promptTemplate := prompts.NewPromptTemplate(NewMealPlanShoppingListTemplate(), []string{"Input"})
	promptInput := map[string]any{
		"Input": combinedInput,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return []models.MealPlanShoppingList{}, fmt.Errorf("formatting prompt: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Override the model we setup initially for a fast and more affordable option
	// since answers are more structured and _easier_ to respond to.
	output, err := n.llm.Call(ctxWithTimeout, prompt, llms.WithModel(string(LLMModelGemini)))
	if err != nil {
		log.Err(err).Msg("Error calling LLM for plan generation")
		return []models.MealPlanShoppingList{}, err
	}

	return ParseJSONFromResponse[[]models.MealPlanShoppingList](output)
}
