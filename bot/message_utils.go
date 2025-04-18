package bot

import (
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/i18n" // Import the i18n package
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// InitialMenuMessageParams constructs the initial menu message parameters for a Telegram bot user.
// It generates a message with inline keyboard options based on the user's premium status.
// Premium users have additional options to set dietary restrictions, macros, body weight, and plan duration.
// All users can add and view favorite foods and generate meal plans.
// If the user is not premium, an upgrade option is displayed.
func InitialMenuMessageParams(chatID int64, user *database.User, numberOfCreatedPlans int64) *tgbot.SendMessageParams {
	var buttons [][]tgmodels.InlineKeyboardButton

	userLang := user.Language // Get user language

	if user.IsPremium {
		buttons = [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         i18n.GetText(i18n.KeySetDietaryRestrictionsButton, userLang),
					CallbackData: string(CallbackDataSetDietaryRestrictions),
				},
				{
					Text:         i18n.GetText(i18n.KeySetMacrosButton, userLang),
					CallbackData: string(CallbackDataAddMacroTargets),
				},
			},
			{
				{
					Text:         i18n.GetText(i18n.KeySetBodyWeightButton, userLang),
					CallbackData: string(CallbackDataSetBodyWeight),
				},
				{
					Text:         i18n.GetText(i18n.KeySetPlanDurationButton, userLang),
					CallbackData: string(CallbackDataAddPlanSchedule),
				},
			},
			{
				{
					Text:         i18n.GetText(i18n.KeyAddFavoriteFoodsButton, userLang),
					CallbackData: string(CallbackDataAddUserFood),
				},
				{
					Text:         i18n.GetText(i18n.KeyViewAddedFoodsButton, userLang),
					CallbackData: string(CallbackDataViewUserFoods),
				},
			},
			{
				{
					// Use this helper in this case
					Text:         getGenerateMealPlanMessageFromUser(user, numberOfCreatedPlans),
					CallbackData: string(CallbackDataCreateMealPlan),
				},
			},
			{
				{
					Text:         i18n.GetText(i18n.KeyGenerateWeekendPlanButton, userLang),
					CallbackData: string(CallbackDataCreateWeekendPlan),
				},
			},
			{
				{
					Text:         i18n.GetText(i18n.KeyChangeLanguageButton, userLang),
					CallbackData: string(CallbackDataChangeLanguage),
				},
			},
		}
	} else {
		buttons = [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         i18n.GetText(i18n.KeyAddFavoriteFoodsButton, userLang),
					CallbackData: string(CallbackDataAddUserFood),
				},
				{
					Text:         i18n.GetText(i18n.KeyViewAddedFoodsButton, userLang),
					CallbackData: string(CallbackDataViewUserFoods),
				},
			},
			{

				{
					Text:         i18n.GetText(i18n.KeySetDietaryRestrictionsButton, userLang),
					CallbackData: string(CallbackDataSetDietaryRestrictions),
				},
			},
			{
				{
					// Use this helper in this case
					Text:         getGenerateMealPlanMessageFromUser(user, numberOfCreatedPlans),
					CallbackData: string(CallbackDataCreateMealPlan),
				},
			},
			{
				{
					Text:         i18n.GetText(i18n.KeyChangeLanguageButton, userLang),
					CallbackData: string(CallbackDataChangeLanguage),
				},
			},
			// {
			// 	{
			// 		Text:         i18n.GetText(i18n.KeyUpgradePremiumButton, userLang),
			// 		CallbackData: string(CallbackDataMakePayment),
			// 	},
			// },
		}
	}

	return &tgbot.SendMessageParams{
		Text:   i18n.GetText(i18n.KeyInitialMenuPrompt, userLang),
		ChatID: chatID,
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: buttons, // Buttons are already translated
		},
	}
}

// getGenerateMealPlanMessageFromUser returns a string message for generating a meal plan.
// If the user is a premium member, it returns the translated "Generate meal plan".
// Otherwise, it returns a translated message indicating how many plans the user has generated out of the maximum.
func getGenerateMealPlanMessageFromUser(user *database.User, numberOfCreatedPlans int64) string {
	userLang := user.Language
	if user.IsPremium {
		return i18n.GetText(i18n.KeyGenerateMealPlanButtonPremium, userLang)
	}

	// This basically takes into consideration the maximum number of plans.
	// If for some reason someone goes back to a free version after being
	// premium they'd have more plans than allowed, normally. To prevent a situation
	// where the text shows (10/2), for example, we cap the number of plans
	// to not be greater than the maximum allowed
	numberOfCreatedPlansWithLimit := min(numberOfCreatedPlans, MAX_FREE_USERS_PLANS)

	// Use GetTextf for formatted string
	return i18n.GetTextf(i18n.KeyGenerateSamplePlanButtonFormat, userLang, numberOfCreatedPlansWithLimit, MAX_FREE_USERS_PLANS)
}
