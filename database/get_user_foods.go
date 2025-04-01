package database

// GetUserFoods retrieves a paginated list of food names for a given user.
//
// It fetches `limit + 1` items starting from the `offset` to allow checking
// if there's a next page without a separate count query.
func (db *Database) GetUserFoods(userID uint, limit, offset int) ([]string, error) {
	var foods []string
	// Fetch limit + 1 to check if there are more items for the next page
	err := db.Conn.
		Model(&UserFood{}).
		Where("user_id = ?", userID).
		Order("created_at asc").
		Limit(limit+1).
		Offset(offset).
		// Use Pluck for efficiency if only names are needed
		Pluck("name", &foods).Error
	return foods, err
}
