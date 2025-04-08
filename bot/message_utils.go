package bot

import (
	"fmt"

	"github.com/aguxez/ffa/database"
	"github.com/go-telegram/bot"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// InitialMenuMessageParams constructs the initial menu message parameters for a Telegram bot user.
// It generates a message with inline keyboard options based on the user's premium status.
// Premium users have additional options to set dietary restrictions, macros, body weight, and plan duration.
// All users can add and view favorite foods and generate meal plans.
// If the user is not premium, an upgrade option is displayed.
func InitialMenuMessageParams(chatID int64, user *database.User, numberOfCreatedPlans int64) *bot.SendMessageParams {
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
					Text:         getGenerateMealPlanMessageFromUser(user, numberOfCreatedPlans),
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
					Text:         getGenerateMealPlanMessageFromUser(user, numberOfCreatedPlans),
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

// getGenerateMealPlanMessageFromUser returns a string message for generating a meal plan.
// If the user is a premium member, it returns "Generate meal plan".
// Otherwise, it returns a message indicating how many plans the user has generated out of a maximum of 3.
func getGenerateMealPlanMessageFromUser(user *database.User, numberOfCreatedPlans int64) string {
	if user.IsPremium {
		return "Generate meal plan"
	}

	return fmt.Sprintf("Generate sample plan (%v/%v)", numberOfCreatedPlans, MAX_FREE_USERS_PLANS)
}
