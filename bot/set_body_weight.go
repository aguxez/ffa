package bot

import (
	"context"
	"strconv"

	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// SetBodyWeightHandler handles the /start command
type SetBodyWeightHandler struct {
	deps *HandlerDependencies
}

// NewSetBodyWeightHandler creates a new SetBodyWeightHandler
func NewSetBodyWeightHandler(deps *HandlerDependencies) *SetBodyWeightHandler {
	return &SetBodyWeightHandler{
		deps: deps,
	}
}

// Handle handles the message stage for setting body weight
func (h *SetBodyWeightHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	logger := log.With().Uint("userID", userID).Str("handler", "SetBodyWeightHandler").Logger()
	logger.Debug().Msg("Handling set body weight")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Set user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageSetBodyWeight)

	// Edit the previous message's text
	_, err = bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      i18n.GetText(i18n.KeyBodyWeightPrompt, userLang),
	})
	if err != nil {
		// Log the error, but don't send another message as the original might still be visible
		logger.Err(err).Msg("Error sending set body weight prompt message")
		return
	}
}

// SetBodyWeightMessageHandler handles the
type SetBodyWeightMessageHandler struct {
	deps *HandlerDependencies
}

// NewSetBodyWeightMessageHandler creates a new SetBodyWeightMessageHandler
func NewSetBodyWeightMessageHandler(deps *HandlerDependencies) *SetBodyWeightMessageHandler {
	return &SetBodyWeightMessageHandler{
		deps: deps,
	}
}

// Handle handles the message containing the user's body weight
func (h *SetBodyWeightMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "SetBodyWeightMessageHandler").Logger()
	logger.Debug().Msg("Handling set body weight conversation stage")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	parsedWeight, err := strconv.ParseFloat(update.Message.Text, 64)
	if err != nil {
		// Send message informing about the error
		logger.Err(err).Str("input", update.Message.Text).Msg("Error parsing body weight message")
		_, sendErr := bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   i18n.GetText(i18n.KeyBodyWeightInvalidInput, userLang),
		})
		if sendErr != nil {
			logger.Err(sendErr).Msg("Error sending body weight invalid input message")
		}
		// Keep user in the same state to allow re-entry
		return
	}

	// Set user state after we have successfully parsed the weight
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update user's body weight
	err = h.deps.DB.UpdateUserWeight(userID, parsedWeight)
	if err != nil {
		logger.Err(err).Msg("Error updating user's body weight")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.GetText(i18n.KeyBodyWeightSaved, userLang),
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
		logger.Err(err).Msg("Error sending body weight saved confirmation message")
		return
	}
}
