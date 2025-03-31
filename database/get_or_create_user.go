package database

import "gorm.io/gorm"

// GetOrCreateUser fetches a user from the database given a userID, or creates one if it doesn't exist.
//
// The function takes a pointer to a gorm DB and a userID as parameters. It first tries to fetch a user
// given the userID. If the user doesn't exist, it creates one with the given ID. It then returns a pointer
// to the user and an error if there was one.
func (db *Database) GetOrCreateUser(userID uint, chatID int64) (*User, error) {
	var user User

	err := db.Conn.
		Where("id = ?", userID).
		Attrs(User{Model: gorm.Model{ID: userID}, ChatID: &chatID}).
		FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}

	// This is a field we must always have set. Set like this
	// to not break existing conversations.
	// TODO: Remove once everyone has a chatID and make chatID not null
	if user.ChatID == nil {
		user.ChatID = &chatID
	}

	err = db.Conn.Save(&user).Error

	return &user, err
}
