package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// BackToOptionsHandler handles the back to options button
type BackToOptionsHandler struct {
	deps                        *HandlerDependencies
	shouldDeletePreviousMessage bool
}

// NewBackToOptionsHandler creates a new BackToOptionsHandler
func NewBackToOptionsHandler(deps *HandlerDependencies, shouldDeletePreviousMessage bool) *BackToOptionsHandler {
	return &BackToOptionsHandler{
		deps:                        deps,
		shouldDeletePreviousMessage: shouldDeletePreviousMessage,
	}
}

func (h *BackToOptionsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Str("handler", "BackToOptionsHandler").
		Logger()

	logger.Debug().Msg("Handling back to options")

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageStart)

	// Delete the message that triggered this callback if necessary
	if h.shouldDeletePreviousMessage {
		_, err := bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: update.CallbackQuery.Message.Message.ID,
		})
		if err != nil {
			logger.Err(err).Msg("Error deleting message")
			return
		}
	}

	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error finding user")
		return
	}

	numberOfCreatedPlans, err := h.deps.DB.GetCountOfUserPlans(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting number of created plans")
		return
	}

	// Send the initial menu message
	_, err = bot.SendMessage(ctx, InitialMenuMessageParams(chatID, user, numberOfCreatedPlans))
	if err != nil {
		logger.Err(err).Msg("Error sending initial menu message")
	}
}
