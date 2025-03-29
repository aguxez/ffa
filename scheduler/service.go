package scheduler

import (
	"time"

	"github.com/aguxez/ffa/database"
	"github.com/rs/zerolog/log"
)

type SchedulerService struct {
	db *database.Database
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
func NewSchedulerService(db *database.Database) *SchedulerService {
	return &SchedulerService{db: db}
}

func (s *SchedulerService) Run() error {
	log.Debug().Time("currentTime", time.Now()).Msg("Running scheduler service for feedback request")

	users, err := s.db.GetUsersForFeedbackRequest()
	if err != nil {
		return err
	}

	for _, user := range users {
		// Log the user ID
		log.Info().Uint("userID", user.ID).Msg("Sending feedback request")
	}

	return nil
}
