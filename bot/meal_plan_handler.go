package bot

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
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

	// Load user and associations (needed for language and plan generation)
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		// Can't get user language, default to English for error
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorRetrieveUserData, "en"))
		return
	}
	userLang := user.Language

	if h.planType == database.PlanTypeWeekdays {
		h.handleWeekdaysPlanGeneration(ctx, bot, update, user, userLang)
		return
	}

	if h.planType == database.PlanTypeWeekends {
		h.handleWeekendPlanGeneration(ctx, bot, update, user, userLang)
		return
	}
}

// handleWeekdaysPlanGeneration generates a meal plan for weekdays
func (h *MealPlanHandler) handleWeekdaysPlanGeneration(ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
	user *database.User,
	userLang database.UserLanguage,
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
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorRetrieveUserData, userLang))
		return
	}
	if numberOfPlansCreated >= MAX_FREE_USERS_PLANS && !user.IsPremium {
		logger.Debug().Msg("User has reached the maximum number of meal plans")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyMaxPlansReached, userLang))
		return
	}

	h.generatePlan(ctx, bot, update, user, userLang)
}

// handleWeekendPlanGeneration generates a meal plan for weekends
func (h *MealPlanHandler) handleWeekendPlanGeneration(ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
	user *database.User,
	userLang database.UserLanguage,
) {
	userID := uint(update.CallbackQuery.From.ID)

	logger := log.
		With().
		Uint("userID", userID).
		Logger()

	logger.Debug().Msg("Handling weekend plan generation")

	h.generatePlan(ctx, bot, update, user, userLang)
}

// generatePlan generates the actual meal plan
func (h *MealPlanHandler) generatePlan(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, user *database.User, userLang database.UserLanguage) {
	gc := &GenerationContext{
		Ctx:      ctx,
		Bot:      bot,
		Update:   update,
		User:     user,
		UserLang: userLang,
		Logger:   log.With().Uint("userID", user.ID).Logger(),
	}

	if err := h.cleanupPreviousMessage(gc); err != nil {
		return
	}

	planData, err := h.preparePlanData(gc)
	if err != nil {
		h.handlePlanError(gc, err)
		return
	}

	if err := h.sendAndPersistPlan(planData); err != nil {
		h.handlePlanError(gc, err)
	}
}

// GenerationContext bundles common execution parameters
type GenerationContext struct {
	Ctx      context.Context
	Bot      *tgbot.Bot
	Update   *tgmodels.Update
	User     *database.User
	UserLang database.UserLanguage
	Logger   zerolog.Logger
}

// PlanData contains meal plan generation parameters
type PlanData struct {
	LastAccepted  *database.UserMealPlan
	LastRejected  *database.UserMealPlan
	AcceptedCount int64
	*GenerationContext
}

func (h *MealPlanHandler) cleanupPreviousMessage(gc *GenerationContext) error {
	if err := h.removeMarkupOrDeleteMessage(gc.Ctx, gc.Bot, gc.Update, gc.UserLang); err != nil {
		gc.Logger.Err(err).Msg("Message cleanup failed")
		return err
	}
	return nil
}

func (h *MealPlanHandler) preparePlanData(gc *GenerationContext) (*PlanData, error) {
	lastAccepted, lastRejected, err := h.getPreviousPlans(gc.User.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous plans: %w", err)
	}

	acceptedCount, err := h.deps.DB.GetCountOfAcceptedPlans(gc.User.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accepted plans count: %w", err)
	}

	return &PlanData{
		LastAccepted:      lastAccepted,
		LastRejected:      lastRejected,
		AcceptedCount:     acceptedCount,
		GenerationContext: gc,
	}, nil
}

func (h *MealPlanHandler) getPreviousPlans(userID uint) (*database.UserMealPlan, *database.UserMealPlan, error) {
	lastAccepted, err := h.deps.DB.GetUserLastMealPlan(userID, database.PlanTypeWeekdays, true)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	lastRejected, err := h.deps.DB.GetUserLastMealPlan(userID, database.PlanTypeWeekdays, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	return lastAccepted, lastRejected, nil
}

func (h *MealPlanHandler) sendAndPersistPlan(data *PlanData) error {
	plan, err := h.generateMealPlan(data)
	if err != nil {
		return err
	}

	doc, err := h.createPlanDocument(plan, data.UserLang)
	if err != nil {
		return fmt.Errorf("document creation failed: %w", err)
	}

	messageID, err := h.sendPlanDocument(data, doc)
	if err != nil {
		return err
	}

	return h.saveUserPlan(data.User.ID, messageID, plan)
}

func (h *MealPlanHandler) generateMealPlan(data *PlanData) (models.CompleteMealPlan, error) {
	return h.deps.NutritionAgent.GenerateMealPlan(
		data.Ctx,
		data.User,
		data.LastAccepted,
		data.LastRejected,
		data.AcceptedCount,
		h.planType,
	)
}

func (h *MealPlanHandler) createPlanDocument(plan models.CompleteMealPlan, lang database.UserLanguage) (bytes.Buffer, error) {
	doc, err := GenerateDocFromMealPlan(plan, lang)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("pdf generation failed: %w", err)
	}
	return doc, nil
}

func (h *MealPlanHandler) sendPlanDocument(data *PlanData, doc bytes.Buffer) (int, error) {
	msg, err := data.Bot.SendDocument(data.Ctx, &tgbot.SendDocumentParams{
		ChatID:  data.Update.CallbackQuery.Message.Message.Chat.ID,
		Caption: i18n.GetText(i18n.KeyPlanCaption, data.UserLang),
		Document: &tgmodels.InputFileUpload{
			Filename: "personalized_meal_plan.pdf",
			Data:     &doc,
		},
		ReplyMarkup: h.createResponseButtons(data.UserLang),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to send document: %w", err)
	}
	return msg.ID, nil
}

func (h *MealPlanHandler) createResponseButtons(lang database.UserLanguage) *tgmodels.InlineKeyboardMarkup {
	return &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         i18n.GetText(i18n.KeyAcceptButton, lang),
					CallbackData: string(CallBackDataAcceptPlan),
				},
				{
					Text:         i18n.GetText(i18n.KeyRejectButton, lang),
					CallbackData: string(CallbackDataRejectPlan),
				},
			},
		},
	}
}

func (h *MealPlanHandler) saveUserPlan(userID uint, messageID int, plan models.CompleteMealPlan) error {
	input := h.mapPlanToInput(plan)
	err := h.deps.DB.SaveUserMealPlan(userID, messageID, h.planType, input)
	if err != nil {
		return fmt.Errorf("failed to save meal plan: %w", err)
	}
	return nil
}

func (h *MealPlanHandler) mapPlanToInput(plan models.CompleteMealPlan) []database.UserMealPlanFoodInput {
	var input []database.UserMealPlanFoodInput
	for _, food := range plan.Plan {
		input = append(input, database.UserMealPlanFoodInput{
			Name:     food.Food,
			Category: string(food.FoodCategory),
		})
	}
	return input
}

func (h *MealPlanHandler) handlePlanError(gc *GenerationContext, err error) {
	chatID := gc.Update.CallbackQuery.Message.Message.Chat.ID
	gc.Logger.Err(err).Msg("Plan generation error")
	ErrorResponse(gc.Ctx, gc.Bot, chatID, i18n.GetText(i18n.KeyErrorGeneratePlan, gc.UserLang))
}

// removeMarkupOrDeleteMessage modifies or removes the message that triggered
// the callback query. If the message has no media attached, it updates the
// message text to inform the user about the ongoing plan generation process.
// If the message contains media, it deletes the message instead.
func (h *MealPlanHandler) removeMarkupOrDeleteMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, userLang database.UserLanguage) error { // Add language
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
			Text:      i18n.GetText(i18n.KeyGeneratingPlan, userLang),
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
