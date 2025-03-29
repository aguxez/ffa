package scheduler

import (
	"time"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/messaging"
	"github.com/rs/zerolog/log"
)

type SchedulerService struct {
	db        *database.Database
	messenger messaging.Messenger
}

// NewSchedulerService creates a new SchedulerService
//
// Args:
//
//	db: The database to use for the service
//
// Returns:
//
//	A pointer to a new SchedulerService
func NewSchedulerService(db *database.Database, messenger messaging.Messenger) *SchedulerService {
	return &SchedulerService{db: db, messenger: messenger}
}

// Run runs the scheduler service for sending meal plan feedback requests to users.
func (s *SchedulerService) Run() error {
	log.Debug().Time("currentTime", time.Now()).Msg("Running scheduler service for feedback request")

	users, err := s.db.GetUsersForFeedbackRequest()
	if err != nil {
		return err
	}

	for _, user := range users {
		// Log the user ID
		log.Info().Uint("userID", user.ID).Msg("Sending feedback request")

		// Send the feedback request
		err := s.messenger.SendMessage(user, "Testing feedback request")
		if err != nil {
			log.Err(err).Uint("userID", user.ID).Msg("Error sending feedback request")
		}
	}

	return nil
}
