package database

type UserMealPlanFoodInput struct {
	Name     string
	Category string
}

// SaveUserMealPlan inserts a UserMealPlan into the database.
func (db *Database) SaveUserMealPlan(userID uint,
	messageID int,
	planType PlanType,
	foods []UserMealPlanFoodInput,
) error {
	mealPlan := UserMealPlan{
		UserID:    userID,
		MessageID: messageID,
		PlanType:  planType,
	}

	for _, food := range foods {
		mealPlan.Foods = append(mealPlan.Foods, UserMealPlanFood{
			Name:     food.Name,
			Category: food.Category,
		})
	}

	return db.Conn.Create(&mealPlan).Error
}
