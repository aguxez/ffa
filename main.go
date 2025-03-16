package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Dadido3/go-typst"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/filewatch"
	"github.com/aguxez/ffa/models"
)

const (
	generatePlanDuration = 30 * time.Second
)

func main() {
	// Initialize state manager
	sm := &models.StateManager{}

	fileWatcherPaths := []string{"data/foods", "data/targets"}

	// Setup file watcher
	fw, err := filewatch.NewFileWatcher(fileWatcherPaths, sm)
	if err != nil {
		log.Fatalf("error creating file watcher: %v", err)
	}

	// On init, load into memory
	filepath.Walk("data", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Handle file change
		fw.HandleFileChange(path)
		return nil
	})

	// Setup LLM
	llm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
		openai.WithModel("anthropic/claude-3.7-sonnet:beta"),
	)
	if err != nil {
		log.Fatal(err)
	}

	agent := agent.NewNutritionAgent(llm, sm)

	ctx := context.Background()

	log.Print("Generating meal plan...")

	response, err := agent.GenerateMealPlan(ctx)
	if err != nil {
		log.Printf("Error generating meal plan: %v", err)
	}

	planByCategory := make(map[string]string)

	// var planBreakfast string
	// var planLunch string
	// var planDinner string
	// var planDinner
	for _, food := range response.Plan {
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
		= Meal plan #v(1em)

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

		#line(length: 30em)
		Remember to double check and adapt the instructions in this document as it might contain errors or not be extremely accurate for your situation. Specifically in the serving sizes and macros.
		`,
		response.PlanExplanation,
		planByCategory["breakfast"],
		planByCategory["lunch"],
		planByCategory["dinner"],
		planByCategory["snack"],
		response.PlanPreparation.Breakfast,
		response.PlanPreparation.Lunch,
		response.PlanPreparation.Dinner,
	))

	typstCLI := typst.CLI{}

	f, err := os.Create("output.pdf")
	if err != nil {
		log.Fatal("Failed to create file", err)
	}
	defer f.Close()

	if err := typstCLI.Compile(r, f, nil); err != nil {
		log.Fatalf("Failed to compile document: %v.", err)
	}

	// Start file watcher
	go fw.Watch()

	// Start HTTP server
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
