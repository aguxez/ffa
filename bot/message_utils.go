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
		Text:   "Create personalized meal plans based on your preferences.\n\nSelect an option to build your plan:",
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         "Set Macros",
						CallbackData: string(CallbackDataAddMacroTargets),
					},
					{
						Text:         "Set dietary restrictions",
						CallbackData: string(CallbackDataSetDietaryRestrictions),
					},
				},
				{
					{
						Text:         "Add favorite foods",
						CallbackData: string(CallbackDataAddUserFood),
					},
					{
						Text:         "Set meal times",
						CallbackData: string(CallbackDataAddPlanSchedule),
					},
				},
				{
					{
						Text:         "Generate sample plan",
						CallbackData: string(CallbackDataCreateMealPlan),
					},
				},
			},
		},
	}
}
