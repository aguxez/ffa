package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// StartCommandHandler handles the /start command
type StartCommandHandler struct {
	deps *HandlerDependencies
}

// NewStartCommandHandler creates a new StartCommandHandler
func NewStartCommandHandler(deps *HandlerDependencies) *StartCommandHandler {
	return &StartCommandHandler{
		deps: deps,
	}
}

// Handle handles the /start command
func (h *StartCommandHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "StartCommandHandler").Msg("Handling start command")

	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	// Initialize user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageStart)

	// Ensure user exists in database
	_, err := database.GetOrCreateUser(h.deps.DB, userID)
	if err != nil {
		log.Err(err).Msg("Error finding or creating user")
		return
	}

	// Send welcome message with options
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "I can help you create meal plans tailored for your dietary preferences",
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Create Meal Plan",
						CallbackData: string(CallbackDataCreateMealPlan),
					},
				},
				{
					{
						Text:         "Add dietary preference",
						CallbackData: string(CallbackDataAddDietaryPreference),
					},
				},
				{
					{
						Text:         "Add food",
						CallbackData: string(CallbackDataAddUserFood),
					},
					{
						Text:         "Add schedule",
						CallbackData: string(CallbackDataAddPlanSchedule),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending welcome message")
	}
}
