package database

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserDietaryRestrictionsPayload represents a message containing macro nutrient targets
type UserDietaryRestrictionsPayload []string

// SaveUserDietaryRestrictions saves a list of dietary restrictions for a user to the database.
//
// It takes a pointer to a gorm DB, a userID, and a slice of strings representing the dietary restrictions.
// It returns an error if there is a problem saving the foods.
func SaveUserDietaryRestrictions(db *gorm.DB, userID uint, restrictions UserDietaryRestrictionsPayload) error {
	var payload []UserDietaryRestriction
	for _, restriction := range restrictions {
		payload = append(payload, UserDietaryRestriction{
			UserID: userID,
			Name:   strings.ToLower(restriction),
		})
	}

	// If there is a conflicting entry when inserting foods for a specific user
	// we can simply ignore it.
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&payload).Error
}
