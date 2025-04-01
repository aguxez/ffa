package database

// GetUserFoods retrieves a list of food names for a given user from the database.
//
// The function takes a database connection and a user ID as parameters.
// It queries the database for foods associated with the specified user ID
// and returns a slice of food names. It also returns an error if there is
// an issue during the query.
func (db *Database) GetUserFoods(userID uint) ([]string, error) {
	var foods []string
	err := db.Conn.
		Where("user_id = ?", userID).
		Find(&foods).Error
	return foods, err
}
