package database

import "gorm.io/gorm"

// GetUserFoods retrieves a list of foods for a given user ID from the database.
//
// The function takes a db and a user ID as parameters. It retrieves all the user foods
// that have the given user ID and returns them in a slice of UserFood structs.
//
// The function returns an error if there is a problem retrieving the user foods.
func GetUserFoods(db *gorm.DB, userID uint) ([]UserFood, error) {
	var userFoods []UserFood
	err := db.Where("user_id = ?", userID).Find(&userFoods).Error
	return userFoods, err
}
