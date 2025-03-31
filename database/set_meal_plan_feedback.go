package database

type MealPlanFeedbackInput struct {
	IsBreakfastLiked *bool
	IsLunchLiked     *bool
	IsDinnerLiked    *bool
}

// SetMealPlanFeedback updates the latest accepted meal plan of a user with the given feedback.
//
// The function takes a user ID and a MealPlanFeedbackInput struct as parameters.
// It returns an error if there was an issue with the database call.
func (db *Database) SetMealPlanFeedback(userID uint, input MealPlanFeedbackInput) error {
	// Return the latest accepted meal plan for the user
	var mealPlan UserMealPlan
	if err := db.Conn.Where("user_id = ? AND is_accepted = ?", userID, true).Order("created_at DESC").First(&mealPlan).Error; err != nil {
		return err
	}

	// Update the meal plan with the feedback
	return db.Conn.Model(&UserMealPlan{}).
		Where("id = ?", mealPlan.ID).
		Updates(input).Error
}
