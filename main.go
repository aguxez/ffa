package main

import (
	"fmt"

	"github.com/aguxez/ffa/bot"
	"github.com/aguxez/ffa/config"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
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

	db, err := database.Connect(cfg)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error connecting to database: %v", err.Error()))
	}

	// Initialize state manager
	sm := models.NewStateManager()

	// Start the bot
	if err := bot.Init(sm, db); err != nil {
		log.Err(err).Msg("Error initializing bot")
	}
}
