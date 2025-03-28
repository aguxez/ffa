package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// StartCommandHandler handles the /start command
type StartCommandHandler struct {
	deps *HandlerDependencies
}

// NewStartCommandHandler creates a new StartCommandHandler
func NewStartCommandHandler(deps *HandlerDependencies) *StartCommandHandler {
	return &StartCommandHandler{
		deps: deps,
	}
}

// Handle handles the /start command
func (h *StartCommandHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "StartCommandHandler").Msg("Handling start command")

	var userID uint
	var chatID int64

	// This handler can be called from either a callback query triggering it
	// or a normal init message so we check these two fields for the existence
	// of the data we want
	if update.Message != nil {
		userID = uint(update.Message.From.ID)
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		userID = uint(update.CallbackQuery.From.ID)
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	}

	// Initialize user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageStart)

	// Ensure user exists in database
	_, err := h.deps.DB.GetOrCreateUser(userID)
	if err != nil {
		log.Err(err).Msg("Error finding or creating user")
		return
	}

	// Send welcome message with options
	_, err = bot.SendMessage(ctx, InitialMenuMessageParams(chatID))
	if err != nil {
		log.Err(err).Msg("Error sending welcome message")
	}
}
