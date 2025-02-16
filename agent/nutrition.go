package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"

	"github.com/aguxez/ffa/models"
)

// NutritionAgent represents the main agent with memory capabilities
type NutritionAgent struct {
	chain        *chains.LLMChain
	bufferMemory *memory.ConversationWindowBuffer
	state        *models.StateManager
	llm          *openai.LLM
}

type FoodCategory string

const (
	Breakfast FoodCategory = "breakfast"
	Lunch     FoodCategory = "lunch"
	Dinner    FoodCategory = "dinner"
	Snack     FoodCategory = "snack"
)

type MealPlanFood struct {
	Food            string       `json:"food"`
	Weight          string       `json:"weight"`
	Macros          string       `json:"macros"`
	FoodExplanation string       `json:"foodExplanation"`
	FoodCategory    FoodCategory `json:"foodCategory"`
}

type MealPlanResponse struct {
	Plan            []MealPlanFood `json:"plan"`
	PlanExplanation string         `json:"planExplanation"`
	PlanPreparation string         `json:"planPreparation"`
}

func NewNutritionAgent(llm *openai.LLM, state *models.StateManager) *NutritionAgent {
	// Initialize memories. Only remembers last N interactions to keep memory under
	// control
	bufferMem := memory.NewConversationWindowBuffer(5)

	promptTemplate := `
	You are a personal nutritionist with access to historical meal plans and their outcomes.
	Consider the following context when creating a meal plan:

	{{.CombinedInput}}

	Please generate a meal plan that:
	1. Considers the available foods. You are free to suggest similar items, the list
	is to give you context in what foods I like.
	2. Meets the macro targets as much as possible
	3. Considers previous successful meal plans
	4. Includes 5 meals per day with one protein shake being one of them
	5. Is suitable for meal prep (same meals daily)

	Provide portions in grams and include estimated macros per meal if possible.

	For your response, prefer to give portions and weights for all weekdays instead
	of per day while breaking down macros per day. For example: "chicken breast 800g (160 per serving)"

	Stick to this JSON format for the output.

	{
		"plan": [
			{
				"food": string, // The food name
				"weight": string, // The weight of the food in grams
				"macros": string // Protein, fat, carbs, and calories,
				"foodExplanation": string // Why this food was chosen over alternatives,
				"foodCategory": string // The time of the meal
			}
		],
		"planExplanation": string // Explanation of the plan in markdown.
		"planPreparation": string // Explanation of the plan preparation in markdown.
	}

	"foodCategory" must be one of breakfast, lunch, dinner, or snack.

	Ensure for "planPreparation" you solely use the foods in the "plan".

	Make the "planExplanation" and "planPreparation" compatible with some color!
	Ensure the plan is formatted in a way that's displayable in a terminal.
	`

	// Create chain with memory
	chain := chains.NewLLMChain(
		llm,
		prompts.NewPromptTemplate(promptTemplate, []string{"CombinedInput"}),
	)

	return &NutritionAgent{
		chain:        chain,
		bufferMemory: bufferMem,
		state:        state,
		llm:          llm,
	}
}

func (n *NutritionAgent) GenerateMealPlan(ctx context.Context) (MealPlanResponse, error) {
	foods, targets := n.state.GetCurrentState()

	// Get conversation history
	history, err := n.bufferMemory.LoadMemoryVariables(ctx, map[string]any{})
	if err != nil {
		return MealPlanResponse{}, fmt.Errorf("loading memory variables: %w", err)
	}

	// Combine inputs
	combinedInput := fmt.Sprintf("Foods: %v\nTargets: %v\nHistory: %v", foods, targets, history["history"])

	// Create input with context
	input := map[string]interface{}{
		"CombinedInput": combinedInput,
	}

	// Generate plan
	result, err := chains.Call(ctx, n.chain, input)
	if err != nil {
		return MealPlanResponse{}, fmt.Errorf("calling chain: %w", err)
	}

	// Store interaction in memory
	// TODO: Differentiate from accepted and rejected plans
	err = n.bufferMemory.SaveContext(ctx, input, result)
	if err != nil {
		log.Printf("Error saving to memory: %v", err)
	}

	responseText := result["text"].(string)
	responseText = replaceLineBreaks(responseText)

	var parsedResponse MealPlanResponse
	err = json.Unmarshal([]byte(responseText), &parsedResponse)
	if err != nil {
		return MealPlanResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return parsedResponse, nil
}

func replaceLineBreaks(s string) string {
	noLineBreaks := strings.ReplaceAll(s, "\n", "")
	noJsonMarkupStart := strings.ReplaceAll(noLineBreaks, "```json", "")
	noJsonMarkupEnd := strings.ReplaceAll(noJsonMarkupStart, "```", "")

	return noJsonMarkupEnd
}
