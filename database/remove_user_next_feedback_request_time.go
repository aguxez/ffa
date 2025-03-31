package database

// RemoveUserNextFeedbackRequestTime clears the next meal plan feedback request time for a user.
// We clear this value so we do not message the user again unless they approve a new meal plan.
//
// It sets the `next_meal_plan_feedback_request_time` field to nil for the user with the specified userID.
// This function returns an error if there is an issue during the update in the database.
func (db *Database) RemoveUserNextFeedbackRequestTime(userID uint) error {
	return db.Conn.Model(&User{}).
		Where("id = ?", userID).
		Update("next_meal_plan_feedback_request_time", nil).Error
}
