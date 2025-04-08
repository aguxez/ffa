package database

// GetCountOfUserPlans retrieves the count of meal plans a user has generated.
//
// The function takes a user ID as parameter and returns the count of meal plans
// the user has generated. It returns an error if there was an issue with the
// database call.
func (db *Database) GetCountOfUserPlans(userID uint) (int64, error) {
	var count int64
	err := db.Conn.
		Model(&UserMealPlan{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}
