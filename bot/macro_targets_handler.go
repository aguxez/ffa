package bot

import (
	"context"

	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MacroTargetsHandler handles the macro targets button
type MacroTargetsHandler struct {
	deps *HandlerDependencies
}

// Handle handles the macro targets button
func NewMacroTargetsHandler(deps *HandlerDependencies) *MacroTargetsHandler {
	return &MacroTargetsHandler{
		deps: deps,
	}
}

// Handle handles the macro targets button.
func (h *MacroTargetsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	logger := log.With().Uint("userID", userID).Str("handler", "MacroTargetsHandler").Logger()
	logger.Debug().Msg("Handling macro targets button")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		// Send generic error in default language (English) as we couldn't fetch user lang
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddMacroTargets)

	// Edit the callback text
	_, err = bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      i18n.GetText(i18n.KeyMacroTargetsPrompt, userLang),
	})
	if err != nil {
		logger.Err(err).Msg("Error sending macro targets message")
		return
	}
}

// MacroTargetsMessageHandler handles the macro targets message after the callback
type MacroTargetsMessageHandler struct {
	deps *HandlerDependencies
}

// NewMacroTargetsMessageHandler creates a new MacroTargetsMessageHandler
func NewMacroTargetsMessageHandler(deps *HandlerDependencies) *MacroTargetsMessageHandler {
	return &MacroTargetsMessageHandler{
		deps: deps,
	}
}

// Handle handles the macro targets message after the callback
func (h *MacroTargetsMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "MacroTargetsMessageHandler").Logger()
	logger.Debug().Msg("Handling macro targets message")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Update user's state to the beginning
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Parse the message received by the user
	macroTargets, err := h.deps.NaturalLanguageAgent.ParseMacroTargetsMessage(ctx, update.Message.Text)
	if err != nil {
		logger.Err(err).Msg("Error parsing macro targets message")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang)) // Consider a more specific key if needed
		return
	}

	// Update user's macro targets
	err = h.deps.DB.UpdateUserMacros(userID, macroTargets)
	if err != nil {
		logger.Err(err).Msg("Error updating user's macro targets")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.GetText(i18n.KeyMacroTargetsUpdated, userLang),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         i18n.GetText(i18n.KeyBackToOptionsButton, userLang),
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending update message")
		// No need to send another error message here, just log
		return
	}
}
