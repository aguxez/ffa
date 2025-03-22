package bot

import (
	"context"

	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MacroTargetsHandler handles the macro targets button
type MacroTargetsHandler struct {
	deps *HandlerDependencies
}

// Handle handles the macro targets button
func NewMacroTargetsHandler(deps *HandlerDependencies) *MacroTargetsHandler {
	return &MacroTargetsHandler{
		deps: deps,
	}
}

// Handle handles the macro targets button.
func (h *MacroTargetsHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "MacroTargetsHandler").Msg("Handling macro targets button")

	// Update user state
	h.deps.StateManager.SetUserState(uint(update.CallbackQuery.From.ID), models.ConversationStageAddMacroTargets)

	// Edit the callback text
	_, err := bot.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: `Enter your macro targets in any format:
- carbs: 300g, protein: 150g, fat: 50g
- calories: 2500

You can specify just some macros (e.g., only calories). All plans will include estimates for all nutritional values.
`,
	})
	if err != nil {
		log.Err(err).Msg("Error sending macro targets message")
		return
	}
}

// MacroTargetsMessageHandler handles the macro targets message after the callback
type MacroTargetsMessageHandler struct {
	deps *HandlerDependencies
}

// NewMacroTargetsMessageHandler creates a new MacroTargetsMessageHandler
func NewMacroTargetsMessageHandler(deps *HandlerDependencies) *MacroTargetsMessageHandler {
	return &MacroTargetsMessageHandler{
		deps: deps,
	}
}

// Handle handles the macro targets message after the callback
func (h *MacroTargetsMessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "MacroTargetsMessageHandler").Msg("Handling macro targets message")

	// Update user's state to the beginning
	h.deps.StateManager.SetUserState(uint(update.Message.From.ID), models.ConversationStageNull)

	// Parse the message received by the user
	macroTargets, err := h.deps.NaturalLanguageAgent.ParseMacroTargetsMessage(ctx, update.Message.Text)
	if err != nil {
		log.Err(err).Msg("Error parsing macro targets message")
		return
	}

	// Update user's macro targets
	err = database.UpdateUserMacros(h.deps.DB, uint(update.Message.From.ID), macroTargets)
	if err != nil {
		log.Err(err).Msg("Error updating user's macro targets")
		return
	}

	// Send message informing about the update
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Amazing. Your macro targets have been updated!",
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
	if err != nil {
		log.Err(err).Msg("Error sending update message")
		return
	}
}
