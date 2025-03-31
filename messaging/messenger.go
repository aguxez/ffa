package messaging

import "github.com/aguxez/ffa/database"

type Options struct {
	WithInlineKeyboardFeedback bool
}

type Option func(*Options)

// Messenger is an interface for sending messages (bot, pubsub, etc.)
type Messenger interface {
	SendMessage(user database.User, message string, options ...Option) error
}

// WithInlineKeyboardFeedback is an Option that allows to send a message with
// an inline keyboard that allows the user to give feedback about the message.
func WithInlineKeyboardFeedback(value bool) Option {
	return func(o *Options) {
		o.WithInlineKeyboardFeedback = value
	}
}
