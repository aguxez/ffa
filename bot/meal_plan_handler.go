package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MealPlanHandler handles meal plan creation
type MealPlanHandler struct {
	deps *HandlerDependencies
}

// NewMealPlanHandler creates a new MealPlanHandler
func NewMealPlanHandler(deps *HandlerDependencies) *MealPlanHandler {
	return &MealPlanHandler{
		deps: deps,
	}
}

// Handle handles the meal plan creation
func (h *MealPlanHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "MealPlanHandler").Msg("Handling meal plan creation")

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := uint(update.CallbackQuery.From.ID)

	// Load user foods
	userFoods, err := database.GetUserFoods(h.deps.DB, userID)
	if err != nil {
		log.Err(err).Msg("Error getting user foods")
		ErrorResponse(ctx, bot, chatID, "Failed to retrieve your food preferences.")
		return
	}

	if len(userFoods) == 0 {
		log.Debug().Uint("user_id", userID).Msg("No user foods found")
		ErrorResponse(ctx, bot, chatID, "Please add your food preferences first.")
		return
	}

	// Remove the markup button
	if err := h.removeMarkup(ctx, bot, update); err != nil {
		log.Err(err).Msg("Error removing markup")
		return
	}

	// Inform user about generation
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "I'm generating a plan for you now.",
	})

	// Generate meal plan
	plan, err := h.deps.Agent.GenerateMealPlan(ctx, userFoods)
	if err != nil {
		log.Err(err).Msg("Error generating meal plan")
		ErrorResponse(ctx, bot, chatID, "There was an error generating the meal plan. This has been informed.")
		return
	}

	log.Debug().Msg("Created meal plan")

	// Generate PDF
	doc, err := GenerateDocFromMealPlan(plan)
	if err != nil {
		log.Err(err).Msg("Error generating document")
		ErrorResponse(ctx, bot, chatID, "There was an error generating your document. This has been informed.")
		return
	}

	log.Debug().Msg("Generated document from meal plan")

	// Send document
	bot.SendDocument(ctx, &tgbot.SendDocumentParams{
		ChatID: chatID,
		Document: &tgmodels.InputFileUpload{
			Filename: "personalized_meal_plan.pdf",
			Data:     &doc,
		},
	})
}

// removeMarkup removes the inline keyboard markup
func (h *MealPlanHandler) removeMarkup(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) error {
	_, err := bot.EditMessageReplyMarkup(ctx, &tgbot.EditMessageReplyMarkupParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{},
		},
	})
	return err
}
