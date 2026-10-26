package validator_test

import (
	"errors"
	"strings"
	"testing"

	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/presentation/validator"
)

type TestUser struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `validate:"required"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Bio       string `json:"bio" validate:"min=5,max=20"`
	Age       int    `json:"age" validate:"gt=18"`
}

func TestValidateStruct_Valid(t *testing.T) {
	validUser := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "Hello world",
		Age:       25,
	}

	err := validator.ValidateStruct(validUser)
	if err != nil {
		t.Fatalf("expected nil error for valid struct, got: %v", err)
	}
}

func TestValidateStruct_JSONTagExtraction(t *testing.T) {
	user := TestUser{
		// FirstName missing -> should use "first_name" from json tag
		LastName: "Doe",
		Email:    "john.doe@example.com",
		Bio:      "Hello world",
		Age:      25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected error to be *validator.ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %+v", len(valErr.Errors), valErr.Errors)
	}

	if valErr.Errors[0].Field != "first_name" {
		t.Errorf("expected field 'first_name', got '%s'", valErr.Errors[0].Field)
	}
}

func TestValidateStruct_FallbackFieldNameToLowerCase(t *testing.T) {
	user := TestUser{
		FirstName: "John",
		// LastName missing -> has no json tag, should fallback to lowercase field name "lastname"
		Email: "john.doe@example.com",
		Bio:   "Hello world",
		Age:   25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(valErr.Errors))
	}

	if valErr.Errors[0].Field != "lastname" {
		t.Errorf("expected field 'lastname', got '%s'", valErr.Errors[0].Field)
	}
}

func TestValidateStruct_Required(t *testing.T) {
	user := TestUser{
		LastName: "Doe",
		Email:    "john.doe@example.com",
		Bio:      "Hello world",
		Age:      25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "first_name is required"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_Email(t *testing.T) {
	user := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "invalid-email",
		Bio:       "Hello world",
		Age:       25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "email must be a valid email address"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_Min(t *testing.T) {
	user := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "Hey", // min=5
		Age:       25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "bio must be at least 5 characters"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_Max(t *testing.T) {
	user := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "This bio is far too long for the maximum limit", // max=20
		Age:       25,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "bio must be at most 20 characters"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_Gt(t *testing.T) {
	user := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "Hello world",
		Age:       18, // gt=18, so 18 fails
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "age must be greater than 18"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_Unwrap(t *testing.T) {
	user := TestUser{
		FirstName: "",
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.Is(err, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(err, domainerrors.ErrInvalidInput) to be true")
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != domainerrors.ErrInvalidInput {
		t.Errorf("expected unwrapped error to be ErrInvalidInput, got: %v", unwrapped)
	}
}

func TestValidationError_Error_Format(t *testing.T) {
	valErr := &validator.ValidationError{
		Errors: []validator.FieldError{
			{Field: "email", Error: "email is required"},
			{Field: "password", Error: "password must be at least 6 characters"},
		},
	}

	expected := "email is required, password must be at least 6 characters"
	if valErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Error())
	}

	emptyErr := &validator.ValidationError{}
	if emptyErr.Error() != "validation error" {
		t.Errorf("expected 'validation error', got %q", emptyErr.Error())
	}

	var nilErr *validator.ValidationError
	if nilErr.Error() != "validation error" {
		t.Errorf("expected 'validation error', got %q", nilErr.Error())
	}
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	user := TestUser{
		FirstName: "",
		LastName:  "",
		Email:     "invalid",
		Bio:       "hi",
		Age:       10,
	}

	err := validator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 5 {
		t.Errorf("expected 5 errors, got %d", len(valErr.Errors))
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "first_name is required") {
		t.Errorf("expected error string to contain 'first_name is required', got: %s", errStr)
	}
	if !strings.Contains(errStr, "lastname is required") {
		t.Errorf("expected error string to contain 'lastname is required', got: %s", errStr)
	}
}

func TestValidateStruct_UUID(t *testing.T) {
	type UUIDStruct struct {
		ID string `json:"id" validate:"uuid"`
	}

	invalid := UUIDStruct{ID: "not-a-uuid"}
	err := validator.ValidateStruct(invalid)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *validator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "id must be a valid UUID"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_NilAndNonStruct(t *testing.T) {
	errNil := validator.ValidateStruct(nil)
	if errNil == nil {
		t.Fatalf("expected validation error for nil, got nil")
	}
	if !errors.Is(errNil, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(errNil, domainerrors.ErrInvalidInput) to be true")
	}

	errNonStruct := validator.ValidateStruct("just a string")
	if errNonStruct == nil {
		t.Fatalf("expected validation error for non-struct, got nil")
	}
	if !errors.Is(errNonStruct, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(errNonStruct, domainerrors.ErrInvalidInput) to be true")
	}
}

