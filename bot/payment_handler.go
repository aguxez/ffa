package bot

import (
	"context"

	"github.com/aguxez/ffa/i18n" // Import i18n
	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const (
	PREMIUM_PRICE_XTR = 1500
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
	userID := uint(update.CallbackQuery.From.ID)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	logger := log.
		With().
		Uint("userID", userID).
		Logger()

	// Get user language
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		logger.Err(err).Msg("Error getting user")
		ErrorResponse(ctx, bot, chatID, i18n.GetText(i18n.KeyErrorResponseGeneric, "en"))
		return
	}
	userLang := user.Language

	// Send invoice
	_, err = bot.SendInvoice(ctx, &tgbot.SendInvoiceParams{
		ChatID: update.CallbackQuery.From.ID,
		Title:  i18n.GetText(i18n.KeyInvoiceTitle, userLang),
		// We set the payload as an empty string because we do not have various products
		// so the purchase will be for a single item and we do not need metadata
		// for this as of now.
		Payload:     "premium-purchase",
		Description: i18n.GetText(i18n.KeyInvoiceDescription, userLang),
		Currency:    "XTR",
		Prices: []tgmodels.LabeledPrice{
			{
				Label:  i18n.GetText(i18n.KeyInvoicePriceLabel, userLang),
				Amount: PREMIUM_PRICE_XTR,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Error sending invoice")
		return
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
	userID := uint(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	paymentID := update.Message.SuccessfulPayment.TelegramPaymentChargeID

	logger := log.
		With().
		Uint("userID", userID).
		Str("paymentID", paymentID).
		Logger()

	// Get user language first. If this fails, we cannot proceed safely.
	user, err := h.deps.DB.GetUser(userID)
	if err != nil {
		// Log the error and inform the user (using English as fallback)
		logger.Err(err).Msg("Critical: Error getting user after successful payment. Cannot update premium status.")
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   i18n.GetText(i18n.KeyErrorPaymentProcessing, "en"),
		})
		return
	}
	userLang := user.Language

	// Now attempt to update user status
	err = h.deps.DB.UpdateUserPremiumStatus(userID, true, &paymentID)
	if err != nil {
		logger.Err(err).Msg("Error updating user premium status after successful payment")
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   i18n.GetText(i18n.KeyErrorPaymentProcessing, userLang),
		})
		return
	}

	// Send translated success message
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.GetText(i18n.KeyPaymentSuccess, userLang),
		ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{
					{
						Text:         i18n.GetText(i18n.KeyBackToOptionsButton, userLang),
						CallbackData: string(CallbackDataBackToOptions),
					},
				},
			},
		},
	})
}
