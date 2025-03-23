// bot/message_handler.go
package bot

import (
	"context"

	"github.com/aguxez/ffa/models"
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles all text messages
type MessageHandler struct {
	deps     *HandlerDependencies
	handlers map[models.ConversationStage]Handler
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(deps *HandlerDependencies) *MessageHandler {
	h := &MessageHandler{
		deps:     deps,
		handlers: make(map[models.ConversationStage]Handler),
	}

	h.handlers[models.ConversationStageAddUserFood] = NewUserFoodMessageHandler(deps)
	h.handlers[models.ConversationStageAddMacroTargets] = NewMacroTargetsMessageHandler(deps)
	h.handlers[models.ConversationStageSetDietaryRestrictions] = NewSetDietaryRestrictionsMessageHandler(deps)
	h.handlers[models.ConversationStageSetBodyWeight] = NewSetBodyWeightMessageHandler(deps)

	return h
}

// Handle handles incoming text messages
func (h *MessageHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	log.Debug().Str("handler", "MessageHandler").Msg("Handling text message")

	// Guard against nil message or From field
	if update.Message == nil || update.Message.From == nil {
		log.Debug().Msg("Message or From field is nil")
		return
	}

	userID := uint(update.Message.From.ID)

	// Get user state
	userState := h.deps.StateManager.GetUserState(userID)
	if userState == nil {
		log.Debug().Uint("userID", userID).Msg("No valid state for user")
		return
	}

	// Find and execute the appropriate handler for the current conversation stage
	if handler, exists := h.handlers[userState.Stage]; exists {
		handler.Handle(ctx, bot, update)
	} else {
		log.Debug().
			Uint("userID", userID).
			Str("stage", string(userState.Stage)).
			Msg("No handler for current conversation stage")
	}
}
