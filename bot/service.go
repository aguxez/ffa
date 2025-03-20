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
	bot          *tgbot.Bot
	agent        *agent.NutritionAgent
	db           *gorm.DB
	stateManager *models.StateManager
	handlers     *HandlerRegistry
}

// NewBotService creates a new BotService
func NewBotService(stateManager *models.StateManager, db *gorm.DB) (*BotService, error) {
	// Create telegram bot
	bot, err := tgbot.New(os.Getenv("TELEGRAM_BOT_API_KEY"))
	if err != nil {
		return nil, err
	}

	// Create LLM
	llm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
		openai.WithModel("anthropic/claude-3.7-sonnet:beta"),
	)
	if err != nil {
		log.Err(err).Msg("Failed to create LLM")
		return nil, err
	}

	// Create agent
	agent := agent.NewNutritionAgent(llm, stateManager)

	// Create handler registry
	handlers := NewHandlerRegistry()

	return &BotService{
		bot:          bot,
		agent:        agent,
		db:           db,
		stateManager: stateManager,
		handlers:     handlers,
	}, nil
}

// RegisterHandlers registers all handlers
func (s *BotService) RegisterHandlers() {
	// Create handler dependencies
	handlerDeps := &HandlerDependencies{
		Agent:        s.agent,
		DB:           s.db,
		StateManager: s.stateManager,
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
