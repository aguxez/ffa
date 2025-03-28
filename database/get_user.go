package database

// GetUser retrieves a user from the database by userID.
//
// The function takes a pointer to a gorm.DB and a userID as parameters.
// It queries the database for a user with the specified userID and preloads
// some relevant associations. It returns a pointer to the
// User and an error if there was an issue during the query.

func (db *Database) GetUser(userID uint) (*User, error) {
	var user User
	err := db.Conn.
		Where("id = ?", userID).
		Preload("Foods").
		Preload("DietaryRestrictions").
		Preload("PlanSchedule").
		First(&user).Error

	return &user, err
}
