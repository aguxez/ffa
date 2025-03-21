package bot

import (
	"github.com/go-telegram/bot"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// InitialMenuMessageParams returns a SendMessageParams object that sends an initial
// welcome message with inline keyboard options to the chat.
func InitialMenuMessageParams(chatID int64) *bot.SendMessageParams {
	return &tgbot.SendMessageParams{
		Text:   "I can help you create meal plans tailored for your dietary preferences.\n\nHere are some things you can do for your plans.",
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Add Macro Targets",
						CallbackData: string(CallbackDataAddMacroTargets),
					},
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
				{
					{
						Text:         "Create test plan",
						CallbackData: string(CallbackDataCreateMealPlan),
					},
				},
			},
		},
	}
}
