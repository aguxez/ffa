package database

import "gorm.io/gorm"

// GetOrCreateUser fetches a user from the database given a userID, or creates one if it doesn't exist.
//
// The function takes a pointer to a gorm DB and a userID as parameters. It first tries to fetch a user
// given the userID. If the user doesn't exist, it creates one with the given ID. It then returns a pointer
// to the user and an error if there was one.
func (db *Database) GetOrCreateUser(userID uint) (*User, error) {
	var user User

	err := db.Conn.Where("id = ?", userID).Attrs(User{Model: gorm.Model{ID: userID}}).FirstOrCreate(&user).Error

	return &user, err
}
