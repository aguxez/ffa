package database

import "gorm.io/gorm"

type UserMealPlanFoodInput struct {
	Name     string
	Category string
}

// SaveUserMealPlan inserts a UserMealPlan into the database.
// It takes a UserMealPlan and a GORM database object as parameters.
// It returns an error if there was an issue with the database call.
func SaveUserMealPlan(db *gorm.DB, userID uint, messageID int, foods []UserMealPlanFoodInput) error {
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

	return db.Create(&mealPlan).Error
}
