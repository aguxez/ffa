package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// PlanScheduleHandler handles adding plan schedule
type PlanScheduleHandler struct {
	deps *HandlerDependencies
}

// NewPlanScheduleHandler creates a new PlanScheduleHandler
func NewPlanScheduleHandler(deps *HandlerDependencies) *PlanScheduleHandler {
	return &PlanScheduleHandler{
		deps: deps,
	}
}

// Handle handles the plan schedule addition
func (h *PlanScheduleHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "PlanScheduleHandler").Logger()
	logger.Debug().Msg("Handling plan schedule addition")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddPlanSchedule)

	h.removeMarkupAndEditMessage(ctx, bot, update, userLang)
}

// removeMarkupAndEditMessage removes the inline keyboard markup and edits the
// previous message's text
func (h *PlanScheduleHandler) removeMarkupAndEditMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, userLang database.UserLanguage) { // Add language
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Edit the message's markup
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      i18n.GetText(i18n.KeyPlanScheduleSelectPrompt, userLang),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         i18n.GetText(i18n.KeyPlanScheduleFiveDaysButton, userLang),
						CallbackData: string(CallbackDataPlanScheduleWeekly),
					},
					{
						Text:         i18n.GetText(i18n.KeyPlanScheduleThreeDaysButton, userLang),
						CallbackData: string(CallbackDataPlanScheduleThreeDays),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending plan schedule selection message")
	}
}

// PlanScheduleUpdateHandler handles updating plan schedule
type PlanScheduleUpdateHandler struct {
	deps         *HandlerDependencies
	intervalDays uint
}

// NewPlanScheduleUpdateHandler creates a new PlanScheduleUpdateHandler
func NewPlanScheduleUpdateHandler(deps *HandlerDependencies, intervalDays uint) *PlanScheduleUpdateHandler {
	return &PlanScheduleUpdateHandler{
		deps:         deps,
		intervalDays: intervalDays,
	}
}

// Handle handles the plan schedule update
func (h *PlanScheduleUpdateHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "PlanScheduleUpdateHandler").Logger()
	logger.Debug().Msg("Handling plan schedule update")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update settings in database
	err = h.deps.DB.UpdateUserPlanSchedule(userID, h.intervalDays)
	if err != nil {
		logger.Err(err).Msg("Error updating user plan schedule")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, userLang))
		return
	}

	h.removeMarkupAndEditMessage(ctx, bot, update, userLang)
}

func (h *PlanScheduleUpdateHandler) removeMarkupAndEditMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, userLang database.UserLanguage) { // Add language
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Edit the message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      i18n.GetText(i18n.KeyPlanScheduleUpdated, userLang),
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
		log.Err(err).Msg("Error sending plan schedule update confirmation message")
		return
	}
}
