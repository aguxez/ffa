package i18n

import (
	"fmt"

	"github.com/aguxez/ffa/database"
)

// Define keys for translatable strings
const (
	KeyMacroTargetsPrompt                     = "macro_targets_prompt"
	KeyMacroTargetsUpdated                    = "macro_targets_updated"
	KeyBackToOptionsButton                    = "back_to_options_button"
	KeyFeedbackBreakfastQuestion              = "feedback_breakfast_question"
	KeyFeedbackLunchQuestion                  = "feedback_lunch_question"
	KeyFeedbackDinnerQuestion                 = "feedback_dinner_question"
	KeyFeedbackEnjoyedButton                  = "feedback_enjoyed_button"
	KeyFeedbackDislikedButton                 = "feedback_disliked_button"
	KeyFeedbackThanks                         = "feedback_thanks"
	KeyErrorRetrieveFoodPrefs                 = "error_retrieve_food_prefs"
	KeyErrorRetrieveUserData                  = "error_retrieve_user_data"
	KeyMaxPlansReached                        = "max_plans_reached"
	KeyGeneratingPlan                         = "generating_plan"
	KeyErrorGetLastPlan                       = "error_get_last_plan"
	KeyErrorGeneratePlan                      = "error_generate_plan"
	KeyErrorGenerateDoc                       = "error_generate_doc"
	KeyPlanCaption                            = "plan_caption"
	KeyAcceptButton                           = "accept_button"
	KeyRejectButton                           = "reject_button"
	KeyErrorSendDoc                           = "error_send_doc"
	KeySetDietaryRestrictionsButton           = "set_dietary_restrictions_button"
	KeySetMacrosButton                        = "set_macros_button"
	KeySetBodyWeightButton                    = "set_body_weight_button"
	KeySetPlanDurationButton                  = "set_plan_duration_button"
	KeyAddFavoriteFoodsButton                 = "add_favorite_foods_button"
	KeyViewAddedFoodsButton                   = "view_added_foods_button"
	KeyGenerateMealPlanButtonPremium          = "generate_meal_plan_button_premium"
	KeyGenerateSamplePlanButtonFormat         = "generate_sample_plan_button_format" // Format string
	KeyGenerateWeekendPlanButton              = "generate_weekend_plan_button"
	KeyUpgradePremiumButton                   = "upgrade_premium_button"
	KeyInitialMenuPrompt                      = "initial_menu_prompt"
	KeyInvoiceTitle                           = "invoice_title"
	KeyInvoiceDescription                     = "invoice_description"
	KeyInvoicePriceLabel                      = "invoice_price_label"
	KeyErrorPaymentProcessing                 = "error_payment_processing"
	KeyPaymentSuccess                         = "payment_success"
	KeyPlanSavedConfirmation                  = "plan_saved_confirmation"
	KeyPlanScheduleSelectPrompt               = "plan_schedule_select_prompt"
	KeyPlanScheduleFiveDaysButton             = "plan_schedule_five_days_button"
	KeyPlanScheduleThreeDaysButton            = "plan_schedule_three_days_button"
	KeyPlanScheduleUpdated                    = "plan_schedule_updated"
	KeyBodyWeightPrompt                       = "body_weight_prompt"
	KeyBodyWeightInvalidInput                 = "body_weight_invalid_input"
	KeyBodyWeightSaved                        = "body_weight_saved"
	KeyDietaryRestrictionsPrompt              = "dietary_restrictions_prompt"
	KeyDietaryRestrictionsUpdated             = "dietary_restrictions_updated"
	KeyUserFoodAddPrompt                      = "user_food_add_prompt"
	KeyErrorParseFoodPrefs                    = "error_parse_food_prefs"
	KeyErrorSaveFoodPrefs                     = "error_save_food_prefs"
	KeyUserFoodUpdated                        = "user_food_updated"
	KeyInvalidPaginationRequest               = "invalid_pagination_request"
	KeyErrorRetrieveFoodList                  = "error_retrieve_food_list"
	KeyUserFoodNoFoodsAdded                   = "user_food_no_foods_added"
	KeyUserFoodNoMoreFoods                    = "user_food_no_more_foods"
	KeyUserFoodListHeaderFormat               = "user_food_list_header_format" // Format string
	KeyPreviousButton                         = "previous_button"
	KeyNextButton                             = "next_button"
	KeyErrorPremiumCheck                      = "error_premium_check"
	KeyPremiumRequired                        = "premium_required"
	KeyFeedbackStartButton                    = "feedback_start_button"
	KeyErrorResponseGeneric                   = "error_response_generic" // Generic error for ErrorResponse func
	KeyMealPlanDocumentBreakfast              = "meal_plan_document_breakfast"
	KeyMealPlanDocumentLunch                  = "meal_plan_document_lunch"
	KeyMealPlanDocumentDinner                 = "meal_plan_document_dinner"
	KeyMealPlanDocumentSnack                  = "meal_plan_document_snack"
	KeyMealPlanDocumentPreparation            = "meal_plan_document_preparation"
	KeyMealPlanDocumentShoppingList           = "meal_plan_document_shopping_list"
	KeyMealPlanDocumentTableHeaderFood        = "meal_plan_document_table_header_food"
	KeyMealPlanDocumentTableHeaderMacros      = "meal_plan_document_table_header_macro"
	KeyMealPlanDocumentTableHeaderExplanation = "meal_plan_document_table_header_explanation"
	KeyChangeLanguageButton                   = "change_language_button"
	KeyChangeLanguagePrompt                   = "change_language_prompt"
	KeyLanguageChanged                        = "language_changed"
	KeySupportText                            = "support_text"
)

var translations = map[string]map[database.UserLanguage]string{
	KeyMacroTargetsPrompt: {
		database.UserLanguageEnglish: `Enter your macro targets in any format:
- carbs: 300g, protein: 150g, fat: 50g
- calories: 2500

You can specify just some macros (e.g., only calories). All plans will include estimates for all nutritional values.`,
		database.UserLanguageSpanish: `Introduce tus objetivos de macros en cualquier formato:
- carbohidratos: 300g, proteína: 150g, grasa: 50g
- calorías: 2500

También puedes especificar solo algunos macros (p. ej., solo calorías). Todos los planes incluirán estimaciones para todos los valores nutricionales.`,
	},
	KeyMacroTargetsUpdated: {
		database.UserLanguageEnglish: "Amazing. Your macro targets have been updated!",
		database.UserLanguageSpanish: "¡Genial! Tus objetivos de macros han sido actualizados.",
	},
	KeyBackToOptionsButton: {
		database.UserLanguageEnglish: "↼ Back to options",
		database.UserLanguageSpanish: "↼ Volver a opciones",
	},
	KeyFeedbackBreakfastQuestion: {
		database.UserLanguageEnglish: "How do you feel about the breakfast? (1/3)",
		database.UserLanguageSpanish: "¿Qué te pareció el desayuno? (1/3)",
	},
	KeyFeedbackLunchQuestion: {
		database.UserLanguageEnglish: "How do you feel about the lunch? (2/3)",
		database.UserLanguageSpanish: "¿Qué te pareció el almuerzo? (2/3)",
	},
	KeyFeedbackDinnerQuestion: {
		database.UserLanguageEnglish: "How do you feel about the dinner? (3/3)",
		database.UserLanguageSpanish: "¿Qué te pareció la cena? (3/3)",
	},
	KeyFeedbackEnjoyedButton: {
		database.UserLanguageEnglish: "Enjoyed❤️",
		database.UserLanguageSpanish: "Me gustó❤️",
	},
	KeyFeedbackDislikedButton: {
		database.UserLanguageEnglish: "Disliked☹️",
		database.UserLanguageSpanish: "No me gustó☹️",
	},
	KeyFeedbackThanks: {
		database.UserLanguageEnglish: "Thanks for your feedback. Your recommendations will be updated for your next meal plan.",
		database.UserLanguageSpanish: "Gracias por tus comentarios. Tus recomendaciones se actualizarán para tu próximo plan de comidas.",
	},
	KeyErrorRetrieveFoodPrefs: {
		database.UserLanguageEnglish: "Failed to retrieve your food preferences.",
		database.UserLanguageSpanish: "Error al recuperar tus preferencias de comida.",
	},
	KeyErrorRetrieveUserData: {
		database.UserLanguageEnglish: "Failed to retrieve user data.",
		database.UserLanguageSpanish: "Error al recuperar los datos del usuario.",
	},
	KeyMaxPlansReached: {
		database.UserLanguageEnglish: "You have reached the maximum number of meal plans. You can upgrade to enjoy all features!",
		database.UserLanguageSpanish: "Has alcanzado el número máximo de planes de comidas. ¡Puedes comprar RawPlan Premium para disfrutar de todas las funciones!",
	},
	KeyGeneratingPlan: {
		database.UserLanguageEnglish: "I'm generating a plan for you. This can take a minute.",
		database.UserLanguageSpanish: "Estoy generando un plan para ti. Esto puede tardar un minuto.",
	},
	KeyErrorGetLastPlan: {
		database.UserLanguageEnglish: "There was an error getting your last accepted meal plan. This has been informed.",
		database.UserLanguageSpanish: "Hubo un error al obtener tu último plan de comidas aceptado. Esto ha sido informado.",
	},
	KeyErrorGeneratePlan: {
		database.UserLanguageEnglish: "There was an error generating the meal plan. This has been informed.",
		database.UserLanguageSpanish: "Hubo un error al generar el plan de comidas. Esto ha sido informado.",
	},
	KeyErrorGenerateDoc: {
		database.UserLanguageEnglish: "There was an error generating your document. This has been informed.",
		database.UserLanguageSpanish: "Hubo un error al generar tu documento. Esto ha sido informado.",
	},
	KeyPlanCaption: {
		database.UserLanguageEnglish: "Here is our recommended meal plan based on the schedule and information you have shared with us!",
		database.UserLanguageSpanish: "¡Aquí tienes nuestro plan de comidas recomendado basado en el horario y la información que has compartido con nosotros!",
	},
	KeyAcceptButton: {
		database.UserLanguageEnglish: "Accept ✅",
		database.UserLanguageSpanish: "Aceptar ✅",
	},
	KeyRejectButton: {
		database.UserLanguageEnglish: "Reject ❌",
		database.UserLanguageSpanish: "Rechazar ❌",
	},
	KeyErrorSendDoc: {
		database.UserLanguageEnglish: "There was an error sending your document. This has been informed.",
		database.UserLanguageSpanish: "Hubo un error al enviar tu documento. Esto ha sido informado.",
	},
	KeySetDietaryRestrictionsButton: {
		database.UserLanguageEnglish: "Dietary restrictions",
		database.UserLanguageSpanish: "Restricciones dietéticas",
	},
	KeySetMacrosButton: {
		database.UserLanguageEnglish: "Set Macros",
		database.UserLanguageSpanish: "Establecer Macros",
	},
	KeySetBodyWeightButton: {
		database.UserLanguageEnglish: "Set body weight",
		database.UserLanguageSpanish: "Peso corporal",
	},
	KeySetPlanDurationButton: {
		database.UserLanguageEnglish: "Set duration of plans",
		database.UserLanguageSpanish: "Duración de los planes",
	},
	KeyAddFavoriteFoodsButton: {
		database.UserLanguageEnglish: "Add favorite foods",
		database.UserLanguageSpanish: "Añadir comidas favoritas",
	},
	KeyViewAddedFoodsButton: {
		database.UserLanguageEnglish: "View added foods",
		database.UserLanguageSpanish: "Ver comidas añadidas",
	},
	KeyGenerateMealPlanButtonPremium: {
		database.UserLanguageEnglish: "Generate meal plan",
		database.UserLanguageSpanish: "Generar plan",
	},
	KeyGenerateSamplePlanButtonFormat: {
		database.UserLanguageEnglish: "Generate week plan (%v/%v)",
		database.UserLanguageSpanish: "Generar plan semanal (%v/%v)",
	},
	KeyGenerateWeekendPlanButton: {
		database.UserLanguageEnglish: "Generate weekend plan",
		database.UserLanguageSpanish: "Generar plan de fin de semana",
	},
	KeyUpgradePremiumButton: {
		database.UserLanguageEnglish: "⭐Upgrade to premium⭐",
		database.UserLanguageSpanish: "⭐Vuélvete premium⭐",
	},
	KeyInitialMenuPrompt: {
		database.UserLanguageEnglish: "Create personalized meal plans based on certain preferences.\n\nSelect an option to build your plan:",
		database.UserLanguageSpanish: "Crea planes de comidas personalizados basados en tus preferencias.\n\nSelecciona una opción para construir tu plan:",
	},
	KeyInvoiceTitle: {
		database.UserLanguageEnglish: "RawPlan Premium",
		database.UserLanguageSpanish: "RawPlan Premium",
	},
	KeyInvoiceDescription: {
		database.UserLanguageEnglish: `Access to all features of RawPlan and future developments.
Get a higher limit number of meal plans per week and timed follow-ups.
This is a non-recurring subscription to get all features of the product!`,
		database.UserLanguageSpanish: `Acceso a todas las funciones de RawPlan y desarrollos futuros.
Obtén un límite mayor de planes de comidas por semana y seguimientos programados.
¡Esta es una suscripción no recurrente para obtener todas las funciones del producto!`,
	},
	KeyInvoicePriceLabel: {
		database.UserLanguageEnglish: "Premium test",
		database.UserLanguageSpanish: "Prueba Premium",
	},
	KeyErrorPaymentProcessing: {
		database.UserLanguageEnglish: "There was an error processing your successful payment. This has been informed.",
		database.UserLanguageSpanish: "Hubo un error al procesar tu pago exitoso. Esto ha sido informado.",
	},
	KeyPaymentSuccess: {
		database.UserLanguageEnglish: "You have successfully upgraded to RawPlan Premium!",
		database.UserLanguageSpanish: "¡Has actualizado exitosamente a RawPlan Premium!",
	},
	KeyPlanSavedConfirmation: {
		database.UserLanguageEnglish: "Your plan has been saved and will be used for further feedback and check-ins.",
		database.UserLanguageSpanish: "Tu plan ha sido guardado y se utilizará para futuros comentarios y seguimientos.",
	},
	KeyPlanScheduleSelectPrompt: {
		database.UserLanguageEnglish: "Select how often your plans should be for.",
		database.UserLanguageSpanish: "Selecciona con qué frecuencia deben ser tus planes.",
	},
	KeyPlanScheduleFiveDaysButton: {
		database.UserLanguageEnglish: "Five days",
		database.UserLanguageSpanish: "Cinco días",
	},
	KeyPlanScheduleThreeDaysButton: {
		database.UserLanguageEnglish: "Every 3 days",
		database.UserLanguageSpanish: "Cada 3 días",
	},
	KeyPlanScheduleUpdated: {
		database.UserLanguageEnglish: "Successfully updated your plan schedule!",
		database.UserLanguageSpanish: "¡Horario de tu plan actualizado con éxito!",
	},
	KeyBodyWeightPrompt: {
		database.UserLanguageEnglish: "Tell me your body weight in kilograms.",
		database.UserLanguageSpanish: "Dime tu peso corporal en kilogramos.",
	},
	KeyBodyWeightInvalidInput: {
		database.UserLanguageEnglish: "Please enter a valid number for your body weight.",
		database.UserLanguageSpanish: "Por favor, introduce un número válido para tu peso corporal.",
	},
	KeyBodyWeightSaved: {
		database.UserLanguageEnglish: "Your body weight has been saved!",
		database.UserLanguageSpanish: "¡Tu peso corporal ha sido guardado!",
	},
	KeyDietaryRestrictionsPrompt: {
		database.UserLanguageEnglish: "Tell me any food allergies, intolerances, or dietary preferences (vegan, vegetarian, gluten-free, etc.). Your meal plans will automatically exclude restricted ingredients.",
		database.UserLanguageSpanish: "Indícame cualquier alergia alimentaria, intolerancia o preferencia dietética (vegano, vegetariano, sin gluten, etc.). Tus planes de comidas excluirán automáticamente los ingredientes restringidos.",
	},
	KeyDietaryRestrictionsUpdated: {
		database.UserLanguageEnglish: "Amazing. Your dietary restrictions have been updated!",
		database.UserLanguageSpanish: "¡Genial! Tus restricciones dietéticas han sido actualizadas.",
	},
	KeyUserFoodAddPrompt: {
		database.UserLanguageEnglish: "Share foods you enjoy to include in your meal plans. Any items you add will be combined with your existing saved favorites.",
		database.UserLanguageSpanish: "Comparte comidas que disfrutes para incluirlas en tus planes. Cualquier elemento que añadas se combinará con tus favoritos ya guardados.",
	},
	KeyErrorParseFoodPrefs: {
		database.UserLanguageEnglish: "There was an error parsing your food preferences. Please try again.",
		database.UserLanguageSpanish: "Hubo un error al procesar tus preferencias de comida. Por favor, inténtalo de nuevo.",
	},
	KeyErrorSaveFoodPrefs: {
		database.UserLanguageEnglish: "There was an error saving your food preferences. Please try again.",
		database.UserLanguageSpanish: "Hubo un error al guardar tus preferencias de comida. Por favor, inténtalo de nuevo.",
	},
	KeyUserFoodUpdated: {
		database.UserLanguageEnglish: "Successfully updated your food preferences!",
		database.UserLanguageSpanish: "¡Preferencias de comida actualizadas con éxito!",
	},
	KeyInvalidPaginationRequest: {
		database.UserLanguageEnglish: "Invalid pagination request.",
		database.UserLanguageSpanish: "Solicitud de paginación inválida.",
	},
	KeyErrorRetrieveFoodList: {
		database.UserLanguageEnglish: "Error retrieving your food list.",
		database.UserLanguageSpanish: "Error al recuperar tu lista de comidas.",
	},
	KeyUserFoodNoFoodsAdded: {
		database.UserLanguageEnglish: "You haven't added any favorite foods yet.",
		database.UserLanguageSpanish: "Aún no has añadido ninguna comida favorita.",
	},
	KeyUserFoodNoMoreFoods: {
		database.UserLanguageEnglish: "No more foods found.",
		database.UserLanguageSpanish: "No se encontraron más comidas.",
	},
	KeyUserFoodListHeaderFormat: {
		database.UserLanguageEnglish: "Your Foods (%d-%d):\n",
		database.UserLanguageSpanish: "Tus Comidas (%d-%d):\n",
	},
	KeyPreviousButton: {
		database.UserLanguageEnglish: "⬅️ Previous",
		database.UserLanguageSpanish: "⬅️ Anterior",
	},
	KeyNextButton: {
		database.UserLanguageEnglish: "Next ➡️",
		database.UserLanguageSpanish: "Siguiente ➡️",
	},
	KeyErrorPremiumCheck: {
		database.UserLanguageEnglish: "Sorry, there was an error checking your premium status.",
		database.UserLanguageSpanish: "Lo sentimos, hubo un error al verificar tu estado premium.",
	},
	KeyPremiumRequired: {
		database.UserLanguageEnglish: "This feature requires a premium subscription.",
		database.UserLanguageSpanish: "Esta función requiere una suscripción premium.",
	},
	KeyFeedbackStartButton: {
		database.UserLanguageEnglish: "Start ✅",
		database.UserLanguageSpanish: "Empezar ✅",
	},
	KeyErrorResponseGeneric: {
		database.UserLanguageEnglish: "An error occurred. Please try again later.",
		database.UserLanguageSpanish: "Ocurrió un error. Por favor, inténtalo de nuevo.",
	},
	KeyMealPlanDocumentBreakfast: {
		database.UserLanguageEnglish: "Breakfast",
		database.UserLanguageSpanish: "Desayuno",
	},
	KeyMealPlanDocumentLunch: {
		database.UserLanguageEnglish: "Lunch",
		database.UserLanguageSpanish: "Almuerzo",
	},
	KeyMealPlanDocumentDinner: {
		database.UserLanguageEnglish: "Dinner",
		database.UserLanguageSpanish: "Cena",
	},
	KeyMealPlanDocumentSnack: {
		database.UserLanguageEnglish: "Snack",
		database.UserLanguageSpanish: "Snack",
	},
	KeyMealPlanDocumentPreparation: {
		database.UserLanguageEnglish: "Preparation",
		database.UserLanguageSpanish: "Preparación",
	},
	KeyMealPlanDocumentShoppingList: {
		database.UserLanguageEnglish: "Shopping List",
		database.UserLanguageSpanish: "Lista de compras",
	},
	KeyMealPlanDocumentTableHeaderFood: {
		database.UserLanguageEnglish: "Food",
		database.UserLanguageSpanish: "Comida",
	},
	KeyMealPlanDocumentTableHeaderMacros: {
		database.UserLanguageEnglish: "Macros",
		database.UserLanguageSpanish: "Macros",
	},
	KeyMealPlanDocumentTableHeaderExplanation: {
		database.UserLanguageEnglish: "Explanation",
		database.UserLanguageSpanish: "Explicación",
	},
	KeyChangeLanguageButton: {
		database.UserLanguageEnglish: "Change language 🇬🇧🇪🇸",
		database.UserLanguageSpanish: "Cambiar idioma 🇬🇧🇪🇸",
	},
	KeyChangeLanguagePrompt: {
		database.UserLanguageEnglish: "Select your preferred language.",
		database.UserLanguageSpanish: "Selecciona tu idioma preferido.",
	},
	KeyLanguageChanged: {
		database.UserLanguageEnglish: "Your language preference has been updated",
		database.UserLanguageSpanish: "Tu preferencia de idioma ha sido actualizada",
	},
	KeySupportText: {
		database.UserLanguageEnglish: "If you have any questions or need assistance, please reach out to our support team at support@rawplanbot.com. We're here to help!",
		database.UserLanguageSpanish: "Si tienes alguna pregunta o necesitas ayuda, ponte en contacto con nuestro equipo de soporte en support@rawplanbot.com. Estamos aqui para ayudarte!",
	},
}

// GetText retrieves the translated string for a given key and language.
// It defaults to English if the specified language or key is not found.
func GetText(key string, lang database.UserLanguage) string {
	if langMap, ok := translations[key]; ok {
		if text, ok := langMap[lang]; ok {
			return text
		}
		// Fallback to English if specific language translation is missing
		if text, ok := langMap[database.UserLanguageEnglish]; ok {
			return text
		}
	}
	// Fallback if key is entirely missing (should not happen ideally)
	fmt.Printf("Warning: Translation key '%s' not found\n", key)
	return key // Return the key itself as a last resort
}

// GetTextf retrieves a formatted translated string.
func GetTextf(key string, lang database.UserLanguage, args ...interface{}) string {
	formatString := GetText(key, lang)
	return fmt.Sprintf(formatString, args...)
}
