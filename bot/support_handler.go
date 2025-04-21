package bot

import (
	"context"

	"github.com/aguxez/ffa/i18n"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// SupportHandler handles the /support command
type SupportHandler struct {
	deps *HandlerDependencies
}

// NewSupportHandler creates a new SupportHandler
func NewSupportHandler(deps *HandlerDependencies) *SupportHandler {
	return &SupportHandler{
		deps: deps,
	}
}

// Handle handles the /support command
func (h *SupportHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	logger := log.
		With().
		Int64("chatID", chatID).
		Int64("userID", userID).
		Str("handler", "SupportHandler").
		Logger()

	logger.Debug().Msg("Handling support command")

	user, err := h.deps.DB.GetUser(uint(userID))
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Send support message
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.GetText(i18n.KeySupportText, userLang),
	})

	if err != nil {
		logger.Err(err).Msg("Error sending support message")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}
}
