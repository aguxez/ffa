package database

import "gorm.io/gorm"

// UserMacrosPayload represents a message containing macro nutrient targets
type UserMacrosPayload struct {
	Protein  uint32 `json:"protein"`
	Fat      uint32 `json:"fat"`
	Carbs    uint32 `json:"carbs"`
	Calories uint   `json:"calories"`
}

// UpdateUserMacros updates the macro nutrient targets for a user in the database.
//
// It updates the existing user with the given ID by setting the macro nutrient targets
// to the given values. It returns an error if there is a problem updating the user.
func UpdateUserMacros(db *gorm.DB, userID uint, macros UserMacrosPayload) error {
	return db.Save(&User{
		Model:    gorm.Model{ID: userID},
		Protein:  &macros.Protein,
		Fat:      &macros.Fat,
		Carbs:    &macros.Carbs,
		Calories: &macros.Calories,
	},
	).Error
}
