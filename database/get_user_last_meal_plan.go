package database

// GetUserLastMealPlan retrieves the last accepted or rejected meal plan
//
// The function takes a user ID as parameter and returns a pointer to a UserMealPlan
// and an error.
func (db *Database) GetUserLastMealPlan(userID uint, planType PlanType, acceptedPlan bool) (*UserMealPlan, error) {
	var mealPlan UserMealPlan
	var orderColumn string

	if acceptedPlan {
		orderColumn = "accepted_at"
	} else {
		orderColumn = "rejected_at"
	}

	err := db.Conn.
		Where("user_id = ? AND accepted_at IS NOT NULL AND plan_type = ?", userID, planType).
		Order(orderColumn + " DESC").
		Preload("Foods").
		First(&mealPlan).Error
	if err != nil {
		return nil, err
	}

	return &mealPlan, err
}
