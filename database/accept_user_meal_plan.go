package database

import (
	"time"

	"github.com/aguxez/ffa/config"
	"gorm.io/gorm"
)

// AcceptUserMealPlan accepts a user meal plan and updates the user's next
// meal plan feedback request time two days from now.
//
// It takes a user ID and a message ID as parameters and returns an error if there was
// an issue with the database call.
func (db *Database) AcceptUserMealPlan(userID uint, messageID int) error {
	// Add two days to current time
	nextTimeForFeedbackRequest := getTimeToAskForFeedback(db.Config)

	return db.Conn.Transaction(func(tx *gorm.DB) error {
		// Update the user meal plan
		err := tx.Model(&UserMealPlan{}).
			Where("user_id = ?", userID).
			Where("message_id = ?", messageID).
			Update("is_accepted", true).Error
		if err != nil {
			return err
		}

		// Update the next time we have to ask the user for feedback on this plan
		return tx.Model(&User{}).
			Where("id = ?", userID).
			Update("next_meal_plan_feedback_request_time", nextTimeForFeedbackRequest).Error
	})
}

// getTimeToAskForFeedback returns the time when we should ask a user
// for feedback on their meal plan. It takes a AppConfig object as a
// parameter and returns a time.Time object.
func getTimeToAskForFeedback(cfg config.AppConfig) time.Time {
	if cfg.Environment == "production" {
		// Two days from now
		return time.Now().Add(time.Hour * 24 * 2)
	}

	// Closer for development
	return time.Now().Add(time.Minute * 2)
}
