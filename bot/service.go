package bot

import (
	"context"
	"os"
	"os/signal"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/config"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
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

// NewBotService creates a new BotService
func NewBotService(stateManager *models.StateManager, db *database.Database, cfg config.AppConfig) (*BotService, error) {
	// Create telegram bot
	bot, err := tgbot.New(cfg.TelegramBotAPIKey)
	if err != nil {
		return nil, err
	}

	// Create LLM for nutrition tasks
	nutritionLlm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(cfg.OpenRouterAPIKey),
		openai.WithModel(string(agent.LLMModelClaude)),
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
		openai.WithModel(string(agent.LLMModelGemini)),
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

	// Register callback handlers
	callbackHandlers := map[CallbackData]Handler{
		CallbackDataCreateMealPlan:        NewMealPlanHandler(handlerDeps),
		CallbackDataAddUserFood:           NewUserFoodHandler(handlerDeps),
		CallbackDataAddPlanSchedule:       NewPlanScheduleHandler(handlerDeps),
		CallbackDataPlanScheduleWeekly:    NewPlanScheduleUpdateHandler(handlerDeps, 7),
		CallbackDataPlanScheduleThreeDays: NewPlanScheduleUpdateHandler(handlerDeps, 3),
		// We wat this command to always return to the 'main menu' so to speak
		CallbackDataBackToOptions: NewBackToOptionsHandler(handlerDeps, true),

		// This one triggers the same event as `CallbackDataBackToOptions`
		// but doesn't delete the previous state. This is important to maintain
		// conversation context like if we send a meal plan to the user. We do not want
		// the message to be deleted when they go back.
		CallbackDataBackToOptionsNoDelete:  NewBackToOptionsHandler(handlerDeps, false),
		CallbackDataAddMacroTargets:        NewMacroTargetsHandler(handlerDeps),
		CallbackDataSetDietaryRestrictions: NewSetDietaryRestrictionsHandler(handlerDeps),
		CallbackDataSetBodyWeight:          NewSetBodyWeightHandler(handlerDeps),

		CallBackDataAcceptPlan: NewPlanAcceptanceHandler(handlerDeps, true),
		CallbackDataRejectPlan: NewPlanAcceptanceHandler(handlerDeps, false),
	}

	for callbackData, handler := range callbackHandlers {
		// Register handlers to the registry
		s.handlers.Register(string(callbackData), handler)

		// Register callback handlers
		s.bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
			string(callbackData),
			tgbot.MatchTypeExact,
			handler.Handle,
		)
	}

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
func (s *BotService) SendMessage(user database.User, message string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_, err := s.bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: user.ChatID,
		Text:   message,
	})
	return err
}
