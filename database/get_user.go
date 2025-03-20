package database

import "gorm.io/gorm"

// GetUser retrieves a User from the database with the given ID.
//
// The function returns nil if the user does not exist and an error if there
// was a problem retrieving the user.
func GetUser(db *gorm.DB, userID uint) (*User, error) {
	var user User
	err := db.First(&user, userID).Error
	return &user, err
}
