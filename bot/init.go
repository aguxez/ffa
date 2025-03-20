package bot

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	"github.com/go-telegram/bot"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms/openai"
	"gorm.io/gorm"
)

type LLMBot struct {
	tgBot *tgbot.Bot
	llm   *agent.NutritionAgent
	db    *gorm.DB
	sm    *models.StateManager
}

// NewLLMBot creates a new telegram bot with an OpenRouter LLM backend.
//
// The bot is started and ready to receive messages.
//
// Telegram bot API key and OpenRouter API key are read from the environment
// variables TELEGRAM_BOT_API_KEY and OPENROUTER_API_KEY respectively.
func NewLLMBot(sm *models.StateManager, db *gorm.DB) (*LLMBot, error) {
	b, err := bot.New(os.Getenv("TELEGRAM_BOT_API_KEY"))
	if err != nil {
		return &LLMBot{}, err
	}

	llm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
		openai.WithModel("anthropic/claude-3.7-sonnet:beta"),
	)
	if err != nil {
		log.Err(err).Msg("Failed to create LLM")
		return &LLMBot{}, err
	}

	return &LLMBot{
		tgBot: b,
		llm:   agent.NewNutritionAgent(llm, sm),
		db:    db,
		sm:    sm,
	}, nil
}

func (b *LLMBot) RegisterHandlers() {
	b.tgBot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		b.startCommandHandler,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		"create_meal_plan",
		tgbot.MatchTypeExact,
		b.createMealPlanHandler,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		"add_user_food",
		tgbot.MatchTypeExact,
		b.addUserFoodHandler,
	)

	// Works as a match all handler
	b.tgBot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"",
		tgbot.MatchTypeContains,
		b.messageHandler,
	)
}

// Init initializes and starts the Telegram bot.
//
// It takes a models.StateManager and registers the start command and the
// create meal plan callback handler. It also starts the Telegram bot and
// returns an error if any.
func Init(sm *models.StateManager, db *gorm.DB) error {
	b, err := NewLLMBot(sm, db)
	if err != nil {
		return err
	}

	log.Info().Msg("Starting Telegram bot...")

	b.RegisterHandlers()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b.tgBot.Start(ctx)

	return nil
}

func (b *LLMBot) addUserFoodHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "addUserFoodHandler").Msg("Received message")

	// Set user state in the state manager
	b.sm.SetUserState(uint(update.CallbackQuery.From.ID), "add_user_food")

	_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   `Sure, send me some of the foods you enjoy so I can reference them. These items will be added to the ones you already have saved instead of replacing them.`,
	})
	if err != nil {
		log.Err(err).Msg("Error sending message to add user foods")
	}
}

func (b *LLMBot) createMealPlanHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "createMealPlanHandler").Msg("Received message")

	// First remove the markup button the user pressed
	_, err := bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error editing message")
	}

	// Then inform the user about the state
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "I'm generating a plan for you now.",
	})

	// Then proceed to generate the meal plan
	plan, err := b.llm.GenerateMealPlan(ctx)
	if err != nil {
		log.Err(err).Msg("Error generating meal plan")

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			Text: "There was an error generating the meal plan. This has been informed.",
		})
	}

	log.Debug().Msg("Created meal plan")

	doc, err := GenerateDocFromMealPlan(plan)
	if err != nil {
		log.Err(err).Msg("Error generating document")

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			Text: "There was an error generating your document. This has been informed.",
		})
	}

	log.Debug().Msg("Generated document from meal plan")

	bot.SendDocument(ctx, &tgbot.SendDocumentParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Document: &tgmodels.InputFileUpload{
			Filename: "personalized_meal_plan.pdf",
			Data:     &doc,
		},
	})
}

func (b *LLMBot) startCommandHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "startCommandHandler").Msg("Received message")

	// initialize user state
	b.sm.SetUserState(uint(update.Message.From.ID), "start")

	err := createUserIfNotExists(b.db, update)
	if err != nil {
		log.Err(err).Msg("Error creating user")
		return
	}

	// Send a message saying what the bot does and shows buttons
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "I can help you create meal plans tailored for your dietary preferences",
		ChatID: update.Message.Chat.ID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Create Meal Plan",
						CallbackData: "create_meal_plan",
					},
					{
						Text:         "Configure schedule",
						CallbackData: "configure_meal_plan_schedule",
					},
				},
				{
					{
						Text:         "Add food",
						CallbackData: "add_user_food",
					},
					{
						Text:         "Add schedule",
						CallbackData: "add_plan_schedule",
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending message")
		return
	}
}

func createUserIfNotExists(db *gorm.DB, update *tgmodels.Update) error {
	// Fetch user, create if it doesn't exist
	user, err := database.GetUser(db, uint(update.Message.From.ID))

	if user != nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		log.Err(err).Msg("Error fetching user")
		return err
	}

	log.Debug().Msg("User not found, creating...")

	err = database.CreateUser(db, uint(update.Message.From.ID))
	if err != nil {
		log.Err(err).Msg("Error creating user")
		return err
	}

	return nil
}

func (b *LLMBot) messageHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "messageHandler").Msg("Received message")

	// Guard against nil message or From field
	if update.Message == nil || update.Message.From == nil {
		log.Debug().Msg("Message or From field is nil")
		return
	}

	// Get user state - this could be nil if state doesn't exist yet
	userState := b.sm.GetUserState(uint(update.Message.From.ID))

	// Check if user state exists and has a valid conversation stage
	if userState == nil {
		log.Debug().Int64("userID", update.Message.From.ID).Msg("No valid state for user")
		return
	}

	// Now safely access the conversation stage
	switch userState.ConversationStage {
	case "add_user_food":
		b.addUserFoodFromMsg(ctx, bot, update)
	default:
		log.Debug().Msg("No message handler matched")
		return
	}
}

func (b *LLMBot) addUserFoodFromMsg(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := update.Message.From.ID
	log.Debug().Int64("userID", userID).Msg("Adding food to user")

	// Take the messages, for now assume they come separated by a newline.
	// TODO: Add an LLM call to parse message from natural language
	foodsFromMessage := strings.Split(update.Message.Text, "\n")

	err := database.SaveUserFoods(b.db, uint(update.Message.From.ID), foodsFromMessage)
	if err != nil {
		log.Err(err).Int64("userID", userID).Msg("Error saving user foods")

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			Text:   "There was an error saving your food. This has been informed. Please try again.",
			ChatID: update.Message.Chat.ID,
		})
	}

	// Set user state in the state manager
	b.sm.SetUserState(uint(update.Message.From.ID), "")

	// Reply to user
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "Successfuly updated your food preferences!",
		ChatID: update.Message.Chat.ID,
	})
}
