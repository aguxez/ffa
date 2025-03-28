package bot

import (
	"context"
	"strconv"

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

// Handle handles the message stage for the dietary restrictions
func (h *SetBodyWeightHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.
		Debug().
		Str("handler", "SetBodyWeightHandler").
		Msg("Handling set body weight")

	// Set user state
	h.deps.StateManager.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageSetBodyWeight)

	// Edit the previous message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "Tell me your body weight in kilograms.",
	})
	if err != nil {
		log.Err(err).Msg("Error sending set dietary restrictions message")
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

// Handle handles the set dietary restrictions command
func (h *SetBodyWeightMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "SetBodyWeightMessageHandler").Msg("Handling set body weight conversation stage")

	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	parsedWeight, err := strconv.ParseFloat(update.Message.Text, 64)
	if err != nil {
		// Send message informing about the error
		log.Err(err).Msg("Error parsing dietary restrictions message")

		_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "Please enter a valid number for your body weight.",
		})
		if err != nil {
			log.Err(err).Msg("Error sending dietary restrictions error message")
		}

		return
	}

	// Set user state after we have parsed the weight
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update user's body weight
	err = h.deps.DB.UpdateUserWeight(userID, parsedWeight)
	if err != nil {
		log.Err(err).Msg("Error updating user's dietary restrictions")
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "Your body weight has been saved!",
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
