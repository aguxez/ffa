package main

import (
	"fmt"
	"net/http"

	"github.com/aguxez/ffa/bot"
	"github.com/aguxez/ffa/config"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	"github.com/aguxez/ffa/scheduler"
	cron "github.com/robfig/cron"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Initialize app config
	cfg, err := config.NewAppConfig()
	if err != nil {
		log.Err(err).Msg("Error initializing app config")
		return
	}

	// Start server for healthcheck
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error connecting to database: %v", err.Error()))
	}

	// Initialize state manager
	sm := models.NewStateManager()

	botService, err := bot.NewBotService(sm, db, cfg)
	if err != nil {
		log.Err(err).Msg("Error creating bot service")
	}

	// Initialize a scheduler service
	scheduler := scheduler.NewSchedulerService(db, botService)

	// Setup the cron
	c := cron.New()
	err = c.AddFunc(getCronTime(cfg), func() { scheduler.Run() })
	if err != nil {
		log.Err(err).Msg("Error adding cron job")
		return
	}

	c.Start()

	// Start the bot and block main thread
	if err := botService.Init(); err != nil {
		log.Err(err).Msg("Error initializing bot")
	}
}

// getCronTime returns the cron time to be used for the scheduler service
// based on the environment
func getCronTime(cfg config.AppConfig) string {
	if cfg.Environment == "production" {
		return "@hourly"
	}

	// Every minute for development
	return "@every 1m"
}
