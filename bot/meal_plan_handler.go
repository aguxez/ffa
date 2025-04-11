package bot

import (
	"context"
	"errors"

	"github.com/aguxez/ffa/database"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// MealPlanHandler handles meal plan creation
type MealPlanHandler struct {
	deps *HandlerDependencies
	// This helps us determine which type of plan to generate when handling requests
	planType database.PlanType
}

// NewMealPlanHandler creates a new MealPlanHandler
func NewMealPlanHandler(deps *HandlerDependencies, planType database.PlanType) *MealPlanHandler {
	return &MealPlanHandler{
		deps:     deps,
		planType: planType,
	}
}

// Handle handles the meal plan creation
func (h *MealPlanHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Str("handler", "MealPlanHandler").
		Str("planType", string(h.planType)).
		Logger()

	logger.Debug().Msg("Handling meal plan creation")

	// Load user and associations
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user foods")
		ErrorResponse(ctx, bot, chatID, "Failed to retrieve your food preferences.")
		return
	}

	if h.planType == database.PlanTypeWeekdays {
		h.handleWeekdaysPlanGeneration(ctx, bot, update, user)
		return
	}

	if h.planType == database.PlanTypeWeekends {
		h.handleWeekendPlanGeneration(ctx, bot, update, user)
		return
	}
}

// handleWeekdaysPlanGeneration generates a meal plan for weekdays
func (h *MealPlanHandler) handleWeekdaysPlanGeneration(ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
	user *database.User,
) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Logger()

	logger.Debug().Msg("Handling weekdays plan generation")

	// Check if user has not reached the maximum number of meal plans
	numberOfPlansCreated, err := h.deps.DB.GetCountOfUserPlans(user.ID)
	if err != nil {
		logger.Err(err).Msg("Error getting count of user plans")
		ErrorResponse(ctx, bot, chatID, "Failed to retrieve user data.")
		return
	}
	if numberOfPlansCreated >= MAX_FREE_USERS_PLANS && !user.IsPremium {
		logger.Debug().Msg("User has reached the maximum number of meal plans")
		ErrorResponse(ctx, bot, chatID, "You have reached the maximum number of meal plans. You can upgrade to enjoy all features!")
		return
	}

	h.generatePlan(ctx, bot, update, user)
}

// handleWeekendPlanGeneration generates a meal plan for weekends
func (h *MealPlanHandler) handleWeekendPlanGeneration(ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
	user *database.User,
) {
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Logger()

	logger.Debug().Msg("Handling weekend plan generation")

	h.generatePlan(ctx, bot, update, user)
}

// generatePlan generates the actual meal plan
func (h *MealPlanHandler) generatePlan(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, user *database.User) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Logger()

	// Remove the markup button
	if err := h.removeMarkupOrDeleteMessage(ctx, bot, update); err != nil {
		logger.Err(err).Msg("Error removing markup")
		return
	}

	// Get last accepted meal plan
	lastAcceptedMealPlan, err := h.deps.DB.GetUserLastAcceptedMealPlan(userID, database.PlanTypeWeekdays)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Err(err).Msg("Error getting last accepted meal plan")
		ErrorResponse(ctx, bot, chatID, "There was an error getting your last accepted meal plan. This has been informed.")
		return
	}

	// Generate meal plan
	plan, err := h.deps.NutritionAgent.GenerateMealPlan(ctx, user, lastAcceptedMealPlan, h.planType)
	if err != nil {
		logger.Err(err).Msg("Error generating meal plan")
		ErrorResponse(ctx, bot, chatID, "There was an error generating the meal plan. This has been informed.")
		return
	}

	logger.Debug().Msg("Created meal plan")

	// Generate PDF
	doc, err := GenerateDocFromMealPlan(plan)
	if err != nil {
		logger.Err(err).Msg("Error generating document")
		ErrorResponse(ctx, bot, chatID, "There was an error generating your document. This has been informed.")
		return
	}

	logger.Debug().Msg("Generated document from meal plan")

	// Send document
	msg, err := bot.SendDocument(ctx, &tgbot.SendDocumentParams{
		ChatID:  chatID,
		Caption: "Here is our recommended meal plan based on the schedule and information you have shared with us!",
		Document: &tgmodels.InputFileUpload{
			Filename: "personalized_meal_plan.pdf",
			Data:     &doc,
		},
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Accept ✅",
						CallbackData: string(CallBackDataAcceptPlan),
					},
					{
						Text:         "Reject ❌",
						CallbackData: string(CallbackDataRejectPlan),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending document")
		ErrorResponse(ctx, bot, chatID, "There was an error sending your document. This has been informed.")
		return
	}

	// Save the meal plan to reference later
	var userMealPlanInput []database.UserMealPlanFoodInput
	for _, planFood := range plan.Plan {
		userMealPlanInput = append(userMealPlanInput, database.UserMealPlanFoodInput{
			Name:     planFood.Food,
			Category: string(planFood.FoodCategory),
		})
	}

	err = h.deps.DB.SaveUserMealPlan(userID, msg.ID, h.planType, userMealPlanInput)
	if err != nil {
		logger.Err(err).
			Int("messageID", msg.ID).
			Msg("Error saving user meal plan")
		return
	}
}

// removeMarkupOrDeleteMessage modifies or removes the message that triggered
// the callback query. If the message has no media attached, it updates the
// message text to inform the user about the ongoing plan generation process.
// If the message contains media, it deletes the message instead.
func (h *MealPlanHandler) removeMarkupOrDeleteMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) error {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	var fileID string
	if update.CallbackQuery.Message.Message.Document != nil {
		fileID = update.CallbackQuery.Message.Message.Document.FileID
	}

	// Message has no media attached to it. The message that triggered it only has text
	if fileID == "" {
		// Edit the previous message's text to inform about plan generation
		_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      "I'm generating a plan for you. This can take a minute.",
		})

		return err
	}

	// Remove the message as it has media attached to it
	_, err := bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: messageID,
	})

	return err
}
