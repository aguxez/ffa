package messaging

import "github.com/aguxez/ffa/database"

// Messenger is an interface for sending messages (bot, pubsub, etc.)
type Messenger interface {
	SendMessage(user database.User, message string) error
}
