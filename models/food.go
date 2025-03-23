package models

type Food struct {
	Name string
}

type PlanPreparation struct {
	Breakfast string `json:"breakfast"`
	Lunch     string `json:"lunch"`
	Dinner    string `json:"dinner"`
}

// MealPlanFood represents a food in the meal plan
type MealPlanFood struct {
	Food            string       `json:"food"`
	Weight          string       `json:"weight"`
	Macros          string       `json:"macros"`
	FoodExplanation string       `json:"foodExplanation"`
	FoodCategory    FoodCategory `json:"foodCategory"`
}

// MealPlanResponse represents the response to a meal plan request
type MealPlanResponse struct {
	Plan            []MealPlanFood `json:"plan"`
	PlanExplanation string         `json:"planExplanation"`

	// TODO: Separate this into another LLM call for better precision
	PlanPreparation PlanPreparation `json:"planPreparation"`
}

// MealPlanShoppingList represents the shopping list for a meal plan
type MealPlanShoppingList struct {
	FoodGroup string   `json:"foodGroup"`
	Foods     []string `json:"foods"`
}

// CompleteMealPlan represents the full meal plan, includes a combination of different types
type CompleteMealPlan struct {
	MealPlanResponse
	ShoppingList []MealPlanShoppingList `json:"shoppingList"`
}

// FoodCategory represents the category of the food
type FoodCategory string
