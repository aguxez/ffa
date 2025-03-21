package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Foods        []UserFood
	PlanSchedule *UserPlanSchedule `gorm:"default:NULL"`

	Protein  *uint32 `gorm:"default:NULL"`
	Fat      *uint32 `gorm:"default:NULL"`
	Carbs    *uint32 `gorm:"default:NULL"`
	Calories *uint   `gorm:"default:NULL"`
}

type UserFood struct {
	gorm.Model

	UserID uint   `gorm:"uniqueIndex:user_foods_user_id_name_idx;constraint:OnDelete:CASCADE;not null"`
	Name   string `gorm:"uniqueIndex:user_foods_user_id_name_idx;not null"`
}

type UserPlanSchedule struct {
	gorm.Model

	UserID uint `gorm:"uniqueIndex;constraint:OnDelete:CASCADE;not null"`

	IntervalDays      uint      `gorm:"not null"`
	StartDate         time.Time `gorm:"not null"`
	LastExecutionDate *time.Time
}
