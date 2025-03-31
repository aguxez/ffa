package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MealType represents the type of meal for feedback
type MealType string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
)

// FeedbackResponse represents user's response to meal quality
type FeedbackResponse bool

const (
	FeedbackResponseGood FeedbackResponse = true
	FeedbackResponseBad  FeedbackResponse = false
)

// MealFeedbackConfig defines the configuration for each meal feedback step
type MealFeedbackConfig struct {
	MealType       MealType
	CurrentStage   models.ConversationStage
	NextStage      models.ConversationStage
	Question       string
	GoodCallbackID CallbackData
	BadCallbackID  CallbackData
}

// MealPlanFeedbackHandler handles meal plan feedback collection
type MealPlanFeedbackHandler struct {
	deps   *HandlerDependencies
	config map[models.ConversationStage]MealFeedbackConfig
}

// NewMealPlanFeedbackHandler creates a new MealPlanFeedbackHandler
func NewMealPlanFeedbackHandler(deps *HandlerDependencies) *MealPlanFeedbackHandler {
	// Define the flow configuration
	config := map[models.ConversationStage]MealFeedbackConfig{
		models.ConversationStageMealPlanFeedbackBreakfast: {
			MealType:       MealTypeBreakfast,
			CurrentStage:   models.ConversationStageMealPlanFeedbackBreakfast,
			NextStage:      models.ConversationStageMealPlanFeedbackLunch,
			Question:       "How do you feel about the breakfast? (1/3)",
			GoodCallbackID: CallbackDataMealPlanFeedbackBreakfastGood,
			BadCallbackID:  CallbackDataMealPlanFeedbackBreakfastBad,
		},
		models.ConversationStageMealPlanFeedbackLunch: {
			MealType:       MealTypeLunch,
			CurrentStage:   models.ConversationStageMealPlanFeedbackLunch,
			NextStage:      models.ConversationStageMealPlanFeedbackDinner,
			Question:       "How do you feel about the lunch? (2/3)",
			GoodCallbackID: CallbackDataMealPlanFeedbackLunchGood,
			BadCallbackID:  CallbackDataMealPlanFeedbackLunchBad,
		},
		models.ConversationStageMealPlanFeedbackDinner: {
			MealType:       MealTypeDinner,
			CurrentStage:   models.ConversationStageMealPlanFeedbackDinner,
			NextStage:      models.ConversationStageNull, // End of flow
			Question:       "How do you feel about the dinner? (3/3)",
			GoodCallbackID: CallbackDataMealPlanFeedbackDinnerGood,
			BadCallbackID:  CallbackDataMealPlanFeedbackDinnerBad,
		},
	}

	return &MealPlanFeedbackHandler{
		deps:   deps,
		config: config,
	}
}

// Handle processes meal plan feedback based on user state
func (h *MealPlanFeedbackHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	logger := log.With().Str("handler", "MealPlanFeedbackHandler").Logger()
	logger.Debug().Msg("Handling meal plan feedback")

	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID
	userState := h.deps.StateManager.GetUserState(userID)

	// Handle start of feedback process
	if update.CallbackQuery.Data == string(CallbackDataMealPlanFeedbackStart) {
		h.startFeedbackProcess(ctx, bot, userID, chatID, messageID)
		return
	}

	// If no active feedback stage, start with breakfast
	if userState.Stage == models.ConversationStageNull {
		h.startFeedbackProcess(ctx, bot, userID, chatID, messageID)
		return
	}

	// Process feedback for current meal
	h.processMealFeedback(ctx, bot, update, userID, chatID, messageID, userState.Stage)
}

// startFeedbackProcess initiates the feedback collection process
func (h *MealPlanFeedbackHandler) startFeedbackProcess(
	ctx context.Context,
	bot *tgbot.Bot,
	userID uint,
	chatID int64,
	messageID int,
) {
	log.Debug().Msg("Starting feedback process with breakfast")

	// Set user state to breakfast feedback stage
	h.deps.StateManager.SetUserState(userID, models.ConversationStageMealPlanFeedbackBreakfast)

	// Send breakfast feedback question
	h.sendFeedbackQuestion(ctx, bot, chatID, messageID, h.config[models.ConversationStageMealPlanFeedbackBreakfast])
}

// processMealFeedback handles feedback for a specific meal type
func (h *MealPlanFeedbackHandler) processMealFeedback(
	ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
	userID uint,
	chatID int64,
	messageID int,
	currentStage models.ConversationStage,
) {
	logger := log.With().
		Str("handler", "MealPlanFeedbackHandler").
		Int("userID", int(userID)).
		Str("stage", string(currentStage)).
		Logger()

	logger.Debug().Msg("Processing meal feedback")

	// Get config for current stage
	config, exists := h.config[currentStage]
	if !exists {
		logger.Error().Msg("Invalid conversation stage")
		return
	}

	// Determine if user liked the meal
	callbackData := update.CallbackQuery.Data
	isMealLiked := callbackData == string(config.GoodCallbackID)

	// Update database with user's response
	err := h.saveFeedbackToDatabase(userID, config.MealType, isMealLiked)
	if err != nil {
		logger.Err(err).Msg("Error updating feedback in database")
		return
	}

	// Move to next stage
	h.deps.StateManager.SetUserState(userID, config.NextStage)

	// If this was the last meal, send completion message
	if config.NextStage == models.ConversationStageNull {
		h.sendCompletionMessage(ctx, bot, chatID, messageID)
		return
	}

	// Send question for next meal
	h.sendFeedbackQuestion(ctx, bot, chatID, messageID, h.config[config.NextStage])
}

// saveFeedbackToDatabase stores user's meal feedback
func (h *MealPlanFeedbackHandler) saveFeedbackToDatabase(userID uint, mealType MealType, liked bool) error {
	input := database.MealPlanFeedbackInput{}

	// Set appropriate field based on meal type
	switch mealType {
	case MealTypeBreakfast:
		input.IsBreakfastLiked = &liked
	case MealTypeLunch:
		input.IsLunchLiked = &liked
	case MealTypeDinner:
		input.IsDinnerLiked = &liked
	}

	return h.deps.DB.SetMealPlanFeedback(userID, input)
}

// sendFeedbackQuestion sends a meal feedback prompt to the user
func (h *MealPlanFeedbackHandler) sendFeedbackQuestion(
	ctx context.Context,
	bot *tgbot.Bot,
	chatID int64,
	messageID int,
	config MealFeedbackConfig,
) {
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      config.Question,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Enjoyed❤️",
						CallbackData: string(config.GoodCallbackID),
					},
					{
						Text:         "Disliked☹️",
						CallbackData: string(config.BadCallbackID),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Int64("chatID", chatID).Msg("Error sending feedback question")
	}
}

// sendCompletionMessage sends the final message after collecting all feedback
func (h *MealPlanFeedbackHandler) sendCompletionMessage(ctx context.Context,
	bot *tgbot.Bot,
	chatID int64,
	messageID int,
) {
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "Thanks for your feedback. Your recommendations will be updated for your next meal plan.",
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
		},
	})
	if err != nil {
		log.Err(err).Int64("chatID", chatID).Msg("Error sending completion message")
	}
}
