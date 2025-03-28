package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Foods        []UserFood
	PlanSchedule *UserPlanSchedule `gorm:"default:NULL"`

	Protein             *uint32 `gorm:"default:NULL;column:protein_grams"`
	Fat                 *uint32 `gorm:"default:NULL;column:fat_grams"`
	Carbs               *uint32 `gorm:"default:NULL;column:carbs_grams"`
	Calories            *uint   `gorm:"default:NULL"`
	DietaryRestrictions []UserDietaryRestriction
	Weight              *float64 `gorm:"default:NULL"`
	UserMealPlans       []UserMealPlan
}

type UserFood struct {
	gorm.Model

	UserID uint   `gorm:"uniqueIndex:user_foods_user_id_name_idx;constraint:OnDelete:CASCADE;not null"`
	Name   string `gorm:"uniqueIndex:user_foods_user_id_name_idx;not null"`
}

type UserPlanSchedule struct {
	gorm.Model

	UserID uint `gorm:"uniqueIndex;constraint:OnDelete:CASCADE;not null"`

	IntervalDays      uint      `gorm:"not null;default:5"`
	StartDate         time.Time `gorm:"not null"`
	LastExecutionDate *time.Time
}

type UserDietaryRestriction struct {
	gorm.Model

	UserID uint   `gorm:"uniqueIndex:user_dietary_restrictions_user_id_name_idx;constraint:OnDelete:CASCADE;not null"`
	Name   string `gorm:"uniqueIndex:user_dietary_restrictions_user_id_name_idx;not null"`
}

// UserMealPlan represents a meal plan of a user (usually accepted plans)
type UserMealPlan struct {
	gorm.Model

	UserID     uint `gorm:"constraint:OnDelete:CASCADE;not null"`
	IsAccepted bool `gorm:"not null;default:false"`
	// Represents the message ID that has the meal plan attached
	MessageID int                `gorm:"not null"`
	Foods     []UserMealPlanFood `gorm:"foreignKey:UserMealPlanID"`
}

type UserMealPlanFood struct {
	gorm.Model

	UserMealPlanID uint   `gorm:"uniqueIndex:user_meal_plan_foods_user_meal_plan_id_name_idx;constraint:OnDelete:CASCADE;not null"`
	Name           string `gorm:"uniqueIndex:user_meal_plan_foods_user_meal_plan_id_name_idx;not null"`
	Category       string `gorm:"not null"`
}
