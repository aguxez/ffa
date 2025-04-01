package bot

import (
	"context"
	"fmt"
	"strconv" // Added for parsing offset
	"strings"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const (
	UserFoodPageSize = 5
)

// UserFoodHandler handles adding user food preferences
type UserFoodHandler struct {
	deps *HandlerDependencies
}

// NewUserFoodHandler creates a new UserFoodHandler
func NewUserFoodHandler(deps *HandlerDependencies) *UserFoodHandler {
	return &UserFoodHandler{
		deps: deps,
	}
}

// Handle handles the user food preference addition
func (h *UserFoodHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "UserFoodHandler").Msg("Handling user food addition")

	userID := uint(update.CallbackQuery.From.ID)

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddUserFood)

	// Send instructions
	h.editMessage(ctx, bot, update)
}

// UserFoodMessageHandler handles processing food items from user messages
type UserFoodMessageHandler struct {
	deps *HandlerDependencies
}

// NewUserFoodMessageHandler creates a new UserFoodMessageHandler
func NewUserFoodMessageHandler(deps *HandlerDependencies) *UserFoodMessageHandler {
	return &UserFoodMessageHandler{
		deps: deps,
	}
}

// Handle processes a message containing food items
func (h *UserFoodMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	logger := log.With().Str("handler", "UserFoodMessageHandler").Uint("userID", uint(update.Message.From.ID)).Logger()

	logger.Debug().Msg("Processing food message")

	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	// Parse food items from natural language
	userFoods, err := h.deps.NaturalLanguageAgent.ParseUserFoodsMessage(ctx, update.Message.Text)
	if err != nil {
		logger.Err(err).Msg("Error parsing user food message")
		ErrorResponse(ctx, bot, chatID, "There was an error parsing your food preferences. Please try again.")
		return
	}

	// Save to database
	err = h.deps.DB.SaveUserFoods(userID, userFoods)
	if err != nil {
		logger.Err(err).Msg("Error saving user foods")
		ErrorResponse(ctx, bot, chatID, "There was an error saving your food preferences. Please try again.")
		return
	}

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Confirm to user
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   "Successfully updated your food preferences!",
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "↼ Back to options",
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
}

// editMessage removes the inline keyboard markup and edits the
// previous message's text
func (h *UserFoodHandler) editMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) error {
	// Edits the previous message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      `Share foods you enjoy to include in your meal plans. Any items you add will be combined with your existing saved favorites.`,
	})

	return err
}

// ViewUserFoodHandler handles viewing user food preferences
type ViewUserFoodHandler struct {
	deps *HandlerDependencies
}

// NewViewUserFoodHandler creates a new ViewUserFoodHandler
func NewViewUserFoodHandler(deps *HandlerDependencies) *ViewUserFoodHandler {
	return &ViewUserFoodHandler{
		deps: deps,
	}
}

// Handle handles the user food preference view
func (v *ViewUserFoodHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	messageID := update.CallbackQuery.Message.Message.ID

	logger := log.With().Str("handler", "ViewUserFoodHandler").Uint("userID", userID).Logger()

	logger.Debug().Msg("Handling user food view callback")

	// Default to 0 for initial view
	offset := 0

	callbackData := update.CallbackQuery.Data

	// Parses the page number from the callback data
	if strings.HasPrefix(callbackData, string(CallbackDataViewUserFoodOffset)) {
		offsetStr := strings.TrimPrefix(callbackData, string(CallbackDataViewUserFoodOffset))
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Err(err).Str("callbackData", callbackData).Msg("Error parsing offset from callback data")
			ErrorResponse(ctx, bot, chatID, "Invalid pagination request.")
			return
		}

		if parsedOffset < 0 {
			logger.Warn().Int("parsedOffset", parsedOffset).Msg("Received negative offset, defaulting to 0")
			parsedOffset = 0
		}

		offset = parsedOffset
		logger.Debug().Int("offset", offset).Msg("Handling pagination request")
	} else {
		logger.Debug().Msg("Handling initial view request")
	}

	v.sendPaginatedFoods(ctx, bot, chatID, messageID, userID, offset)
}

// sendPaginatedFoods fetches and sends a specific page of user foods
func (v *ViewUserFoodHandler) sendPaginatedFoods(ctx context.Context, bot *tgbot.Bot, chatID int64, messageID int, userID uint, offset int) {
	logger := log.With().Str("handler", "ViewUserFoodHandler").Uint("userID", userID).Int("offset", offset).Logger()

	userFoods, err := v.deps.DB.GetUserFoods(userID, UserFoodPageSize, offset)
	if err != nil {
		logger.Err(err).Msg("Error finding user foods")
		ErrorResponse(ctx, bot, chatID, "Error retrieving your food list.")
		return
	}

	// This logic works similar to cursor based pagination so we do not fetch all items of a user to know
	// whether or not we can show the next page
	hasNextPage := len(userFoods) > UserFoodPageSize
	if hasNextPage {
		// Remove the extra item used for checking existence of the next page.
		userFoods = userFoods[:UserFoodPageSize]
	}

	var messageText strings.Builder
	if len(userFoods) == 0 && offset == 0 {
		messageText.WriteString("You haven't added any favorite foods yet.")
	} else if len(userFoods) == 0 && offset > 0 {
		// Should ideally not happen with correct button logic, but handle defensively
		messageText.WriteString("No more foods found.")
	} else {
		messageText.WriteString(fmt.Sprintf("Your Foods (Items %d-%d):\n", offset+1, offset+len(userFoods)))
		for i, food := range userFoods {
			messageText.WriteString(fmt.Sprintf("%d. %s\n", offset+i+1, food))
		}
	}

	// Build keyboard
	keyboard := [][]tgmodels.InlineKeyboardButton{}
	navigationRow := []tgmodels.InlineKeyboardButton{}

	if offset > 0 {
		prevOffset := offset - UserFoodPageSize

		if prevOffset < 0 {
			// Ensure offset doesn't go below 0
			prevOffset = 0
		}
		navigationRow = append(navigationRow, tgmodels.InlineKeyboardButton{
			Text:         "⬅️ Previous",
			CallbackData: fmt.Sprintf("%s%d", CallbackDataViewUserFoodOffset, prevOffset),
		})
	}

	if hasNextPage {
		nextOffset := offset + UserFoodPageSize
		navigationRow = append(navigationRow, tgmodels.InlineKeyboardButton{
			Text:         "Next ➡️",
			CallbackData: fmt.Sprintf("%s%d", CallbackDataViewUserFoodOffset, nextOffset),
		})
	}

	if len(navigationRow) > 0 {
		keyboard = append(keyboard, navigationRow)
	}

	// Add Back button
	keyboard = append(keyboard, []tgmodels.InlineKeyboardButton{
		{Text: "↼ Back to options", CallbackData: string(CallbackDataBackToOptions)},
	})

	// Send edited message
	_, err = bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        messageText.String(),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{InlineKeyboard: keyboard},
	})
	if err != nil {
		logger.Err(err).Msg("Error editing message for food view")
	}
}
