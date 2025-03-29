package database

// UpdateUserWeight updates the weight for a user in the database.
//
// It updates the existing user with the given ID by setting the weight
// to the given value. It returns an error if there is a problem updating
// the user.
func (db *Database) UpdateUserWeight(userID uint, weight float64) error {
	return db.Conn.
		Model(&User{}).
		Where("id = ?", userID).
		Update("weight", weight).Error
}
