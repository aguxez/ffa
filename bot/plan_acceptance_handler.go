package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// PlanAcceptanceHandler handles the back to options button
type PlanAcceptanceHandler struct {
	deps           *HandlerDependencies
	isPlanAccepted bool
}

// NewPlanAcceptanceHandler creates a new PlanAcceptanceHandler
func NewPlanAcceptanceHandler(deps *HandlerDependencies, isPlanAccepted bool) *PlanAcceptanceHandler {
	return &PlanAcceptanceHandler{
		deps:           deps,
		isPlanAccepted: isPlanAccepted,
	}
}

// Handle handles the event whenever a user accepts or rejects a plan
func (h *PlanAcceptanceHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "PlanAcceptanceHandler").Msg("Handling plan acceptance")

	userID := uint(update.CallbackQuery.From.ID)

	// Reset user state
	h.deps.StateManager.SetUserState(userID, models.ConversationStageNull)

	// If the user approved the plan, we want to save it in the database for further
	// feedback down the line
	if h.isPlanAccepted {
		// TODO: Save plan to database
		return
	}

	// Create a new plan for the user
	// TODO: Limit plan generation from this handler
	h.deps.HandlerRegistry.Get(string(CallbackDataCreateMealPlan)).Handle(ctx, bot, update)
}
