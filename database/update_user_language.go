package database

// UpdateUserLanguage updates the language for a user in the database.
//
// It takes a user ID and a language as parameters and updates the existing user
// with the given ID by setting the language to the given value. It returns an
// error if there is a problem updating the user.
func (db *Database) UpdateUserLanguage(userID int64, language UserLanguage) error {
	return db.
		Conn.
		Model(&User{}).
		Where("id = ?", userID).
		Update("language", language).Error
}
