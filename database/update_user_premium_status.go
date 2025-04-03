package database

import "gorm.io/gorm"

// UpdateUserPremiumStatus updates the premium status of a user in the database.
//
// The function takes a database connection, a user ID, a boolean indicating whether the user is premium or not,
// and an optional payment ID. If the payment ID is not nil, it inserts a new UserPayment record with the given
// payment ID and user ID. Then, it updates the `is_premium` field of the user with the given ID to the given boolean
// value. The function returns an error if there is a problem with the database call.
func (db *Database) UpdateUserPremiumStatus(userID uint, isPremium bool, paymentID *string) error {
	return db.Conn.Transaction(func(tx *gorm.DB) error {
		if paymentID != nil {
			err := tx.Create(&UserPayment{PaymentID: *paymentID, UserID: userID}).Error
			if err != nil {
				return err
			}
		}

		return db.Conn.Model(&User{}).
			Where("id = ?", userID).Update("is_premium", isPremium).Error
	})
}
