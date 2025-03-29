package main

import (
	"fmt"

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
	cfg := config.NewAppConfig()
	if cfg == nil {
		log.Error().Msg("Error initializing app config")
		return
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error connecting to database: %v", err.Error()))
	}

	// Initialize state manager
	sm := models.NewStateManager()

	// Initialize a scheduler service
	scheduler := scheduler.NewSchedulerService(db)

	// Setup the cron
	c := cron.New()
	c.AddFunc(getCronTime(cfg), func() { scheduler.Run() })
	c.Start()

	// Start the bot
	if err := bot.Init(sm, db); err != nil {
		log.Err(err).Msg("Error initializing bot")
	}
}

// getCronTime returns the cron time to be used for the scheduler service
// based on the environment
func getCronTime(cfg *config.AppConfig) string {
	if cfg.Environment == "production" {
		return "@hourly"
	}

	// Every minute for development
	return "@every 1m"
}
