package bot

import (
	"context"
	"os"
	"os/signal"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
	"gorm.io/gorm"
)

// BotService encapsulates the Telegram bot functionality
type BotService struct {
	bot                  *tgbot.Bot
	nutritionAgent       *agent.NutritionAgent
	naturalLanguageAgent *agent.NaturalLanguageAgent
	db                   *gorm.DB
	stateManager         *models.StateManager
	handlers             *HandlerRegistry
}

// NewBotService creates a new BotService
func NewBotService(stateManager *models.StateManager, db *gorm.DB) (*BotService, error) {
	// Create telegram bot
	bot, err := tgbot.New(os.Getenv("TELEGRAM_BOT_API_KEY"))
	if err != nil {
		return nil, err
	}

	// Create LLM for nutrition tasks
	nutritionLlm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
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
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
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
		CallbackDataBackToOptionsNoDelete: NewBackToOptionsHandler(handlerDeps, false),
		CallbackDataAddMacroTargets:       NewMacroTargetsHandler(handlerDeps),
	}

	for callbackData, handler := range callbackHandlers {
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
func Init(stateManager *models.StateManager, db *gorm.DB) error {
	service, err := NewBotService(stateManager, db)
	if err != nil {
		return err
	}

	service.RegisterHandlers()
	return service.Start()
}
