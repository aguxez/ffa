package bot

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/models"
	"github.com/go-telegram/bot"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/tmc/langchaingo/llms/openai"
)

type LLMBot struct {
	tgBot *tgbot.Bot
	llm   *agent.NutritionAgent
}

// NewLLMBot creates a new telegram bot with an OpenRouter LLM backend.
//
// The bot is started and ready to receive messages.
//
// Telegram bot API key and OpenRouter API key are read from the environment
// variables TELEGRAM_BOT_API_KEY and OPENROUTER_API_KEY respectively.
func NewLLMBot(sm *models.StateManager) (*LLMBot, error) {
	b, err := bot.New(os.Getenv("TELEGRAM_BOT_API_KEY"), tgbot.WithDefaultHandler(defaultMsgHandler))
	if err != nil {
		return &LLMBot{}, err
	}

	llm, err := openai.New(
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithToken(os.Getenv("OPENROUTER_API_KEY")),
		openai.WithModel("anthropic/claude-3.7-sonnet:beta"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// agent := agent.NewNutritionAgent(llm, sm)

	return &LLMBot{
		tgBot: b,
		llm:   agent.NewNutritionAgent(llm, &models.StateManager{}),
	}, nil
}

func (b *LLMBot) RegisterHandlers() {
	b.tgBot.RegisterHandler(tgbot.HandlerTypeMessageText,
		"/start",
		tgbot.MatchTypeExact,
		startCommandHandler,
	)

	b.tgBot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData,
		"create_meal_plan",
		tgbot.MatchTypeExact,
		b.createMealPlanHandler,
	)
}

func Init(sm *models.StateManager) error {
	b, err := NewLLMBot(sm)
	if err != nil {
		return err
	}

	log.Print("Starting Telegram bot...")

	b.RegisterHandlers()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b.tgBot.Start(ctx)

	return nil
}

func (b *LLMBot) createMealPlanHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Print("Generating meal plan...")

	// First remove the markup button the user pressed
	_, err := bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
		},
	})
	if err != nil {
		log.Printf("Error editing message: %v", err)
	}

	// Then inform the user about the state
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "I'm generating a plan for you now.",
	})

	// Then proceed to generate the meal plan
	plan, err := b.llm.GenerateMealPlan(ctx)
	if err != nil {
		log.Printf("Error generating meal plan: %v", err)

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			Text: "There was an error generating the meal plan. This has been informed.",
		})
	}

	log.Print("Created meal plan")

	doc, err := GenerateDocFromMealPlan(plan)
	if err != nil {
		log.Printf("Error generating document: %v", err)

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			Text: "There was an error generating your document. This has been informed.",
		})
	}

	log.Print("Generated document from meal plan")

	bot.SendDocument(ctx, &tgbot.SendDocumentParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Document: &tgmodels.InputFileUpload{
			Filename: "personalized_meal_plan.pdf",
			Data:     &doc,
		},
	})
}

func startCommandHandler(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Print("Received call on startCommandHandler")

	// Send a message saying what the bot does and shows buttons
	_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "I can help you create meal plans tailored for your dietary preferences",
		ChatID: update.Message.Chat.ID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Create Meal Plan",
						CallbackData: "create_meal_plan",
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}
}

func defaultMsgHandler(_ context.Context, _ *bot.Bot, update *tgmodels.Update) {
	log.Printf("Received update: %v", update.Message.Text)
}
