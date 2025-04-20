package bot

import (
	"context"
	"os"
	"os/signal"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/config"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
	"github.com/aguxez/ffa/messaging"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
)

// BotService encapsulates the Telegram bot functionality
type BotService struct {
	bot                  *tgbot.Bot
	nutritionAgent       *agent.NutritionAgent
	naturalLanguageAgent *agent.NaturalLanguageAgent
	db                   *database.Database
	stateManager         *models.StateManager
	handlers             *HandlerRegistry
}

type HandlerWithConstraints struct {
	handler   Handler
	isPremium bool
}

// NewBotService creates a new BotService
func NewBotService(stateManager *models.StateManager, db *database.Database, cfg config.AppConfig) (*BotService, error) {
	opts := make([]tgbot.Option, 0)
	if cfg.Environment == "development" {
		opts = append(opts, tgbot.UseTestEnvironment())
	}

	// Create telegram bot
	bot, err := tgbot.New(cfg.TelegramBotAPIKey, opts...)
	if err != nil {
		return nil, err
	}

	// Create LLM for nutrition tasks
	nutritionLlm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(cfg.OpenRouterAPIKey),
		openai.WithModel(string(agent.LLMModelGeminiTwoFlash)),
	)
	if err != nil {
		log.Err(err).Msg("Failed to create LLM")
		return nil, err
	}

	// Create nutrition agent
	nutritionAgent := agent.NewNutritionAgent(nutritionLlm, stateManager)

	// Create LLM for natural language tasks.
	naturalLanguageLlm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(cfg.OpenRouterAPIKey),
		openai.WithModel(string(agent.LLMModelGeminiTwoFlashLite)),
	)
	if err != nil {
		log.Err(err).Msg("Failed to create LLM")
		return nil, err
	}
	naturalLanguageAgent := agent.NewNaturalLanguageAgent(naturalLanguageLlm)

	// Create handler registry
	handlers := NewHandlerRegistry()

	return &BotService{
		bot:                  bot,
		nutritionAgent:       nutritionAgent,
		naturalLanguageAgent: naturalLanguageAgent,
		db:                   db,
		stateManager:         stateManager,
		handlers:             handlers,
	}, nil
}

// RegisterHandlers registers all handlers
func (s *BotService) RegisterHandlers() {
	// Create handler dependencies
	handlerDeps := &HandlerDependencies{
		NutritionAgent:       s.nutritionAgent,
		DB:                   s.db,
		StateManager:         s.stateManager,
		NaturalLanguageAgent: s.naturalLanguageAgent,
		HandlerRegistry:      s.handlers,
	}

	// Register command handlers
	s.bot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		NewStartCommandHandler(handlerDeps).Handle,
	)

	// Menu is the same handler as start
	s.bot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"/menu",
		tgbot.MatchTypeExact,
		NewStartCommandHandler(handlerDeps).Handle,
	)

	feedbackRequestHandler := NewMealPlanFeedbackHandler(handlerDeps)
	paymentHandler := NewPaymentHandler(handlerDeps)

	// Register callback handlers
	callbackHandlers := map[CallbackData]HandlerWithConstraints{
		// Meal Plan Creation and Configuration Handlers
		// These handlers manage the initial setup and configuration of meal plans
		CallbackDataCreateMealPlan:         {handler: NewMealPlanHandler(handlerDeps, database.PlanTypeWeekdays), isPremium: false},
		CallbackDataAddUserFood:            {handler: NewUserFoodHandler(handlerDeps), isPremium: false},
		CallbackDataSetDietaryRestrictions: {handler: NewSetDietaryRestrictionsHandler(handlerDeps), isPremium: false},
		CallbackDataCreateWeekendPlan:      {handler: NewMealPlanHandler(handlerDeps, database.PlanTypeWeekends), isPremium: true},

		// Meal Plan Schedule Handlers
		// These handlers manage the frequency and scheduling of meal plans
		CallbackDataAddPlanSchedule:       {handler: NewPlanScheduleHandler(handlerDeps), isPremium: true},
		CallbackDataPlanScheduleWeekly:    {handler: NewPlanScheduleUpdateHandler(handlerDeps, 7), isPremium: true},
		CallbackDataPlanScheduleThreeDays: {handler: NewPlanScheduleUpdateHandler(handlerDeps, 3), isPremium: true},

		// User Profile and Nutrition Target Handlers
		// These handlers manage user-specific data and nutritional goals
		CallbackDataAddMacroTargets: {handler: NewMacroTargetsHandler(handlerDeps), isPremium: true},
		CallbackDataSetBodyWeight:   {handler: NewSetBodyWeightHandler(handlerDeps), isPremium: true},
		// The callback to view the user foods is handled separately below with prefix matching

		// Navigation and Control Handlers
		// These handlers manage user interface navigation
		// We want this command to always return to the 'main menu' so to speak
		CallbackDataBackToOptions: {handler: NewBackToOptionsHandler(handlerDeps, true), isPremium: false},
		// This one triggers the same event as `CallbackDataBackToOptions`
		// but doesn't delete the previous state. This is important to maintain
		// conversation context like if we send a meal plan to the user. We do not want
		// the message to be deleted when they go back.
		CallbackDataBackToOptionsNoDelete: {handler: NewBackToOptionsHandler(handlerDeps, false), isPremium: false},

		// Plan Review Handlers
		// These handlers manage user acceptance or rejection of proposed meal plans
		CallBackDataAcceptPlan: {handler: NewPlanAcceptanceHandler(handlerDeps, true), isPremium: false},
		CallbackDataRejectPlan: {handler: NewPlanAcceptanceHandler(handlerDeps, false), isPremium: false},

		// Meal Plan Feedback Handlers
		// These handlers collect and process user feedback on meal plans
		CallbackDataMealPlanFeedbackStart: {handler: feedbackRequestHandler, isPremium: false},
		// Breakfast Feedback
		CallbackDataMealPlanFeedbackBreakfastGood: {handler: feedbackRequestHandler, isPremium: false},
		CallbackDataMealPlanFeedbackBreakfastBad:  {handler: feedbackRequestHandler, isPremium: false},
		// Lunch Feedback
		CallbackDataMealPlanFeedbackLunchGood: {handler: feedbackRequestHandler, isPremium: false},
		CallbackDataMealPlanFeedbackLunchBad:  {handler: feedbackRequestHandler, isPremium: false},
		// Dinner Feedback
		CallbackDataMealPlanFeedbackDinnerGood: {handler: feedbackRequestHandler, isPremium: false},
		CallbackDataMealPlanFeedbackDinnerBad:  {handler: feedbackRequestHandler, isPremium: false},

		// Payments
		CallbackDataMakePayment: {handler: paymentHandler, isPremium: false},

		// Settings
		CallbackDataChangeLanguage:   {handler: NewSettingsHandler(handlerDeps), isPremium: false},
		CallbackDataChangeLanguageEn: {handler: NewLanguageHandler(handlerDeps, CallbackDataChangeLanguageEn), isPremium: false},
		CallbackDataChangeLanguageEs: {handler: NewLanguageHandler(handlerDeps, CallbackDataChangeLanguageEs), isPremium: false},
	}

	// Register exact match handlers first
	for callbackData, handlerWithConstraints := range callbackHandlers {
		// Register handlers to the registry
		s.handlers.Register(string(callbackData), handlerWithConstraints.handler)

		// Register callback handlers
		s.bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
			string(callbackData),
			tgbot.MatchTypeExact,
			// Use the handler from the struct
			handlerWithConstraints.handler.Handle,
			tgbot.Middleware(func(next tgbot.HandlerFunc) tgbot.HandlerFunc {
				return handlerDeps.EnsurePremium(next, handlerWithConstraints)
			}),
		)
	}

	// Register prefix match handler for food pagination AFTER exact matches
	viewUserFoodHandler := NewViewUserFoodHandler(handlerDeps)

	s.bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		// Use the constant prefix
		string(CallbackDataViewUserFoodOffset),
		// Match callbacks starting with the prefix
		tgbot.MatchTypePrefix,
		viewUserFoodHandler.Handle,
	)
	// Also register the initial exact match for viewing foods
	s.bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		string(CallbackDataViewUserFoods),
		tgbot.MatchTypeExact,
		viewUserFoodHandler.Handle,
	)

	// These handlers are specifically for payments. Whenever we have a successful payment,
	// we do not want it to go through the default conversational handler, but still use the
	// payment handler
	s.bot.RegisterHandlerMatchFunc(func(update *tgmodels.Update) bool {
		return update.Message != nil && update.Message.SuccessfulPayment != nil
	}, paymentHandler.Handle)

	s.bot.RegisterHandlerMatchFunc(func(update *tgmodels.Update) bool {
		return update.PreCheckoutQuery != nil
	}, paymentHandler.Handle)

	// Register message handler for conversation
	s.bot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"",
		tgbot.MatchTypeContains,
		NewMessageHandler(handlerDeps).Handle,
	)
}

// Start starts the bot and blocks until interrupted
func (s *BotService) Start() error {
	log.Info().Msg("Starting Telegram bot...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	s.bot.Start(ctx)

	return nil
}

// Init initializes and starts the Telegram bot
func (s *BotService) Init() error {
	s.RegisterHandlers()
	return s.Start()
}

// SendMessage sends a message to a user on telegram.
// Part of the `messaging.Messenger` interface
func (s *BotService) SendMessage(user database.User, message string, options ...messaging.Option) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	defaultOpts := messaging.Options{
		WithInlineKeyboardFeedback: false,
	}

	for _, option := range options {
		option(&defaultOpts)
	}

	inlineKeyboardButtons := [][]tgmodels.InlineKeyboardButton{}
	// Determine user language for button text
	userLang := user.Language
	if defaultOpts.WithInlineKeyboardFeedback {
		inlineKeyboardButtons = append(inlineKeyboardButtons, []tgmodels.InlineKeyboardButton{
			{
				Text:         i18n.GetText(i18n.KeyFeedbackStartButton, userLang),
				CallbackData: string(CallbackDataMealPlanFeedbackStart),
			},
		})
	}

	_, err := s.bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: *user.ChatID, // Dereference pointer
		Text:   message,      // Assume message is already constructed/translated if needed by caller
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: inlineKeyboardButtons,
		},
	})
	return err
}

// EnsurePremium is a middleware that ensures a user is premium before allowing
// them to do something. If the user is not premium, a message is sent to the
// user explaining that the feature requires a premium subscription and a button
// is presented for the user to upgrade.
func (deps *HandlerDependencies) EnsurePremium(next tgbot.HandlerFunc, handlerWithConstraints HandlerWithConstraints) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
		userID := uint(0)
		chatID := int64(0)

		// Determine userID, chatID, and attempt to get user language
		if update.Message != nil {
			userID = uint(update.Message.From.ID)
			chatID = update.Message.Chat.ID
		} else if update.CallbackQuery != nil {
			userID = uint(update.CallbackQuery.From.ID)
			if update.CallbackQuery.Message.Message != nil {
				chatID = update.CallbackQuery.Message.Message.Chat.ID
			}
		}
		// If userID is still 0, something is wrong
		if userID == 0 || chatID == 0 {
			log.Error().Msg("Could not determine userID or chatID in EnsurePremium middleware")
			return
		}

		logger := log.With().Uint("userID", userID).Int64("chatID", chatID).Logger()

		// Attempt to get user language
		user, err := deps.DB.GetUser(userID)
		if err != nil {
			logger.Warn().Err(err).Msg("Could not get user language in EnsurePremium, defaulting to English")
		}
		userLang := user.Language

		// Check premium status if required by the handler
		if handlerWithConstraints.isPremium {
			isPremium, err := deps.DB.IsUserPremium(userID)
			if err != nil {
				logger.Err(err).Msg("Error checking if user is premium")
				if chatID != 0 {
					bot.SendMessage(ctx, &tgbot.SendMessageParams{
						ChatID: chatID,
						Text:   i18n.GetText(i18n.KeyErrorPremiumCheck, userLang),
					})
				}
				return
			}

			if !isPremium {
				logger.Warn().Msg("User is not premium but tried to access a premium feature")
				if chatID != 0 {
					bot.SendMessage(ctx, &tgbot.SendMessageParams{
						ChatID: chatID,
						Text:   i18n.GetText(i18n.KeyPremiumRequired, userLang),
						ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
							InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
								{
									{
										Text:         i18n.GetText(i18n.KeyUpgradePremiumButton, userLang),
										CallbackData: string(CallbackDataMakePayment),
									},
								},
							},
						},
					})
				}
				// Stop processing for non-premium user on premium feature
				return
			}
		}

		// User is premium or feature doesn't require it, proceed to the actual handler
		next(ctx, bot, update)
	}
}
