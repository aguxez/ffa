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
			Update("accepted_at", time.Now()).Error
		if err != nil {
			return err
		}

		// Update the next time we have to ask the user for feedback on this plan
		return tx.Model(&User{}).
			Where("id = ?", userID).
			Update("next_meal_plan_feedback_request_time", nextTimeForFeedbackRequest).Error
	})
}

// getTimeToAskForFeedback calculates the future time when feedback should be requested
// from the user, based on the configured number of minutes specified in the
// application configuration.
func getTimeToAskForFeedback(cfg config.AppConfig) time.Time {
	return time.Now().Add(time.Minute * time.Duration(cfg.MealPlanFeedbackRequestMinutes))
}
