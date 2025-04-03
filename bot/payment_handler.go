package bot

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

// PaymentHandler handles payment
type PaymentHandler struct {
	deps *HandlerDependencies
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(deps *HandlerDependencies) *PaymentHandler {
	return &PaymentHandler{
		deps: deps,
	}
}

func (h *PaymentHandler) Handle(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	logger := log.
		With().
		Str("handler", "PaymentHandler").
		Logger()

	if update.CallbackQuery != nil {
		logger.Debug().Msg("Handling initial payment request")
		h.handleInitialPaymentRequest(ctx, bot, update)
		return
	}

	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		logger.Debug().Msg("Handling successful payment")
		h.handleSuccessfulPayment(ctx, bot, update)
	}

	if update.PreCheckoutQuery != nil {
		logger.Debug().Msg("Handling payment confirmation")
		h.handlePreCheckoutQuery(ctx, bot, update)
	}
}

// handleInitialPaymentRequest handles the initial payment request
func (h *PaymentHandler) handleInitialPaymentRequest(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	userID := update.CallbackQuery.From.ID
	logger := log.
		With().
		Int64("userID", userID).
		Logger()

	// Send invoice
	_, err := bot.SendInvoice(ctx, &tgbot.SendInvoiceParams{
		ChatID: userID,
		Title:  "RawPlan Premium",
		Description: `
Access to all features of RawPlan and future developments.
Get a higher limit number of meal plans per week and timed follow-ups.
This is a non-recurring subscription to get all features of the product!
`,
		Currency: "XTR",
		Prices: []tgmodels.LabeledPrice{
			{
				Label:  "Premium test",
				Amount: 100,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending invoice")
	}
}

// handlePreCheckoutQuery handles the payment confirmation
func (h *PaymentHandler) handlePreCheckoutQuery(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	preCheckoutQueryID := update.PreCheckoutQuery.ID

	log.Debug().
		Str("preCheckoutQueryID", preCheckoutQueryID).
		Msg("Handling pre checkout query")

	// Simply answer the pre checkout query
	bot.AnswerPreCheckoutQuery(ctx, &tgbot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: preCheckoutQueryID,
		OK:                 true,
	})
}

// handleSuccessfulPayment handles the successful payment
func (h *PaymentHandler) handleSuccessfulPayment(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	// Update user status
	userID := uint(update.Message.From.ID)
	paymentID := update.Message.SuccessfulPayment.TelegramPaymentChargeID

	logger := log.
		With().
		Uint("userID", userID).
		Str("paymentID", paymentID).
		Logger()

	err := h.deps.DB.UpdateUserPremiumStatus(userID, true, &paymentID)
	if err != nil {
		logger.Err(err).Msg("Error updating user status")

		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: userID,
			Text: `
There was an error processing your successful payment. This has been informed.
`,
		})
		return
	}

	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `
You have successfully upgraded to RawPlan Premium!
`,
	})
}
