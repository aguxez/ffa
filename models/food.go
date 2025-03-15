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
	Plan            []MealPlanFood  `json:"plan"`
	PlanExplanation string          `json:"planExplanation"`
	PlanPreparation PlanPreparation `json:"planPreparation"`
}

// FoodCategory represents the category of the food
type FoodCategory string
