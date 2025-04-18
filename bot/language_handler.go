package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// SettingsHandler handles settings for a user
type SettingsHandler struct {
	deps *HandlerDependencies
}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler(deps *HandlerDependencies) *SettingsHandler {
	return &SettingsHandler{
		deps: deps,
	}
}

// Handle handles the settings command (to set languages, etc.)
func (h *SettingsHandler) Handle(
	ctx context.Context,
	bot *tgbot.Bot,
	update *tgmodels.Update,
) {
	logger := log.
		With().
		Str("handler", "SettingsHandler").
		Logger()

	logger.Debug().Msg("Handling settings")

	// Get user language
	user, err := h.deps.DB.GetUser(uint(update.CallbackQuery.From.ID))
	if err != nil {
		logger.Err(err).Msg("Error getting user")

		ErrorResponse(
			ctx,
			bot,
			update.CallbackQuery.Message.Message.Chat.ID,
			i18n.GetText(i18n.KeyErrorResponseGeneric, "en"),
		)

		return
	}
	userLang := user.Language

	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   i18n.GetText(i18n.KeyChangeLanguagePrompt, userLang),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "🇬🇧 English",
						CallbackData: string(CallbackDataChangeLanguageEn),
					},
					{
						Text:         "🇪🇸 Español",
						CallbackData: string(CallbackDataChangeLanguageEs),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending message")
	}
}

type LanguageHandler struct {
	deps     *HandlerDependencies
	language CallbackData
}

// NewLanguageHandler creates a new LanguageHandler
func NewLanguageHandler(deps *HandlerDependencies, language CallbackData) *LanguageHandler {
	return &LanguageHandler{
		deps:     deps,
		language: language,
	}
}

func (h *LanguageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	logger := log.
		With().
		Str("handler", "LanguageHandler").
		Int64("userID", update.CallbackQuery.From.ID).
		Logger()

	logger.Debug().Msg("Handling language change")

	// Update user language on DB
	var userLangInput database.UserLanguage

	switch h.language {
	case CallbackDataChangeLanguageEn:
		userLangInput = database.UserLanguageEnglish
	case CallbackDataChangeLanguageEs:
		userLangInput = database.UserLanguageSpanish
	}

	err := h.deps.DB.UpdateUserLanguage(update.CallbackQuery.From.ID, userLangInput)
	if err != nil {
		logger.Err(err).Msg("Error updating user language")
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   i18n.GetText(i18n.KeyLanguageChanged, userLangInput),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         i18n.GetText(i18n.KeyBackToOptionsButton, userLangInput),
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending update message")
		// No need to send another error message here, just log
		return
	}
}
