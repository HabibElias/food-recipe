# Design Spec: DTO Validation System

**Date:** 2026-10-26  
**Status:** Approved  
**Related Reference:** `../../work/DSLServer/internal/utils/validator.go`

---

## 1. Objective

Implement a robust, centralized DTO validation mechanism matching the Clean Architecture and validator patterns used in `DSLServer`:
- Use `github.com/go-playground/validator/v10` for tag-based struct validation.
- Register JSON field name resolution so client error messages reference API payload fields (`email`, `first_name`) instead of Go struct names.
- Provide a `ValidationError` type that implements `Error() string` and unwraps to `domainerrors.ErrInvalidInput`.
- Expose a clean `.Validate() error` method on all input DTOs.
- Integrate validation in HTTP handlers immediately following request parsing, preserving Hasura Action error contracts (`{"message": "..."}`).

---

## 2. Architecture & Components

### 2.1 Validator Utility (`internal/infrastructure/validator`)
Located in `internal/infrastructure/validator/validator.go`:

```go
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
		if name == "" || name == "-" {
			return fld.Name
		}
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		return name
	})
}

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

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

func (v *ValidationError) Unwrap() error {
	return domainerrors.ErrInvalidInput
}

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return &ValidationError{Errors: []FieldError{{Field: "", Error: "invalid input"}}}
	}
	errors := make([]FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		errors = append(errors, FieldError{
			Field: strings.ToLower(fieldErr.Field()),
			Error: getErrorMessage(fieldErr),
		})
	}
	return &ValidationError{Errors: errors}
}
```

#### User-Friendly Error Messages
- `required`: `<field> is required`
- `email`: `<field> must be a valid email address`
- `min`: `<field> must be at least <param> characters` (or `>= <param>` for numbers/slices)
- `max`: `<field> must be at most <param> characters`
- `gt`: `<field> must be greater than <param>`
- `uuid`: `<field> must be a valid UUID`
- `dive`: recurses into elements of slices/arrays

---

### 2.2 DTO Validation Rules & Interfaces

All input DTOs implement:
```go
type Validatable interface {
	Validate() error
}
```

#### Authentication DTOs (`internal/application/dto/auth_dto.go`)
- **`LoginInputDTO`**:
  ```go
  type LoginInputDTO struct {
      Email    string `json:"email" validate:"required,email"`
      Password string `json:"password" validate:"required,min=6"`
  }
  func (d LoginInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```
- **`SignupInputDTO`**:
  ```go
  type SignupInputDTO struct {
      FirstName string `json:"first_name" validate:"required,min=2,max=100"`
      LastName  string `json:"last_name" validate:"required,min=2,max=100"`
      Email     string `json:"email" validate:"required,email"`
      Username  string `json:"username" validate:"required,min=3,max=50"`
      Password  string `json:"password" validate:"required,min=6,max=100"`
  }
  func (d SignupInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```
- **`MeInputDTO`**:
  ```go
  type MeInputDTO struct {
      Token string `json:"token" validate:"required"`
  }
  func (d MeInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```

#### Payment DTOs (`internal/application/dto/payment_dto.go`)
- **`InitializePaymentInputDTO`**:
  ```go
  type InitializePaymentInputDTO struct {
      Amount   float64 `json:"amount" validate:"required,gt=0"`
      RecipeID int     `json:"recipe_id" validate:"required,gt=0"`
      Phone    string  `json:"phone,omitempty" validate:"omitempty,min=9"`
  }
  func (d InitializePaymentInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```
- **`VerifyPaymentInputDTO`**:
  ```go
  type VerifyPaymentInputDTO struct {
      TrxRef string `json:"trx_ref" validate:"required"`
  }
  func (d VerifyPaymentInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```

#### Media DTOs (`internal/application/dto/media_dto.go`)
- **`RecipeImageItemDTO`**:
  ```go
  type RecipeImageItemDTO struct {
      ImageName      string `json:"image_name" validate:"required"`
      ImageType      string `json:"image_type" validate:"required"`
      ImageBase64Str string `json:"image_base64str" validate:"required"`
      IsThumbnail    bool   `json:"is_thumbnail"`
  }
  ```
- **`UploadRecipeImagesInputDTO`**:
  ```go
  type UploadRecipeImagesInputDTO struct {
      RecipeID     int                  `json:"recipe_id" validate:"required,gt=0"`
      RecipeImages []RecipeImageItemDTO `json:"recipe_images" validate:"required,min=1,dive"`
  }
  func (d UploadRecipeImagesInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```
- **`UserImageItemDTO`**:
  ```go
  type UserImageItemDTO struct {
      ImageName      string `json:"image_name" validate:"required"`
      ImageType      string `json:"image_type" validate:"required"`
      ImageBase64Str string `json:"image_base64str" validate:"required"`
  }
  ```
- **`UploadUserImageInputDTO`**:
  ```go
  type UploadUserImageInputDTO struct {
      UserID string           `json:"user_id" validate:"required,uuid"`
      Image  UserImageItemDTO `json:"image" validate:"required"`
  }
  func (d UploadUserImageInputDTO) Validate() error { return validator.ValidateStruct(d) }
  ```

#### Notification DTOs (`internal/application/dto/notification_dto.go`)
- **`LikeEventDataNewDTO`**:
  ```go
  type LikeEventDataNewDTO struct {
      ID        uuid.UUID `json:"id" validate:"required"`
      UserID    uuid.UUID `json:"user_id" validate:"required"`
      RecipeID  int       `json:"recipe_id" validate:"required,gt=0"`
      CreatedAt string    `json:"created_at"`
  }
  func (d LikeEventDataNewDTO) Validate() error { return validator.ValidateStruct(d) }
  ```

---

### 2.3 Response Handling Integration

In `internal/presentation/response/errors.go`:
```go
var valErr *validator.ValidationError
if errors.As(err, &valErr) {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "message": valErr.Error(),
        "errors":  valErr.Errors,
    })
}
```
This guarantees that Hasura Actions receive the standard `{"message": "..."}` format while clients receive both the message summary and the granular list of field errors.

---

## 3. Testing Plan

1. **Unit Tests for Validator Engine**:
   - `internal/infrastructure/validator/validator_test.go`:
     - Test valid struct passes with `nil`.
     - Test invalid email, missing required fields, min/max constraints, gt constraints.
     - Test JSON tag name mapping (e.g. `first_name` reported instead of `FirstName`).
     - Test `Unwrap()` returns `domainerrors.ErrInvalidInput`.

2. **Unit Tests for DTO Validation**:
   - `internal/application/dto/dto_test.go`:
     - Test each DTO's `.Validate()` with valid data and invalid combinations.
     - Test nested struct & slice validation (`dive`).

3. **Handler & Presentation Tests**:
   - `internal/presentation/http/handlers/handlers_test.go`:
     - Test endpoint responses with malformed/missing fields return HTTP 400 with expected error messages.
