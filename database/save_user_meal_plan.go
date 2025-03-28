package database

type UserMealPlanFoodInput struct {
	Name     string
	Category string
}

// SaveUserMealPlan inserts a UserMealPlan into the database.
// It takes a UserMealPlan and a GORM database object as parameters.
// It returns an error if there was an issue with the database call.
func (db *Database) SaveUserMealPlan(userID uint, messageID int, foods []UserMealPlanFoodInput) error {
	mealPlan := UserMealPlan{
		UserID:    userID,
		MessageID: messageID,
	}

	for _, food := range foods {
		mealPlan.Foods = append(mealPlan.Foods, UserMealPlanFood{
			Name:     food.Name,
			Category: food.Category,
		})
	}

	return db.Conn.Create(&mealPlan).Error
}
