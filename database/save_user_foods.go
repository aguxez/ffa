package database

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveUserFoods saves a list of user foods to the database.
// It takes a slice of food names and a user ID, and creates a slice of UserFood
// structs from that.
// We do not care about normalization in this instance since we receive this
// data from the user. If we are able to parse them correctly from natural lang
// then we can structure the data better.
//
// The function returns an error if there is a problem saving the foods.
func SaveUserFoods(db *gorm.DB, userID uint, foodNames []string) error {
	var userFoods []UserFood
	for _, food := range foodNames {
		userFoods = append(userFoods, UserFood{
			UserID: userID,
			Name:   strings.ToLower(food),
		})
	}

	// If there is a conflicting entry when inserting foods for a specific user
	// we can simply ignore it.
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&userFoods).Error
}
