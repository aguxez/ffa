package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// PlanAcceptanceHandler handles the back to options button
type PlanAcceptanceHandler struct {
	deps           *HandlerDependencies
	isPlanAccepted bool
}

// NewPlanAcceptanceHandler creates a new PlanAcceptanceHandler
func NewPlanAcceptanceHandler(deps *HandlerDependencies, isPlanAccepted bool) *PlanAcceptanceHandler {
	return &PlanAcceptanceHandler{
		deps:           deps,
		isPlanAccepted: isPlanAccepted,
	}
}

// Handle handles the event whenever a user accepts or rejects a plan
func (h *PlanAcceptanceHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().
		Str("handler", "PlanAcceptanceHandler").
		Bool("isPlanAccepted", h.isPlanAccepted).
		Msg("Handling plan acceptance")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// If the user approved the plan, we want to save it in the database for further
	// feedback down the line
	if h.isPlanAccepted {
		if err := h.deps.DB.AcceptUserMealPlan(userID, messageID); err != nil {
			log.Err(err).Msg("Error accepting plan")
			return
		}

		// Remove markup from previous message
		_, err := bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
			ChatID:    chatID,
			MessageID: messageID,
			ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
			},
		})
		if err != nil {
			log.Err(err).Msg("Error removing markup from previous message")
			return
		}

		_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "Your plan has been saved and will be used for further feedback and check-ins.",
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
			log.Err(err).Msg("Error sending message")
			return
		}

		return
	}

	// Create a new plan for the user
	// TODO: Limit plan generation from this handler
	h.deps.HandlerRegistry.Get(string(CallbackDataCreateMealPlan)).Handle(ctx, bot, update)
}
