package database

import "gorm.io/gorm"

type UserFood struct {
	gorm.Model

	UserID uint   `gorm:"uniqueIndex:user_foods_user_id_name_idx"`
	Name   string `gorm:"uniqueIndex:user_foods_user_id_name_idx"`
}
