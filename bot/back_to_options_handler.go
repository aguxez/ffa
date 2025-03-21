package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MealPlanHandler handles meal plan creation
type BackToOptionsHandler struct {
	deps *HandlerDependencies
}

// NewBackToOptionsHandler creates a new BackToOptionsHandler
func NewBackToOptionsHandler(deps *HandlerDependencies) *BackToOptionsHandler {
	return &BackToOptionsHandler{
		deps: deps,
	}
}

func (h *BackToOptionsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "BackToOptionsHandler").Msg("Handling back to options")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageStart)

	// Delete the message that triggered this callback
	_, err := bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		log.Err(err).Msg("Error deleting message")
		return
	}

	// Send the initial menu message
	_, err = bot.SendMessage(ctx, InitialMenuMessageParams(chatID))
	if err != nil {
		log.Err(err).Msg("Error sending initial menu message")
	}
}
