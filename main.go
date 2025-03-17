package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aguxez/ffa/bot"
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
	// Start file watcher
	go fw.Watch()

	// Start the bot
	go func() {
		if err := bot.Init(sm); err != nil {
			log.Fatal(err)
		}
	}()

	// Start HTTP server
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
