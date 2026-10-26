package response

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/infrastructure/validator"
)

// HandleDomainError maps domain errors to appropriate HTTP responses.
func HandleDomainError(c *fiber.Ctx, err error) error {
	if err == nil {
		return OK(c, nil)
	}

	var valErr *validator.ValidationError
	if errors.As(err, &valErr) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": valErr.Error(),
			"errors":  valErr.Errors,
		})
	}

	if errors.Is(err, domainerrors.ErrUserNotFound) || errors.Is(err, domainerrors.ErrTransactionNotFound) {
		return NotFound(c, err.Error())
	}
	if errors.Is(err, domainerrors.ErrUserAlreadyExists) || errors.Is(err, domainerrors.ErrUsernameTaken) {
		return Conflict(c, err.Error())
	}
	if errors.Is(err, domainerrors.ErrInvalidCredentials) || errors.Is(err, domainerrors.ErrUnauthorized) {
		return Unauthorized(c, err.Error())
	}
	if errors.Is(err, domainerrors.ErrInvalidInput) || errors.Is(err, domainerrors.ErrPaymentFailed) {
		return BadRequest(c, err.Error())
	}
	if errors.Is(err, domainerrors.ErrServiceUnavailable) {
		return ServiceUnavailable(c, err.Error())
	}

	log.Printf("[ERROR] %s %s - unhandled error: %v", c.Method(), c.Path(), err)
	return Internal(c, "internal server error")
}
