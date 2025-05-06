package database

// GetCountOfAcceptedPlans retrieves the count of accepted meal plans a user has generated.
//
// The function takes a user ID as parameter and returns the count of accepted meal plans
// the user has generated. It returns an error if there was an issue with the database call.
func (db *Database) GetCountOfAcceptedPlans(userID uint) (int64, error) {
	var count int64
	err := db.
		Conn.
		Model(&UserMealPlan{}).
		Where("id = ? AND accepted_at IS NOT NULL", userID).
		Count(&count).
		Error

	return count, err
}
