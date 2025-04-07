package database

// GetUserLastAcceptedMealPlan retrieves the latest accepted meal plan of a user.
//
// It fetches a single UserMealPlan row from the database with the highest accepted_at
// value (i.e. the latest accepted meal plan) for the given user ID. If no such row
// exists, it returns a nil pointer and a gorm error.
//
// The function takes a user ID as parameter and returns a pointer to a UserMealPlan
// and an error.
func (db *Database) GetUserLastAcceptedMealPlan(userID uint) (*UserMealPlan, error) {
	var mealPlan UserMealPlan

	err := db.Conn.
		Where("user_id = ? AND accepted_at IS NOT NULL", userID).
		Order("accepted_at DESC").
		Preload("Foods").
		First(&mealPlan).Error
	if err != nil {
		return nil, err
	}

	return &mealPlan, err
}
