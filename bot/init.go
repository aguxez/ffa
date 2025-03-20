package bot

import (
	"context"
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
		string(CallbackDataCreateMealPlan),
		tgbot.MatchTypeExact,
		b.createMealPlanHandler,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		string(CallbackDataAddUserFood),
		tgbot.MatchTypeExact,
		b.addUserFoodHandler,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		string(CallbackDataAddPlanSchedule),
		tgbot.MatchTypeExact,
		b.addPlanSchedule,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		string(CallbackDataPlanScheduleWeekly),
		tgbot.MatchTypeExact,
		b.updatePlanScheduleFromMsg,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		string(CallbackDataPlanScheduleThreeDays),
		tgbot.MatchTypeExact,
		b.updatePlanScheduleFromMsg,
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
	b.sm.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageAddUserFood)

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

	// First, load the user foods
	userFoods, err := database.GetUserFoods(b.db, uint(update.CallbackQuery.From.ID))
	if err != nil {
		log.Err(err).Msg("Error getting user foods")
	}

	if len(userFoods) == 0 {
		// TODO: Send a message saying we have not saved food preferences yet
		log.Debug().Int64("user_id", update.CallbackQuery.From.ID).Msg("No user foods found")
		return
	}

	// then remove the markup button the user pressed
	_, err = bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
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
	plan, err := b.llm.GenerateMealPlan(ctx, userFoods)
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

	userID := uint(update.Message.From.ID)

	// initialize user state
	b.sm.SetUserState(userID, models.ConversationStageStart)

	// Every time this is called we need to ensure the user exists
	_, err := database.GetOrCreateUser(b.db, userID)
	if err != nil {
		log.Err(err).Msg("Error finding or creating user")
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
		log.Err(err).Msg("Error sending message")
		return
	}
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

	switch userState.Stage {

	case models.ConversationStageAddUserFood:
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

		return
	}

	// Set user state in the state manager
	b.sm.SetUserState(uint(update.Message.From.ID), models.ConversationStageNull)

	// Reply to user
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "Successfuly updated your food preferences!",
		ChatID: update.Message.Chat.ID,
	})
}

func (b *LLMBot) addPlanSchedule(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "addPlanSchedule").Msg("Received message")

	// Set user state in the state manager
	b.sm.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageAddPlanSchedule)

	// // Find user plan schedule
	// user, err := database.GetUser(b.db, uint(update.Message.From.ID))
	// if err != nil {
	// 	log.Err(err).Msg("Error fetching user")
	// 	return
	// }

	// Delete message with the markup since we cannot modify the text of the message
	// if we wish to edit it
	_, err := bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		log.Err(err).Msg("Error deleting message")
		return
	}

	// Send an inline keyboard with the options available for intervals
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Please select an interval for your plan schedule.",
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Weekly",
						CallbackData: string(CallbackDataPlanScheduleWeekly),
					},
					{
						Text:         "Every 3 days",
						CallbackData: string(CallbackDataPlanScheduleThreeDays),
					},
				},
			},
		},
	})
	if err != nil {
		log.Err(err).Msg("Error sending plan schedule message")
		return
	}
}

func (b *LLMBot) updatePlanScheduleFromMsg(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "updatePlanScheduleFromMsg").Msg("Received message")

	// Set user state in the state manager
	b.sm.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageNull)

	// Update settings in database
	var intervalDays uint
	if update.CallbackQuery.Data == string(CallbackDataPlanScheduleWeekly) {
		intervalDays = 7
	} else if update.CallbackQuery.Data == string(CallbackDataPlanScheduleThreeDays) {
		intervalDays = 3
	}
	err := database.UpdateUserPlanSchedule(b.db, uint(update.CallbackQuery.From.ID), intervalDays)
	if err != nil {
		log.Err(err).Msg("Error updating user plan schedule")
		return
	}

	// Delete message with the markup since we cannot modify the text of the message
	// if we wish to edit it
	_, err = bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		log.Err(err).Msg("Error deleting message")
		return
	}

	// Send message informing about the update success
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Successfuly updated your plan schedule!",
	})
}
