package bot

import (
	"context"
	"strings"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// UserFoodHandler handles adding user food preferences
type UserFoodHandler struct {
	deps *HandlerDependencies
}

// NewUserFoodHandler creates a new UserFoodHandler
func NewUserFoodHandler(deps *HandlerDependencies) *UserFoodHandler {
	return &UserFoodHandler{
		deps: deps,
	}
}

// Handle handles the user food preference addition
func (h *UserFoodHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "UserFoodHandler").Msg("Handling user food addition")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddUserFood)

	// Send instructions
	_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   `Sure, send me some of the foods you enjoy so I can reference them. These items will be added to the ones you already have saved instead of replacing them.`,
	})
	if err != nil {
		log.Err(err).Msg("Error sending message to add user foods")
	}
}

// UserFoodMessageHandler handles processing food items from user messages
type UserFoodMessageHandler struct {
	deps *HandlerDependencies
}

// NewUserFoodMessageHandler creates a new UserFoodMessageHandler
func NewUserFoodMessageHandler(deps *HandlerDependencies) *UserFoodMessageHandler {
	return &UserFoodMessageHandler{
		deps: deps,
	}
}

// Handle processes a message containing food items
func (h *UserFoodMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	log.Debug().Uint("userID", userID).Msg("Processing food message")

	// Parse food items from message
	foodsFromMessage := strings.Split(update.Message.Text, "\n")

	// Save to database
	err := database.SaveUserFoods(h.deps.DB, userID, foodsFromMessage)
	if err != nil {
		log.Err(err).Uint("userID", userID).Msg("Error saving user foods")
		ErrorResponse(ctx, bot, chatID, "There was an error saving your food preferences. Please try again.")
		return
	}

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Confirm to user
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "Successfully updated your food preferences!",
		ChatID: chatID,
	})
}
