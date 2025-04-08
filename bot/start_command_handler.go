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
	var userID uint
	var chatID int64

	// This handler can be called from either a callback query triggering it
	// or a normal init message so we check these two fields for the existence
	// of the data we want
	if update.Message != nil {
		userID = uint(update.Message.From.ID)
		chatID = update.Message.Chat.ID
	}

	if update.CallbackQuery != nil {
		userID = uint(update.CallbackQuery.From.ID)
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	}

	logger := log.
		With().
		Uint("userID", userID).
		Str("handler", "StartCommandHandler").
		Logger()

	logger.Debug().Msg("Handling start command")

	// Initialize user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageStart)

	// Ensure user exists in database
	user, err := h.deps.DB.GetOrCreateUser(userID, chatID)
	if err != nil {
		logger.Err(err).Msg("Error finding or creating user")
		return
	}

	numberOfCreatedPlans, err := h.deps.DB.GetCountOfUserPlans(user.ID)
	if err != nil {
		logger.Err(err).Msg("Error getting number of created plans")
	}

	// Send welcome message with options
	_, err = bot.SendMessage(ctx, InitialMenuMessageParams(chatID, user, numberOfCreatedPlans))
	if err != nil {
		logger.Err(err).Msg("Error sending welcome message")
	}
}
