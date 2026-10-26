# Design Specification: Clean Architecture Refactor (DSLServer Pattern)

## 1. Overview
This specification defines the complete refactoring of the `food-recipe` backend to adhere to the Clean Architecture patterns established in the `DSLServer` project (`../../work/DSLServer`). 

The current codebase couples HTTP routing (`Chi`), Hasura GraphQL calls, and business logic directly inside HTTP handlers and top-level packages (`handlers/`, `clients/`, `helpers/`). 

The refactored architecture adopts:
- Fiber v2 (`github.com/gofiber/fiber/v2`) for presentation and HTTP routing.
- Strict Clean Architecture layer separation under `cmd/` and `internal/`:
  - `cmd/app`: Dependency injection, server bootstrap, and graceful shutdown.
  - `internal/domain`: Pure entities, repository interfaces, external port interfaces, and domain error definitions (no third-party web or database framework dependencies).
  - `internal/application`: DTOs and application services (use cases).
  - `internal/infrastructure`: Persistence implementations via Hasura GraphQL client, Cloudinary storage adapter, Chapa payment gateway adapter, and security utilities.
  - `internal/presentation`: Fiber v2 handlers, middlewares, routes, and standardized response helpers.
  - `internal/config`: Strongly typed configuration loaded via environment variables.

---

## 2. Target Directory Structure

```
food-recipe/
├── cmd/
│   └── app/
│       └── main.go                         # Server entrypoint & DI wireup
├── internal/
│   ├── config/
│   │   └── config.go                       # Strongly typed Config
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user.go
│   │   │   ├── payment.go
│   │   │   ├── recipe.go
│   │   │   └── notification.go
│   │   ├── repositories/
│   │   │   ├── user_repository.go
│   │   │   ├── payment_repository.go
│   │   │   ├── recipe_repository.go
│   │   │   └── notification_repository.go
│   │   ├── ports/
│   │   │   ├── file_storage.go
│   │   │   └── payment_gateway.go
│   │   └── errors/
│   │       └── errors.go
│   ├── application/
│   │   ├── dto/
│   │   │   ├── auth_dto.go
│   │   │   ├── payment_dto.go
│   │   │   ├── media_dto.go
│   │   │   └── notification_dto.go
│   │   └── services/
│   │       ├── auth_service.go
│   │       ├── payment_service.go
│   │       ├── media_service.go
│   │       └── notification_service.go
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   └── hasura/
│   │   │       ├── client.go
│   │   │       ├── models.go
│   │   │       ├── user_repository.go
│   │   │       ├── payment_repository.go
│   │   │       ├── recipe_repository.go
│   │   │       └── notification_repository.go
│   │   ├── storage/
│   │   │   └── cloudinary/
│   │   │       └── storage.go
│   │   ├── payment/
│   │   │   └── chapa/
│   │   │       └── gateway.go
│   │   └── security/
│   │       ├── jwt.go
│   │       └── password.go
│   └── presentation/
│       ├── http/
│       │   ├── handlers/
│       │   │   ├── auth_handler.go
│       │   │   ├── payment_handler.go
│       │   │   ├── media_handler.go
│       │   │   ├── notification_handler.go
│       │   │   └── health_handler.go
│       │   ├── middleware/
│       │   │   ├── cors.go
│       │   │   ├── logger.go
│       │   │   ├── recover.go
│       │   │   └── auth.go
│       │   └── router/
│       │       └── router.go
│       └── response/
│           ├── response.go
│           └── errors.go
├── go.mod
└── go.sum
```

---

## 3. Domain Layer Specifications

### 3.1 Entities
- **User**: `ID` (`uuid.UUID`), `Email`, `Username`, `FirstName`, `LastName`, `PasswordHash`, `AvatarURL`.
- **Transaction**: `ID` (`uuid.UUID`), `UserID` (`uuid.UUID`), `RecipeID` (`int`), `TrxRef` (`string`), `Amount` (`float64`), `Currency` (`string`), `CheckoutURL` (`string`).
- **Payment**: `ID` (`string`), `UserID` (`uuid.UUID`), `RecipeID` (`int`), `TxRef` (`string`), `Amount` (`float64`), `CheckoutURL` (`string`), `Currency` (`string`), `Status` (`string`).
- **RecipeImage**: `ID` (`string`), `RecipeID` (`int`), `ImageURL` (`string`), `IsThumbnail` (`bool`).
- **Notification**: `ID` (`uuid.UUID`), `UserID` (`uuid.UUID`), `Message` (`string`), `InitiatorID` (`uuid.UUID`).

### 3.2 Repository Interfaces
```go
package repositories

type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    GetByUsername(ctx context.Context, username string) (*entities.User, error)
    Create(ctx context.Context, user *entities.User) (*entities.User, error)
    UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
}

type PaymentRepository interface {
    CreateTransaction(ctx context.Context, tx *entities.Transaction) error
    GetTransactionByRef(ctx context.Context, txRef string) (*entities.Transaction, error)
    CreatePayment(ctx context.Context, payment *entities.Payment) error
}

type RecipeRepository interface {
    GetCreatorFromLikeID(ctx context.Context, likeID uuid.UUID) (recipeOwnerID uuid.UUID, likerUsername string, err error)
    AddRecipeImage(ctx context.Context, img *entities.RecipeImage) error
}

type NotificationRepository interface {
    Create(ctx context.Context, n *entities.Notification) error
}
```

### 3.3 Ports
```go
package ports

type FileStorage interface {
    UploadBase64Image(ctx context.Context, base64Data, publicID, folder string) (string, error)
}

type PaymentInitParams struct {
    Amount      float64
    Currency    string
    FirstName   string
    LastName    string
    Email       string
    TxRef       string
    Phone       string
    CallbackURL string
}

type PaymentInitResult struct {
    CheckoutURL string
    TxRef       string
    Status      string
}

type PaymentVerifyResult struct {
    Status  string
    Message string
}

type PaymentGateway interface {
    Initialize(ctx context.Context, params PaymentInitParams) (*PaymentInitResult, error)
    Verify(ctx context.Context, txRef string) (*PaymentVerifyResult, error)
}
```

### 3.4 Domain Errors
Defined in `internal/domain/errors/errors.go`:
```go
var (
    ErrUserNotFound        = errors.New("user not found")
    ErrUserAlreadyExists   = errors.New("email already registered")
    ErrUsernameTaken       = errors.New("username already taken")
    ErrInvalidCredentials  = errors.New("invalid credentials")
    ErrUnauthorized        = errors.New("unauthorized")
    ErrInvalidInput        = errors.New("invalid input")
    ErrTransactionNotFound = errors.New("transaction not found")
    ErrPaymentFailed       = errors.New("payment failed")
    ErrServiceUnavailable  = errors.New("service temporarily unavailable")
)
```

---

## 4. Application Layer Specifications

### 4.1 Data Transfer Objects (`internal/application/dto`)
- **`auth_dto.go`**:
  - `LoginInputDTO`: `Email`, `Password`.
  - `LoginActionPayload`: `Input`, `SessionVariables`.
  - `LoginOutputDTO`: `UserID`, `Token`.
  - `SignupInputDTO`: `FirstName`, `LastName`, `Email`, `Username`, `Password`.
  - `SignupActionPayload`: `Input`, `SessionVariables`.
  - `SignupOutputDTO`: `UserID`, `Token`.
  - `MeInputDTO`: `Token`.
  - `MeActionPayload`: `Input`, `SessionVariables`.
  - `MeOutputDTO`: `UserID`, `Token`.
- **`payment_dto.go`**:
  - `InitializePaymentInputDTO`: `Amount`, `RecipeID`.
  - `InitializePaymentActionPayload`: `Input`, `SessionVariables`.
  - `InitializePaymentOutputDTO`: `CheckoutURL`, `TxRef`, `Status`.
  - `VerifyPaymentQueryDTO`: `TxRef`.
  - `VerifyPaymentOutputDTO`: `Message`, `TxRef`.
- **`media_dto.go`**:
  - `RecipeImageDTO`: `ImageName`, `ImageType`, `ImageBase64Str`, `IsThumbnail`.
  - `UploadRecipeImagesInputDTO`: `RecipeID`, `RecipeImages []RecipeImageDTO`.
  - `UploadRecipeImagesActionPayload`: `Input`, `SessionVariables`.
  - `UploadRecipeImagesOutputDTO`: `Success bool`.
  - `UploadUserImageInputDTO`: `UserID`, `Image RecipeImageDTO`.
  - `UploadUserImageActionPayload`: `Input`, `SessionVariables`.
  - `UploadUserImageOutputDTO`: `Success bool`.
- **`notification_dto.go`**:
  - `LikeEventPayloadDTO`: Hasura event trigger payload schema for `recipe_like`.

### 4.2 Application Services (`internal/application/services`)
- **`AuthService`**:
  - `Login(ctx, input)`: Validates input, queries user by email via repository, compares bcrypt password, produces Hasura JWT.
  - `Signup(ctx, input)`: Validates input fields, checks email & username uniqueness, hashes password, saves user via repository, generates Hasura JWT.
  - `Me(ctx, token)`: Verifies Hasura JWT, queries user by email via repository, generates fresh Hasura JWT.
- **`PaymentService`**:
  - `InitializePayment(ctx, input, sessionUserID)`: Validates input, retrieves user by ID via repository, calls `PaymentGateway.Initialize()`, records transaction in repository, returns checkout URL and transaction reference.
  - `VerifyPayment(ctx, txRef)`: Calls `PaymentGateway.Verify()`, retrieves transaction by `txRef`, records confirmed payment row via repository.
- **`MediaService`**:
  - `UploadRecipeImages(ctx, input)`: Validates image base64, uploads images to Cloudinary via `FileStorage`, inserts `RecipeImage` records via repository.
  - `UploadUserImage(ctx, input, sessionUserID)`: Validates image, uploads to Cloudinary via `FileStorage`, updates user avatar URL via repository.
- **`NotificationService`**:
  - `HandleLikeNotification(ctx, event)`: Filters for `INSERT` operation, retrieves recipe owner and liker username via `RecipeRepository`, creates `Notification` via `NotificationRepository`.

---

## 5. Infrastructure Layer Specifications

### 5.1 Persistence (`internal/infrastructure/persistence/hasura`)
- Wraps `github.com/hasura/go-graphql-client` with typed GraphQL queries and mutations.
- Implements:
  - `UserRepository`
  - `PaymentRepository`
  - `RecipeRepository`
  - `NotificationRepository`
- Maps GraphQL response records to domain `entities`.

### 5.2 Storage (`internal/infrastructure/storage/cloudinary`)
- Wraps Cloudinary Go SDK.
- Implements `ports.FileStorage`.

### 5.3 Payment (`internal/infrastructure/payment/chapa`)
- Wraps `github.com/Chapa-Et/chapa-go`.
- Implements `ports.PaymentGateway`.

### 5.4 Security (`internal/infrastructure/security`)
- `jwt.go`:
  - `GenerateHasuraJWT(userID, email string, secret string) (string, error)`
  - `VerifyHasuraJWT(tokenStr string, secret string) (jwt.MapClaims, error)`
- `password.go`:
  - `HashPassword(password string) (string, error)` (bcrypt cost 12)
  - `ComparePassword(hash, password string) error`

---

## 6. Presentation Layer Specifications (Fiber v2)

### 6.1 Handlers (`internal/presentation/http/handlers`)
Fiber handlers accepting `*fiber.Ctx`, parsing incoming request bodies via `c.BodyParser()`, invoking corresponding service methods, and delegating response serialization to `response` package:
- `AuthHandler` (`Login`, `Signup`, `Me`)
- `PaymentHandler` (`InitializePayment`, `VerifyPayment`)
- `MediaHandler` (`UploadRecipeImages`, `UploadUserImage`)
- `NotificationHandler` (`HandleLikeNotification`)
- `HealthHandler` (`Healthz`, `Err`)

### 6.2 Middleware (`internal/presentation/http/middleware`)
- `LoggerMiddleware()`
- `RecoverMiddleware()`
- `CORSMiddleware()`
- `AuthMiddleware()`: Validates Hasura Bearer JWT and sets `user_id` and `claims` in `c.Locals`.

### 6.3 Routing (`internal/presentation/http/router`)
Registers all endpoints under `/v1`:
- `GET /v1/healthz`
- `GET /v1/err`
- `POST /v1/login`
- `POST /v1/signup`
- `POST /v1/me`
- `POST /v1/payment`
- `GET /v1/payment/verify`
- `POST /v1/upload-recipe-images`
- `POST /v1/upload-user-image`
- `POST /v1/like-notification`

### 6.4 Response Helpers (`internal/presentation/response`)
- `OK(c, data)`
- `Created(c, data)`
- `BadRequest(c, msg)`
- `Unauthorized(c, msg)`
- `NotFound(c, msg)`
- `Conflict(c, msg)`
- `Internal(c, msg)`
- `HandleDomainError(c, err)`: Translates domain errors to corresponding HTTP status and JSON response.

---

## 7. Migration Plan & Backward Compatibility
1. Update `go.mod` to include `github.com/gofiber/fiber/v2`.
2. Create all directories and files under `internal/` and `cmd/app`.
3. Clean up obsolete top-level folders (`clients/`, `handlers/`, `helpers/`, `initializers/`, `middlewares/`, `handler_readiness.go`, and root `main.go`).
4. Update root `main.go` or root redirect to `cmd/app/main.go` so `go build ./cmd/app` and `go run ./cmd/app` work smoothly.
5. Verify with `go build ./...` and `go vet ./...`.
