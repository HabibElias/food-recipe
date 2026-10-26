package validator

import (
	"fmt"
	"reflect"
	"strings"

	domainerrors "go/food-recipe/internal/domain/errors"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		if name == "" || name == "-" {
			return strings.ToLower(fld.Name)
		}
		return name
	})
}

// RegisterValidation adds a custom validation function to the validator instance.
func RegisterValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// FieldError represents a single field validation failure.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ValidationError aggregates all field errors for a validated struct.
type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

// Error returns a comma-joined list of all validation error messages.
func (v *ValidationError) Error() string {
	if v == nil || len(v.Errors) == 0 {
		return "validation error"
	}
	parts := make([]string, 0, len(v.Errors))
	for _, e := range v.Errors {
		parts = append(parts, e.Error)
	}
	return strings.Join(parts, ", ")
}

// Unwrap allows ValidationError to match domainerrors.ErrInvalidInput via errors.Is.
func (v *ValidationError) Unwrap() error {
	if v == nil {
		return nil
	}
	return domainerrors.ErrInvalidInput
}

// ValidateStruct validates a struct using validator/v10 tags and returns a ValidationError on failure.
func ValidateStruct(s interface{}) error {
	if s == nil {
		return &ValidationError{
			Errors: []FieldError{
				{Field: "", Error: "input cannot be nil"},
			},
		}
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return &ValidationError{
			Errors: []FieldError{
				{Field: "", Error: "input cannot be nil"},
			},
		}
	}

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return &ValidationError{
			Errors: []FieldError{
				{Field: "", Error: "invalid input"},
			},
		}
	}

	errors := make([]FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		errors = append(errors, FieldError{
			Field: fieldErr.Field(),
			Error: getErrorMessage(fieldErr),
		})
	}

	return &ValidationError{Errors: errors}
}

// getErrorMessage maps validator.FieldError tags to human-friendly error messages.
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		switch err.Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
		case reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s must contain at least %s items", field, err.Param())
		default:
			return fmt.Sprintf("%s must be at least %s", field, err.Param())
		}
	case "max":
		switch err.Kind() {
		case reflect.String:
			return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		case reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s must contain at most %s items", field, err.Param())
		default:
			return fmt.Sprintf("%s must be at most %s", field, err.Param())
		}
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
