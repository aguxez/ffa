package database

import "gorm.io/gorm"

// UpdateUserWeight updates the weight for a user in the database.
//
// It updates the existing user with the given ID by setting the weight
// to the given value. It returns an error if there is a problem updating
// the user.
func UpdateUserWeight(db *gorm.DB, userID uint, weight float64) error {
	return db.Save(&User{
		Model:  gorm.Model{ID: userID},
		Weight: &weight,
	},
	).Error
}
