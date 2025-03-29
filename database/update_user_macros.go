package database

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
func (db *Database) UpdateUserMacros(userID uint, macros UserMacrosPayload) error {
	return db.Conn.
		Model(&User{}).
		Where("id = ?", userID).
		Updates(User{
			Protein:  &macros.Protein,
			Fat:      &macros.Fat,
			Carbs:    &macros.Carbs,
			Calories: &macros.Calories,
		}).Error
}
