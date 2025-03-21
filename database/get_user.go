package database

import "gorm.io/gorm"

// GetUser retrieves a user from the database by userID.
//
// The function takes a pointer to a gorm.DB and a userID as parameters.
// It queries the database for a user with the specified userID and preloads
// the associated Foods. It returns a pointer to the User and an error if
// there was an issue during the query.

func GetUser(db *gorm.DB, userID uint) (*User, error) {
	var user User
	err := db.Where("id = ?", userID).Preload("Foods").First(&user).Error
	return &user, err
}
