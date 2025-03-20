package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Foods        []UserFood
	PlanSchedule UserPlanSchedule
}

type UserFood struct {
	gorm.Model

	UserID uint   `gorm:"uniqueIndex:user_foods_user_id_name_idx,constraint:OnDelete:CASCADE"`
	Name   string `gorm:"uniqueIndex:user_foods_user_id_name_idx"`
}

type UserPlanSchedule struct {
	gorm.Model

	UserID uint `gorm:"uniqueIndex,constraint:OnDelete:CASCADE"`

	IntervalDays      uint
	StartDate         time.Time
	LastExecutionDate time.Time
}
