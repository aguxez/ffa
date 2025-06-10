package database

import (
	"time"

	"github.com/aguxez/ffa/config"
	"gorm.io/gorm"
)

// UpdateUserPlanAcceptance updates the acceptance status of a user's meal plan.
// The function deduces which column to updated based on whether the plan was accepted
// or rejected.
//
// It takes a user ID and a message ID as parameters and returns an error if there was
// an issue with the database call.
func (db *Database) UpdateUserPlanAcceptance(userID uint, messageID int, isMealPlanAccepted bool) error {
	// Add two days to current time
	nextTimeForFeedbackRequest := getTimeToAskForFeedback(db.Config)
	var columnToUpdate string

	if isMealPlanAccepted {
		columnToUpdate = "accepted_at"
	} else {
		columnToUpdate = "rejected_at"
	}

	return db.Conn.Transaction(func(tx *gorm.DB) error {
		// Update the user meal plan
		err := tx.Model(&UserMealPlan{}).
			Where("user_id = ?", userID).
			Where("message_id = ?", messageID).
			Update(columnToUpdate, time.Now()).Error
		if err != nil {
			return err
		}

		// Update the next time we have to ask the user for feedback on this plan
		if isMealPlanAccepted {
			return tx.Model(&User{}).
				Where("id = ?", userID).
				Update("next_meal_plan_feedback_request_time", nextTimeForFeedbackRequest).Error
		}

		return nil
	})
}

// getTimeToAskForFeedback calculates the future time when feedback should be requested
// from the user, based on the configured number of minutes specified in the
// application configuration.
func getTimeToAskForFeedback(cfg config.AppConfig) time.Time {
	return time.Now().Add(time.Minute * time.Duration(cfg.MealPlanFeedbackRequestMinutes))
}
