package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/presentation/response"
)

// PaymentHandler handles payment initialization and verification endpoints.
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler constructs a new PaymentHandler.
func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// InitializePayment handles payment initialization Hasura action.
func (h *PaymentHandler) InitializePayment(c *fiber.Ctx) error {
	var payload dto.InitializePaymentActionPayloadDTO
	if err := c.BodyParser(&payload); err != nil {
		return response.BadRequest(c, "invalid request payload")
	}

	input := payload.ToInputDTO()
	if err := input.Validate(); err != nil {
		return response.HandleDomainError(c, err)
	}

	sessionUserID := payload.GetSessionUserID()
	if sessionUserID == "" && payload.SessionVariables != nil {
		for _, key := range []string{"x-hasura-user-id", "hasura_user_id", "HasuraUserID", "user_id"} {
			if v, ok := payload.SessionVariables[key]; ok {
				if s, ok := v.(string); ok && s != "" {
					sessionUserID = s
					break
				}
			}
		}
	}
	if sessionUserID == "" {
		if v, ok := c.Locals("user_id").(string); ok {
			sessionUserID = v
		}
	}

	res, err := h.paymentService.InitializePayment(c.Context(), input, sessionUserID)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}

// VerifyPayment handles Chapa callback/verification endpoint.
func (h *PaymentHandler) VerifyPayment(c *fiber.Ctx) error {
	txRef := strings.TrimSpace(c.Query("trx_ref"))
	if txRef == "" {
		txRef = strings.TrimSpace(c.Query("tx_ref"))
	}

	input := dto.VerifyPaymentInputDTO{TrxRef: txRef}
	if err := input.Validate(); err != nil {
		return response.HandleDomainError(c, err)
	}

	res, err := h.paymentService.VerifyPayment(c.Context(), input.TrxRef)
	if err != nil {
		return response.HandleDomainError(c, err)
	}

	return response.OK(c, res)
}
