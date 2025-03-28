package bot

import (
	"context"

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

// Handle handles the message stage for the dietary restrictions
func (h *SetDietaryRestrictionsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.
		Debug().
		Str("handler", "SetDietaryRestrictionsHandler").
		Msg("Handling set dietary restrictions")

	// Set user state
	h.deps.StateManager.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageSetDietaryRestrictions)

	// Edit the previous message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "Tell me any food allergies, intolerances, or dietary preferences (vegan, vegetarian, gluten-free, etc.). Your meal plans will automatically exclude restricted ingredients.",
	})
	if err != nil {
		log.Err(err).Msg("Error sending set dietary restrictions message")
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

// Handle handles the set dietary restrictions command
func (h *SetDietaryRestrictionsMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "SetDietaryRestrictionsMessageHandler").Msg("Handling set dietary restrictions")

	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	messageTxt := update.Message.Text

	// Set user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Process the message with the LLM
	userDietaryRestrictions, err := h.deps.NaturalLanguageAgent.ParseDietaryRestrictionsMessage(ctx, messageTxt)
	if err != nil {
		log.Err(err).Msg("Error parsing dietary restrictions message")
		return
	}

	// Update user's dietary restrictions
	err = h.deps.DB.SaveUserDietaryRestrictions(userID, userDietaryRestrictions)
	if err != nil {
		log.Err(err).Msg("Error updating user's dietary restrictions")
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "Amazing. Your dietary restrictions have been updated!",
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "↼ Back to options",
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending dietary restrictions message")
		return
	}
}
