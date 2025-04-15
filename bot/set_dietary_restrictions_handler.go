package bot

import (
	"context"

	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

type SetDietaryRestrictionsHandler struct {
	deps *HandlerDependencies
}

// NewSetDietaryRestrictionsHandler creates a new SetDietaryRestrictionsHandler
func NewSetDietaryRestrictionsHandler(deps *HandlerDependencies) *SetDietaryRestrictionsHandler {
	return &SetDietaryRestrictionsHandler{
		deps: deps,
	}
}

// Handle handles the message stage for setting dietary restrictions
func (h *SetDietaryRestrictionsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	logger := log.With().Uint("userID", userID).Str("handler", "SetDietaryRestrictionsHandler").Logger()
	logger.Debug().Msg("Handling set dietary restrictions")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Set user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageSetDietaryRestrictions)

	// Edit the previous message's text
	_, err = bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      i18n.GetText(i18n.KeyDietaryRestrictionsPrompt, userLang),
	})
	if err != nil {
		logger.Err(err).Msg("Error sending set dietary restrictions prompt message")
		return
	}
}

type SetDietaryRestrictionsMessageHandler struct {
	deps *HandlerDependencies
}

// NewSetDietaryRestrictionsMessageHandler creates a new SetDietaryRestrictionsMessageHandler
func NewSetDietaryRestrictionsMessageHandler(deps *HandlerDependencies) *SetDietaryRestrictionsMessageHandler {
	return &SetDietaryRestrictionsMessageHandler{
		deps: deps,
	}
}

// Handle handles the message containing the user's dietary restrictions
func (h *SetDietaryRestrictionsMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	messageTxt := update.Message.Text

	logger := log.With().Uint("userID", userID).Str("handler", "SetDietaryRestrictionsMessageHandler").Logger()
	logger.Debug().Msg("Handling set dietary restrictions message")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Set user state first, before potential errors in processing/saving
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Process the message with the LLM
	userDietaryRestrictions, err := h.deps.NaturalLanguageAgent.ParseDietaryRestrictionsMessage(ctx, messageTxt)
	if err != nil {
		logger.Err(err).Msg("Error parsing dietary restrictions message")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang)) // Inform user
		return
	}

	// Update user's dietary restrictions
	err = h.deps.DB.SaveUserDietaryRestrictions(userID, userDietaryRestrictions)
	if err != nil {
		logger.Err(err).Msg("Error saving user's dietary restrictions")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.GetText(i18n.KeyDietaryRestrictionsUpdated, userLang),
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
		logger.Err(err).Msg("Error sending dietary restrictions update confirmation message")
		return
	}
}
