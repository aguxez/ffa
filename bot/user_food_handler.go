package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
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

// Handle handles the user food preference addition prompt
func (h *UserFoodHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "UserFoodHandler").Logger()
	logger.Debug().Msg("Handling user food addition prompt")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Update user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageAddUserFood)

	// Send instructions
	h.editMessage(ctx, bot, update, userLang)
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
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	logger := log.With().Uint("userID", userID).Str("handler", "UserFoodMessageHandler").Logger()
	logger.Debug().Msg("Processing food message")

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Parse food items from natural language
	userFoods, err := h.deps.NaturalLanguageAgent.ParseUserFoodsMessage(ctx, update.Message.Text)
	if err != nil {
		logger.Err(err).Msg("Error parsing user food message")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorParseFoodPrefs, userLang))
		return
	}

	// Save to database
	err = h.deps.DB.SaveUserFoods(userID, userFoods)
	if err != nil {
		logger.Err(err).Msg("Error saving user foods")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorSaveFoodPrefs, userLang))
		return
	}

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// Confirm to user
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		Text:   i18n.GetText(i18n.KeyUserFoodUpdated, userLang),
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         i18n.GetText(i18n.KeyBackToOptionsButton, userLang),
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending user food update confirmation")
	}
}

// editMessage removes the inline keyboard markup and edits the
// previous message's text to show the food adding prompt
func (h *UserFoodHandler) editMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update, userLang database.UserLanguage) error { // Add language
	// Edits the previous message's text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      i18n.GetText(i18n.KeyUserFoodAddPrompt, userLang),
	})
	if err != nil {
		log.Err(err).Msg("Error editing message for user food prompt")
	}
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

	logger := log.With().Uint("userID", userID).Str("handler", "ViewUserFoodHandler").Logger()
	logger.Debug().Msg("Handling user food view callback")

	// Get user language first
	user, err := v.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Default to 0 for initial view
	offset := 0

	callbackData := update.CallbackQuery.Data

	// Parses the page number from the callback data
	if strings.HasPrefix(callbackData, string(CallbackDataViewUserFoodOffset)) {
		offsetStr := strings.TrimPrefix(callbackData, string(CallbackDataViewUserFoodOffset))
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Err(err).Str("callbackData", callbackData).Msg("Error parsing offset from callback data")
			ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyInvalidPaginationRequest, userLang))
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

	v.sendPaginatedFoods(ctx, bot, chatID, messageID, userID, offset, userLang)
}

// sendPaginatedFoods fetches and sends a specific page of user foods
func (v *ViewUserFoodHandler) sendPaginatedFoods(ctx context.Context, bot *tgbot.Bot, chatID int64, messageID int, userID uint, offset int, userLang database.UserLanguage) { // Add language
	logger := log.With().Uint("userID", userID).Int("offset", offset).Str("handler", "ViewUserFoodHandler").Logger()

	// Fetch one extra item to check if there's a next page
	userFoods, err := v.deps.DB.GetUserFoods(userID, UserFoodPageSize+1, offset)
	if err != nil {
		logger.Err(err).Msg("Error finding user foods")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorRetrieveFoodList, userLang))
		return
	}

	hasNextPage := len(userFoods) > UserFoodPageSize
	if hasNextPage {
		// Remove the extra item used only for checking existence of the next page.
		userFoods = userFoods[:UserFoodPageSize]
	}

	var messageText strings.Builder
	if len(userFoods) == 0 && offset == 0 {
		messageText.WriteString(i18n.GetText(i18n.KeyUserFoodNoFoodsAdded, userLang))
	} else if len(userFoods) == 0 && offset > 0 {
		// Should ideally not happen with correct button logic, but handle defensively
		messageText.WriteString(i18n.GetText(i18n.KeyUserFoodNoMoreFoods, userLang))
	} else {
		// Use GetTextf for the formatted header
		messageText.WriteString(i18n.GetTextf(i18n.KeyUserFoodListHeaderFormat, userLang, offset+1, offset+len(userFoods)))
		for i, food := range userFoods {
			messageText.WriteString(fmt.Sprintf("%d. %s\n", offset+i+1, food)) // Keep item format simple
		}
	}

	// Build keyboard
	keyboard := [][]tgmodels.InlineKeyboardButton{}
	navigationRow := []tgmodels.InlineKeyboardButton{}

	if offset > 0 {
		prevOffset := offset - UserFoodPageSize
		if prevOffset < 0 {
			prevOffset = 0 // Ensure offset doesn't go below 0
		}
		navigationRow = append(navigationRow, tgmodels.InlineKeyboardButton{
			Text:         i18n.GetText(i18n.KeyPreviousButton, userLang),
			CallbackData: fmt.Sprintf("%s%d", CallbackDataViewUserFoodOffset, prevOffset),
		})
	}

	if hasNextPage {
		nextOffset := offset + UserFoodPageSize
		navigationRow = append(navigationRow, tgmodels.InlineKeyboardButton{
			Text:         i18n.GetText(i18n.KeyNextButton, userLang),
			CallbackData: fmt.Sprintf("%s%d", CallbackDataViewUserFoodOffset, nextOffset),
		})
	}

	if len(navigationRow) > 0 {
		keyboard = append(keyboard, navigationRow)
	}

	// Add Back button
	keyboard = append(keyboard, []tgmodels.InlineKeyboardButton{
		{Text: i18n.GetText(i18n.KeyBackToOptionsButton, userLang), CallbackData: string(CallbackDataBackToOptions)}, // Use i18n
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
