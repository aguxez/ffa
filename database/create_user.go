package database

import "gorm.io/gorm"

// CreateUser creates a user in the database.
//
// If the user exists, it doesn't change anything.
//
// The function returns nil if the user was created successfully and an error if
// there was a problem creating the user.
func CreateUser(db *gorm.DB, userID uint) error {
	return db.Create(&User{Model: gorm.Model{ID: userID}}).Error
}
