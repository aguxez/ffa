package database

import (
	"time"

	"gorm.io/gorm"
)

// UpdateUserPlanSchedule updates the plan schedule for a user in the database.
// It takes a db, a user ID and the interval days to update the plan schedule with.
// It updates the existing user with the given ID by setting the plan schedule's
// interval days to the given interval days and sets the plan schedule's start date
// to the current time.
// The function returns an error if there is a problem updating the user.
func UpdateUserPlanSchedule(db *gorm.DB, userID uint, intervalDays uint) error {
	return db.Model(&User{}).
		Where("id = ?", userID).
		Updates(User{PlanSchedule: &UserPlanSchedule{IntervalDays: intervalDays, StartDate: time.Now()}}).Error
}
