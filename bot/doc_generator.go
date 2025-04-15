package bot

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Dadido3/go-typst"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	"github.com/rs/zerolog/log"
)

// GenerateDocFromMealPlan generates a Typst document from a meal plan response.
//
// It takes the meal plan response and creates a Typst document with a table for each
// meal category (breakfast, lunch, dinner, snack) and a section for the preparation
// instructions. We use Typst so we can easily export to PDF and send it to chats.
func GenerateDocFromMealPlan(mealPlan models.CompleteMealPlan, userLanguage database.UserLanguage) (bytes.Buffer, error) {
	planByCategory := make(map[string]string)
	for _, food := range mealPlan.Plan {
		formattedItem := fmt.Sprintf(
			`
				"%s", "%s", "%s",
				`, food.Food, food.Macros, food.FoodExplanation,
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
		// Array of *pleasant* colors for food groups
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

	// Define the template content
	tmplContent := `
		#set text(font: "Inter")
		#grid(
			columns: (auto, auto),
			gutter: 1em,
			align: (horizon, horizon),
			[= Meal Plan],
			[#image("assets/bot-logo.svg", height: 1.2em)]
		)
		{{.PlanExplanation}}
		== {{.BreakfastTitle}}
		#table(
			columns: (auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Macros*], [*Explanation*]
			),
			{{.Breakfast}}
		)
		== {{.LunchTitle}}
		#table(
			columns: (auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Macros*], [*Explanation*]
			),
			{{.Lunch}}
		)
		== {{.DinnerTitle}}
		#table(
			columns: (auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Macros*], [*Explanation*]
			),
			{{.Dinner}}
		)
		== {{.SnackTitle}}
		#table(
			columns: (auto, auto, auto),
			align: horizon,
			table.header(
				[*Food*], [*Macros*], [*Explanation*]
			),
			{{.Snack}}
		)
    == {{.PreparationTitle}}
    === {{.BreakfastTitle}}
		{{.BreakfastPrep}}
		=== {{.LunchTitle}}
		{{.LunchPrep}}
		=== {{.DinnerTitle}}
		{{.DinnerPrep}}
		== {{.ShoppingListTitle}}
		{{.ShoppingList}}
	`

	// Create and parse the template
	tmpl, err := template.New("mealplan").Parse(tmplContent)
	if err != nil {
		log.Err(err).Msg("Failed to parse template")
		return bytes.Buffer{}, err
	}

	// Define the data to populate the template
	data := struct {
		PlanExplanation   string
		Breakfast         string
		Lunch             string
		Dinner            string
		Snack             string
		BreakfastPrep     string
		LunchPrep         string
		DinnerPrep        string
		ShoppingList      string
		BreakfastTitle    string
		LunchTitle        string
		DinnerTitle       string
		SnackTitle        string
		PreparationTitle  string
		ShoppingListTitle string
	}{
		PlanExplanation:   mealPlan.PlanExplanation,
		Breakfast:         planByCategory["breakfast"],
		Lunch:             planByCategory["lunch"],
		Dinner:            planByCategory["dinner"],
		Snack:             planByCategory["snack"],
		BreakfastPrep:     mealPlan.PlanPreparation.Breakfast,
		LunchPrep:         mealPlan.PlanPreparation.Lunch,
		DinnerPrep:        mealPlan.PlanPreparation.Dinner,
		ShoppingList:      shoppingListContent.String(),
		BreakfastTitle:    i18n.GetText(i18n.KeyMealPlanDocumentBreakfast, userLanguage),
		LunchTitle:        i18n.GetText(i18n.KeyMealPlanDocumentLunch, userLanguage),
		DinnerTitle:       i18n.GetText(i18n.KeyMealPlanDocumentDinner, userLanguage),
		SnackTitle:        i18n.GetText(i18n.KeyMealPlanDocumentSnack, userLanguage),
		PreparationTitle:  i18n.GetText(i18n.KeyMealPlanDocumentPreparation, userLanguage),
		ShoppingListTitle: i18n.GetText(i18n.KeyMealPlanDocumentShoppingList, userLanguage),
	}

	// Execute the template and write it to a buffer
	var documentBuffer bytes.Buffer
	if err := tmpl.Execute(&documentBuffer, data); err != nil {
		log.Err(err).Msg("Failed to execute template")
		return bytes.Buffer{}, err
	}

	typstCLI := typst.CLI{}
	// Create a buffer to hold the file state in memory
	var buf bytes.Buffer
	if err := typstCLI.Compile(&documentBuffer, &buf, nil); err != nil {
		log.Err(err).Msg("Failed to compile document")
		return bytes.Buffer{}, err
	}

	return buf, nil
}
