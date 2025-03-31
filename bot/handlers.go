package bot

import (
	"context"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/database"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// CallbackData defines type-safe callback data values
type CallbackData string

const (
	CallbackDataCreateMealPlan         CallbackData = "callback_create_meal_plan"
	CallbackDataSetDietaryRestrictions CallbackData = "callback_set_dietary_restrictions"
	CallbackDataAddUserFood            CallbackData = "callback_add_user_food"
	CallbackDataAddPlanSchedule        CallbackData = "callback_add_plan_schedule"
	CallbackDataPlanScheduleWeekly     CallbackData = "callback_plan_schedule_weekly"
	CallbackDataPlanScheduleThreeDays  CallbackData = "callback_plan_schedule_three_days"
	CallbackDataAddMacroTargets        CallbackData = "callback_add_macro_targets"
	CallbackDataBackToOptions          CallbackData = "callback_back_to_options"
	CallbackDataBackToOptionsNoDelete  CallbackData = "callback_back_to_options_no_delete"
	CallbackDataSetBodyWeight          CallbackData = "callback_set_body_weight"

	// Callback to accept or deny a plan whenever we send it
	CallBackDataAcceptPlan CallbackData = "callback_accept_plan"
	CallbackDataRejectPlan CallbackData = "callback_reject_plan"

	// Callback whenever we start asking for feedback on meal plans
	CallbackDataMealPlanFeedbackStart CallbackData = "callback_meal_plan_feedback_start"

	CallbackDataMealPlanFeedbackBreakfastGood CallbackData = "callback_meal_plan_feedback_breakfast_good"
	CallbackDataMealPlanFeedbackBreakfastBad  CallbackData = "callback_meal_plan_feedback_breakfast_bad"

	CallbackDataMealPlanFeedbackLunchGood CallbackData = "callback_meal_plan_feedback_lunch_good"
	CallbackDataMealPlanFeedbackLunchBad  CallbackData = "callback_meal_plan_feedback_lunch_bad"

	CallbackDataMealPlanFeedbackDinnerGood CallbackData = "callback_meal_plan_feedback_dinner_good"
	CallbackDataMealPlanFeedbackDinnerBad  CallbackData = "callback_meal_plan_feedback_dinner_bad"
)

// Handler is the interface for all handlers
type Handler interface {
	Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update)
}

// HandlerDependencies contains all dependencies needed by handlers
type HandlerDependencies struct {
	NutritionAgent       *agent.NutritionAgent
	NaturalLanguageAgent *agent.NaturalLanguageAgent
	DB                   *database.Database
	StateManager         *models.StateManager
	HandlerRegistry      *HandlerRegistry
}

// HandlerRegistry manages the registration of handlers
type HandlerRegistry struct {
	handlers map[string]Handler
}

// NewHandlerRegistry creates a new HandlerRegistry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]Handler),
	}
}

// Register registers a handler
func (r *HandlerRegistry) Register(key string, handler Handler) {
	r.handlers[key] = handler
}

// Get returns a handler
func (r *HandlerRegistry) Get(key string) Handler {
	return r.handlers[key]
}

// ErrorResponse sends an error response to the user
func ErrorResponse(ctx context.Context, bot *tgbot.Bot, chatID int64, msg string) {
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   msg,
	})
}
