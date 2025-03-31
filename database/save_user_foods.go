package database

import (
	"strings"

	"gorm.io/gorm/clause"
)

type UserFoodsPayload []string

// SaveUserFoods saves a list of user foods to the database.
// It takes a slice of food names and a user ID, and creates a slice of UserFood
// structs from that.
// We do not care about normalization in this instance since we receive this
// data from the user. If we are able to parse them correctly from natural lang
// then we can structure the data better.
//
// The function returns an error if there is a problem saving the foods.
func (db *Database) SaveUserFoods(userID uint, foodNames UserFoodsPayload) error {
	var userFoods []UserFood
	for _, food := range foodNames {
		userFoods = append(userFoods, UserFood{
			UserID: userID,
			Name:   strings.ToLower(food),
		})
	}

	// If there is a conflicting entry when inserting foods for a specific user
	// we can simply ignore it.
	return db.Conn.Clauses(clause.OnConflict{DoNothing: true}).Create(&userFoods).Error
}
