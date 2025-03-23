package bot

import (
	"bytes"
	"fmt"

	"github.com/Dadido3/go-typst"
	"github.com/aguxez/ffa/models"
	"github.com/rs/zerolog/log"
)

// GenerateDocFromMealPlan generates a Typst document from a meal plan response.
//
// It takes the meal plan response and creates a Typst document with a table for each
// meal category (breakfast, lunch, dinner, snack) and a section for the preparation
// instructions. We use Typst so we can easily export to PDF and send it to chats.
func GenerateDocFromMealPlan(mealPlan models.MealPlanResponse) (bytes.Buffer, error) {
	planByCategory := make(map[string]string)
	for _, food := range mealPlan.Plan {
		formattedItem := fmt.Sprintf(
			`
				"%s", "%s", "%s", "%s",
				`, food.Food, food.Weight, food.Macros, food.FoodExplanation,
		)

		switch food.FoodCategory {
		case "breakfast":
			planByCategory["breakfast"] += formattedItem

		case "lunch":
			planByCategory["lunch"] += formattedItem

		case "dinner":
			planByCategory["dinner"] += formattedItem

		case "snack":
			planByCategory["snack"] += formattedItem
		}

	}

	r := bytes.NewBufferString(fmt.Sprintf(
		`
		#set text(font: "Inter")
		#grid(
			columns: (auto, auto),
			gutter: 1em,
			align: (horizon, horizon),
			[= Meal Plan],
			[#image("assets/bot-logo.svg", height: 1.2em)]
		)

		%s

		== Breakfast
		#table(
			columns: (auto, auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Weight*], [*Macros*], [*Explanation*]
			),
			%s
		)

		== Lunch

		#table(
			columns: (auto, auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Weight*], [*Macros*], [*Explanation*]
			),
			%s
		)

		== Dinner
		#table(
			columns: (auto, auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Weight*], [*Macros*], [*Explanation*]
			),
			%s
		)

		== Snack
		#table(
			columns: (auto, auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Weight*], [*Macros*], [*Explanation*]
			),
			%s
		)

		== Preparation
		=== Breakfast
		%s

		=== Lunch
		%s

		=== Dinner
		%s
		`,
		mealPlan.PlanExplanation,
		planByCategory["breakfast"],
		planByCategory["lunch"],
		planByCategory["dinner"],
		planByCategory["snack"],
		mealPlan.PlanPreparation.Breakfast,
		mealPlan.PlanPreparation.Lunch,
		mealPlan.PlanPreparation.Dinner,
	))

	typstCLI := typst.CLI{}

	// Create a buffer to hold the file state in memory
	var buf bytes.Buffer

	if err := typstCLI.Compile(r, &buf, nil); err != nil {
		log.Err(err).Msg("Failed to compile document")
		return bytes.Buffer{}, err
	}

	return buf, nil
}
