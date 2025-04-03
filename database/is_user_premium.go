package database

// IsUserPremium returns whether a user with the given ID is a premium user.
//
// The function takes a database connection and a userID as parameters.
// It queries the database for a user with the specified userID, and
// returns the user's IsPremium field, as well as an error if there was
// an issue with the query.
func (db *Database) IsUserPremium(userID uint) (bool, error) {
	var user User
	err := db.Conn.
		Where("id = ?", userID).
		First(&user).Error

	return user.IsPremium, err
}
