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
	Macros          string       `json:"macros"`
	FoodExplanation string       `json:"foodExplanation"`
	FoodCategory    FoodCategory `json:"foodCategory"`
}

// MealPlan represents the response to a meal plan request
type MealPlan struct {
	Plan []MealPlanFood `json:"plan"`
}

// MealPlanExplanation represents the explanation of a meal plan
type MealPlanExplanation struct {
	PlanExplanation string `json:"planExplanation"`
}

// MealPlanPreparation represents the preparation instructions for a meal plan
type MealPlanPreparation struct {
	PlanPreparation PlanPreparation `json:"planPreparation"`
}

// MealPlanShoppingList represents the shopping list for a meal plan
type MealPlanShoppingList struct {
	FoodGroup string   `json:"foodGroup"`
	Foods     []string `json:"foods"`
}

// CompleteMealPlan represents the full meal plan, includes a combination of different types
type CompleteMealPlan struct {
	MealPlan
	MealPlanExplanation
	MealPlanPreparation
	ShoppingList []MealPlanShoppingList `json:"shoppingList"`
}

// FoodCategory represents the category of the food
type FoodCategory string
