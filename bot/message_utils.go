package bot

import (
	"github.com/aguxez/ffa/database"
	"github.com/go-telegram/bot"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// InitialMenuMessageParams returns a SendMessageParams object that sends an initial
// welcome message with inline keyboard options to the chat.
func InitialMenuMessageParams(chatID int64, user *database.User) *bot.SendMessageParams {
	var buttons [][]tgmodels.InlineKeyboardButton

	if user.IsPremium {
		buttons = [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         "Set dietary restrictions",
					CallbackData: string(CallbackDataSetDietaryRestrictions),
				},
				{
					Text:         "Set Macros",
					CallbackData: string(CallbackDataAddMacroTargets),
				},
			},
			{
				{
					Text:         "Set body weight",
					CallbackData: string(CallbackDataSetBodyWeight),
				},
				{
					Text:         "Set duration of plans",
					CallbackData: string(CallbackDataAddPlanSchedule),
				},
			},
			{
				{
					Text:         "Add favorite foods",
					CallbackData: string(CallbackDataAddUserFood),
				},
				{
					Text:         "View added foods",
					CallbackData: string(CallbackDataViewUserFoods),
				},
			},
			{
				{
					Text:         "Generate sample plan",
					CallbackData: string(CallbackDataCreateMealPlan),
				},
			},
		}
	} else {
		buttons = [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         "Add favorite foods",
					CallbackData: string(CallbackDataAddUserFood),
				},
				{
					Text:         "View added foods",
					CallbackData: string(CallbackDataViewUserFoods),
				},
			},
			{
				{
					Text:         "Generate sample plan",
					CallbackData: string(CallbackDataCreateMealPlan),
				},
			},
			{
				{
					Text:         "⭐Upgrade to premium⭐",
					CallbackData: string(CallbackDataMakePayment),
				},
			},
		}
	}

	return &tgbot.SendMessageParams{
		Text:   "Create personalized meal plans based on your preferences.\n\nSelect an option to build your plan:",
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}
}
