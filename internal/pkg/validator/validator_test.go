package validator_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	domainerrors "go/food-recipe/internal/domain/errors"
	ivalidator "go/food-recipe/internal/pkg/validator"
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

	err := ivalidator.ValidateStruct(validUser)
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
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

func TestValidateStruct_PreserveJSONTagCase(t *testing.T) {
	type CaseTestStruct struct {
		CamelCaseField string `json:"camelCaseField" validate:"required"`
		FallbackField  string `validate:"required"`
	}

	data := CaseTestStruct{}
	err := ivalidator.ValidateStruct(data)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(valErr.Errors))
	}

	// First field should preserve "camelCaseField"
	if valErr.Errors[0].Field != "camelCaseField" {
		t.Errorf("expected field 'camelCaseField', got '%s'", valErr.Errors[0].Field)
	}
	if valErr.Errors[0].Error != "camelCaseField is required" {
		t.Errorf("expected error 'camelCaseField is required', got '%s'", valErr.Errors[0].Error)
	}

	// Second field has no json tag -> fallback lowercased "fallbackfield"
	if valErr.Errors[1].Field != "fallbackfield" {
		t.Errorf("expected field 'fallbackfield', got '%s'", valErr.Errors[1].Field)
	}
	if valErr.Errors[1].Error != "fallbackfield is required" {
		t.Errorf("expected error 'fallbackfield is required', got '%s'", valErr.Errors[1].Error)
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "email must be a valid email address"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_MinMax_String(t *testing.T) {
	userMin := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "Hey", // min=5
		Age:       25,
	}

	err := ivalidator.ValidateStruct(userMin)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expectedMin := "bio must be at least 5 characters"
	if valErr.Errors[0].Error != expectedMin {
		t.Errorf("expected error message %q, got %q", expectedMin, valErr.Errors[0].Error)
	}

	userMax := TestUser{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Bio:       "This bio is far too long for the maximum limit", // max=20
		Age:       25,
	}

	err = ivalidator.ValidateStruct(userMax)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expectedMax := "bio must be at most 20 characters"
	if valErr.Errors[0].Error != expectedMax {
		t.Errorf("expected error message %q, got %q", expectedMax, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_MinMax_Slice(t *testing.T) {
	type SliceStruct struct {
		Items []string `json:"items" validate:"min=2,max=3"`
	}

	tooFew := SliceStruct{Items: []string{"one"}}
	err := ivalidator.ValidateStruct(tooFew)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "items must contain at least 2 items"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Errors[0].Error)
	}

	tooMany := SliceStruct{Items: []string{"one", "two", "three", "four"}}
	err = ivalidator.ValidateStruct(tooMany)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected = "items must contain at most 3 items"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_MinMax_Number(t *testing.T) {
	type NumberStruct struct {
		Count int `json:"count" validate:"min=10,max=50"`
	}

	tooLow := NumberStruct{Count: 5}
	err := ivalidator.ValidateStruct(tooLow)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "count must be at least 10"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Errors[0].Error)
	}

	tooHigh := NumberStruct{Count: 60}
	err = ivalidator.ValidateStruct(tooHigh)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected = "count must be at most 50"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Errors[0].Error)
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
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

	err := ivalidator.ValidateStruct(user)
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

	var nilValErr *ivalidator.ValidationError
	if nilValErr.Unwrap() != nil {
		t.Errorf("expected nil unwrapped from nil *ValidationError, got: %v", nilValErr.Unwrap())
	}
}

func TestValidationError_Error_Format(t *testing.T) {
	valErr := &ivalidator.ValidationError{
		Errors: []ivalidator.FieldError{
			{Field: "email", Error: "email is required"},
			{Field: "password", Error: "password must be at least 6 characters"},
		},
	}

	expected := "email is required, password must be at least 6 characters"
	if valErr.Error() != expected {
		t.Errorf("expected %q, got %q", expected, valErr.Error())
	}

	emptyErr := &ivalidator.ValidationError{}
	if emptyErr.Error() != "validation error" {
		t.Errorf("expected 'validation error', got %q", emptyErr.Error())
	}

	var nilErr *ivalidator.ValidationError
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

	err := ivalidator.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
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
	err := ivalidator.ValidateStruct(invalid)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	var valErr *ivalidator.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *validator.ValidationError, got %T", err)
	}

	expected := "id must be a valid UUID"
	if valErr.Errors[0].Error != expected {
		t.Errorf("expected error message %q, got %q", expected, valErr.Errors[0].Error)
	}
}

func TestValidateStruct_NilAndNonStruct(t *testing.T) {
	errNil := ivalidator.ValidateStruct(nil)
	if errNil == nil {
		t.Fatalf("expected validation error for nil, got nil")
	}
	if !errors.Is(errNil, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(errNil, domainerrors.ErrInvalidInput) to be true")
	}

	var typedNilPtr *TestUser = nil
	errTypedNil := ivalidator.ValidateStruct(typedNilPtr)
	if errTypedNil == nil {
		t.Fatalf("expected validation error for typed nil pointer, got nil")
	}
	if !errors.Is(errTypedNil, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(errTypedNil, domainerrors.ErrInvalidInput) to be true")
	}
	var valErr *ivalidator.ValidationError
	if errors.As(errTypedNil, &valErr) {
		if len(valErr.Errors) != 1 || valErr.Errors[0].Error != "input cannot be nil" {
			t.Errorf("unexpected error payload: %+v", valErr.Errors)
		}
	}

	errNonStruct := ivalidator.ValidateStruct("just a string")
	if errNonStruct == nil {
		t.Fatalf("expected validation error for non-struct, got nil")
	}
	if !errors.Is(errNonStruct, domainerrors.ErrInvalidInput) {
		t.Errorf("expected errors.Is(errNonStruct, domainerrors.ErrInvalidInput) to be true")
	}
}

func TestRegisterValidation(t *testing.T) {
	err := ivalidator.RegisterValidation("custom_starts_with_a", func(fl validator.FieldLevel) bool {
		return strings.HasPrefix(fl.Field().String(), "a")
	})
	if err != nil {
		t.Fatalf("failed to register custom validation: %v", err)
	}

	type CustomStruct struct {
		Code string `json:"code" validate:"custom_starts_with_a"`
	}

	invalid := CustomStruct{Code: "bcd"}
	errInvalid := ivalidator.ValidateStruct(invalid)
	if errInvalid == nil {
		t.Fatalf("expected validation error for custom rule, got nil")
	}

	valid := CustomStruct{Code: "abc"}
	errValid := ivalidator.ValidateStruct(valid)
	if errValid != nil {
		t.Fatalf("expected valid struct to pass, got: %v", errValid)
	}
}
