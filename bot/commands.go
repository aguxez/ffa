package bot

type CallbackData string

const (
	CallbackDataCreateMealPlan        CallbackData = "create_meal_plan"
	CallbackDataAddUserFood           CallbackData = "add_user_food"
	CallbackDataAddPlanSchedule       CallbackData = "add_plan_schedule"
	CallbackDataAddDietaryPreference  CallbackData = "add_dietary_preference"
	CallbackDataPlanScheduleWeekly    CallbackData = "plan_schedule_weekly"
	CallbackDataPlanScheduleThreeDays CallbackData = "plan_schedule_3_days"
)
