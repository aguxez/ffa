package database

import "time"

// GetUsersForFeedbackRequest retrieves a list of users that are due for a
// meal plan feedback request.
func (db *Database) GetUsersForFeedbackRequest(limit int) ([]User, error) {
	var users []User
	err := db.Conn.
		// Ensure we're only focusing on the latest approved meal plan since users can
		// have multiple approved plans within the same window.
		Joins("JOIN (SELECT user_id, MAX(created_at) as latest_approved_date FROM user_meal_plans WHERE is_accepted = true GROUP BY user_id) latest ON latest.user_id = users.id").
		Joins("JOIN user_meal_plans uml ON uml.user_id = users.id AND uml.created_at = latest.latest_approved_date").
		Where(`
		users.chat_id IS NOT NULL AND
		users.is_premium AND
		uml.is_accepted = true AND
		uml.created_at >= ? AND
		users.next_meal_plan_feedback_request_time IS NOT NULL AND
		users.next_meal_plan_feedback_request_time < ?`,
			// Only accepted plans that are no older than 3 days ago
			time.Now().AddDate(0, 0, -3),
			// Only plans where the request time has already passed
			time.Now()).
		Distinct("users.id").
		Limit(limit).
		Select("users.id, users.chat_id").
		Find(&users).Error
	return users, err
}
