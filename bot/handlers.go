package bot

import (
	"context"

	"github.com/aguxez/ffa/agent"
	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

// CallbackData defines type-safe callback data values
type CallbackData string

const (
	CallbackDataCreateMealPlan        CallbackData = "callback_create_meal_plan"
	CallbackDataAddDietaryPreference  CallbackData = "callback_add_dietary_preference"
	CallbackDataAddUserFood           CallbackData = "callback_add_user_food"
	CallbackDataAddPlanSchedule       CallbackData = "callback_add_plan_schedule"
	CallbackDataPlanScheduleWeekly    CallbackData = "callback_plan_schedule_weekly"
	CallbackDataPlanScheduleThreeDays CallbackData = "callback_plan_schedule_three_days"
	CallbackDataAddMacroTargets       CallbackData = "callback_add_macro_targets"
	CallbackDataBackToOptions         CallbackData = "callback_back_to_options"
)

// Handler is the interface for all handlers
type Handler interface {
	Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update)
}

// HandlerDependencies contains all dependencies needed by handlers
type HandlerDependencies struct {
	NutritionAgent       *agent.NutritionAgent
	NaturalLanguageAgent *agent.NaturalLanguageAgent
	DB                   *gorm.DB
	StateManager         *models.StateManager
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
func (r *HandlerRegistry) Register(handlerType, pattern string, matchType tgbot.MatchType, handler Handler) {
	key := handlerType + ":" + pattern
	r.handlers[key] = handler
}

// Get returns a handler
func (r *HandlerRegistry) Get(handlerType, pattern string) Handler {
	key := handlerType + ":" + pattern
	return r.handlers[key]
}

// ErrorResponse sends an error response to the user
func ErrorResponse(ctx context.Context, bot *tgbot.Bot, chatID int64, msg string) {
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   msg,
	})
}
