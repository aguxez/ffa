package bot

import (
	"context"

	"github.com/aguxez/ffa/i18n"
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

	logger := log.With().Uint("userID", userID).Str("handler", "PlanAcceptanceHandler").Logger()

	// Get user language first
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update the plan acceptance in the database
	if err := h.deps.DB.UpdateUserPlanAcceptance(userID, messageID, h.isPlanAccepted); err != nil {
		logger.Err(err).Msg("Error accepting plan")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}

	// If the user approved the plan, we want to save it in the database for further
	// feedback down the line
	if h.isPlanAccepted {
		// Remove markup from previous message
		_, err = bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
			ChatID:    chatID,
			MessageID: messageID,
			ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
			},
		})
		if err != nil {
			// Log error but continue to send confirmation message
			logger.Err(err).Msg("Error removing markup from previous message after plan acceptance")
		}

		// Send confirmation message
		_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   i18n.GetText(i18n.KeyPlanSavedConfirmation, userLang), // Use i18n
			ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
					{
						{
							Text:         i18n.GetText(i18n.KeyBackToOptionsButton, userLang), // Use i18n
							CallbackData: string(CallbackDataBackToOptions),
						},
					},
				},
			},
		})
		if err != nil {
			logger.Err(err).Msg("Error sending plan acceptance confirmation message")
			return
		}

		// Plan accepted, flow ends here
		return
	}

	// Create a new plan for the user
	// TODO: Limit plan generation from this handler
	h.deps.HandlerRegistry.Get(string(CallbackDataCreateMealPlan)).Handle(ctx, bot, update)
}
