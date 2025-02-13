package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/llms/openai"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/api"
	"github.com/aguxez/ffa/filewatch"
	"github.com/aguxez/ffa/models"
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
		openai.WithModel("deepseek/deepseek-r1-distill-llama-70b"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create agent
	agent := agent.NewNutritionAgent(llm, sm)

	// Setup HTTP server
	http.HandleFunc("/mealplan", func(w http.ResponseWriter, r *http.Request) {
		api.HandleMealPlanRequest(agent, w, r)
	})

	// Start file watcher
	go fw.Watch()

	// Start HTTP server
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
