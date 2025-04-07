package database

import "time"

// GetUsersForFeedbackRequest retrieves a list of users that are due for a
// meal plan feedback request.
func (db *Database) GetUsersForFeedbackRequest(limit int) ([]User, error) {
	var users []User

	// This subquery finds the latest accepted meal plan for each user
	subquery := db.Conn.Table("user_meal_plans").
		Select("user_id, MAX(accepted_at) as latest_accepted_at").
		Where("accepted_at IS NOT NULL").
		Group("user_id")

	err := db.Conn.
		Table("users").
		Joins("JOIN (?) AS latest_plans ON users.id = latest_plans.user_id", subquery).
		Joins("JOIN user_meal_plans AS uml ON uml.user_id = users.id AND uml.accepted_at = latest_plans.latest_accepted_at").
		Where(`
					users.chat_id IS NOT NULL AND
					users.is_premium AND
					uml.created_at >= ? AND
					users.next_meal_plan_feedback_request_time IS NOT NULL AND
					users.next_meal_plan_feedback_request_time < ?`,
			// Only accepted plans that are no older than 3 days ago
			time.Now().AddDate(0, 0, -3),
			// Only plans where the request time has already passed
			time.Now()).
		Limit(limit).
		Select("users.id, users.chat_id").
		Find(&users).Error

	return users, err
}
