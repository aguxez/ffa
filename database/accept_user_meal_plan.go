package database

// AcceptUserMealPlan sets the is_accepted field of a UserMealPlan to true in the database.
//
// It takes a pointer to a gorm DB, a userID, and a messageID as parameters. It
// returns an error if there is a problem updating the UserMealPlan.
func (db *Database) AcceptUserMealPlan(userID uint, messageID int) error {
	return db.Conn.Model(&UserMealPlan{}).
		Where("user_id = ?", userID).
		Where("message_id = ?", messageID).
		Update("is_accepted", true).Error
}
