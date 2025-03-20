package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
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
	log.Debug().Str("handler", "PlanScheduleHandler").Msg("Handling plan schedule addition")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddPlanSchedule)

	// Delete message with the markup since we cannot modify the text of the message
	_, err := bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: messageID,
	})
	if err != nil {
		log.Err(err).Msg("Error deleting message")
		return
	}

	// Send schedule options
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "Please select an interval for your plan schedule.",
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Weekly",
						CallbackData: string(CallbackDataPlanScheduleWeekly),
					},
					{
						Text:         "Every 3 days",
						CallbackData: string(CallbackDataPlanScheduleThreeDays),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending plan schedule message")
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
	log.Debug().Str("handler", "PlanScheduleUpdateHandler").Msg("Handling plan schedule update")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update settings in database
	err := database.UpdateUserPlanSchedule(h.deps.DB, userID, h.intervalDays)
	if err != nil {
		log.Err(err).Msg("Error updating user plan schedule")
		return
	}

	// Delete message with the markup
	_, err = bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: messageID,
	})
	if err != nil {
		log.Err(err).Msg("Error deleting message")
		return
	}

	// Confirm to user
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "Successfully updated your plan schedule!",
	})
}
