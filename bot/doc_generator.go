package bot

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Dadido3/go-typst"
	"github.com/aguxez/ffa/models"
	"github.com/rs/zerolog/log"
)

// GenerateDocFromMealPlan generates a Typst document from a meal plan response.
//
// It takes the meal plan response and creates a Typst document with a table for each
// meal category (breakfast, lunch, dinner, snack) and a section for the preparation
// instructions. We use Typst so we can easily export to PDF and send it to chats.
func GenerateDocFromMealPlan(mealPlan models.CompleteMealPlan) (bytes.Buffer, error) {
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

	// Generate shopping list content dynamically
	var shoppingListContent strings.Builder
	shoppingListContent.WriteString(`
		#block(
			width: 100%,
			fill: rgb("#f8fafc"),
			stroke: 0.5pt,
			breakable: false,
			radius: 4pt,
			[
				#pad(
					x: 1cm,
					y: 0.7cm,
					[
						#grid(
							columns: (1fr, 1fr),
							gutter: 2em,`)

	// Use a consistent color for all food groups
	getFoodGroupColor := func(index int) string {
		// Array of _pleasant_ colors for food groups
		colors := []string{
			"#065f46", // Green
			"#9f1239", // Red
			"#1e40af", // Blue
			"#713f12", // Brown
			"#7e22ce", // Purple
			"#0e7490", // Teal
			"#b45309", // Amber
			"#4d7c0f", // Lime
		}

		return colors[index%len(colors)]
	}

	// Add each food group to the grid
	for i, group := range mealPlan.ShoppingList {
		shoppingListContent.WriteString(fmt.Sprintf(`
							[
								#text(weight: "bold", fill: rgb("%s"))[%s]`,
			getFoodGroupColor(i),
			group.FoodGroup))

		for _, food := range group.Foods {
			// Note here we preserve the newline since we use a backtick string. It's important
			// they start in a newline otherwise the list is not correctly formatted
			shoppingListContent.WriteString(fmt.Sprintf(`
								- %s`, food))
		}

		shoppingListContent.WriteString(`
							],`)
	}

	// Close the grid and other tags
	shoppingListContent.WriteString(`
						)
					]
				)
			]
		)`)

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
		== Shopping List
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
		shoppingListContent.String(),
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
