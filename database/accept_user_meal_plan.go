package database

import "gorm.io/gorm"

// AcceptUserMealPlan sets the is_accepted field of a UserMealPlan to true in the database.
//
// It takes a pointer to a gorm DB, a userID, and a messageID as parameters. It
// returns an error if there is a problem updating the UserMealPlan.
func AcceptUserMealPlan(db *gorm.DB, userID uint, messageID int) error {
	return db.Model(&UserMealPlan{}).
		Where("user_id = ?", userID).
		Where("message_id = ?", messageID).
		Update("is_accepted", true).Error
}
