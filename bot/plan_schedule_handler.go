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

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddPlanSchedule)

	h.removeMarkupAndEditMessage(ctx, bot, update)
}

// removeMarkupAndEditMessage removes the inline keyboard markup and edits the
// previous message's text
func (h *PlanScheduleHandler) removeMarkupAndEditMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Edit the message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "Please select an interval for your plan schedule.",
	})
	if err != nil {
		log.Err(err).Msg("Error sending plan schedule message")
		return
	}

	// Edit the message's markup
	_, err = bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:    chatID,
		MessageID: messageID,
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

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Update settings in database
	err := database.UpdateUserPlanSchedule(h.deps.DB, userID, h.intervalDays)
	if err != nil {
		log.Err(err).Msg("Error updating user plan schedule")
		return
	}

	h.removeMarkupAndEditMessage(ctx, bot, update)
}

func (h *PlanScheduleUpdateHandler) removeMarkupAndEditMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	// Edit the message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "Successfully updated your plan schedule!",
	})
	if err != nil {
		log.Err(err).Msg("Error sending plan schedule message")
		return
	}

	// Edit the message's markup
	_, err = bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:    chatID,
		MessageID: messageID,
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
		log.Err(err).Msg("Error sending plan schedule message")
	}
}
