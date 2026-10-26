package hasura_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
	"go/food-recipe/internal/infrastructure/persistence/hasura"
)

func TestInterfaceAdherence(t *testing.T) {
	client := hasura.NewClient(config.HasuraConfig{})

	var _ repositories.UserRepository = hasura.NewUserRepository(client)
	var _ repositories.PaymentRepository = hasura.NewPaymentRepository(client)
	var _ repositories.RecipeRepository = hasura.NewRecipeRepository(client)
	var _ repositories.NotificationRepository = hasura.NewNotificationRepository(client)
}

func TestGQLUUID(t *testing.T) {
	id := uuid.New()
	gqlUUID := hasura.GQLUUID(id)

	if gqlType := gqlUUID.GetGraphQLType(); gqlType != "uuid" {
		t.Fatalf("expected GetGraphQLType() to return 'uuid', got %s", gqlType)
	}

	// Test MarshalJSON
	marshaled, err := json.Marshal(gqlUUID)
	if err != nil {
		t.Fatalf("failed to marshal GQLUUID: %v", err)
	}
	expectedJSON := `"` + id.String() + `"`
	if string(marshaled) != expectedJSON {
		t.Fatalf("expected JSON %s, got %s", expectedJSON, string(marshaled))
	}

	// Test UnmarshalJSON
	var unmarshaled hasura.GQLUUID
	if err := json.Unmarshal(marshaled, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal GQLUUID: %v", err)
	}
	if uuid.UUID(unmarshaled) != id {
		t.Fatalf("expected unmarshaled UUID %s, got %s", id, uuid.UUID(unmarshaled))
	}

	// Test UnmarshalJSON with "null" and empty string
	var nullUnmarshaled hasura.GQLUUID
	if err := json.Unmarshal([]byte("null"), &nullUnmarshaled); err != nil {
		t.Fatalf("failed to unmarshal null: %v", err)
	}
	if uuid.UUID(nullUnmarshaled) != uuid.Nil {
		t.Fatalf("expected uuid.Nil for null, got %s", uuid.UUID(nullUnmarshaled))
	}

	var emptyUnmarshaled hasura.GQLUUID
	if err := json.Unmarshal([]byte(`""`), &emptyUnmarshaled); err != nil {
		t.Fatalf("failed to unmarshal empty string: %v", err)
	}
	if uuid.UUID(emptyUnmarshaled) != uuid.Nil {
		t.Fatalf("expected uuid.Nil for empty string, got %s", uuid.UUID(emptyUnmarshaled))
	}

	// Test MarshalText and UnmarshalText
	text, err := gqlUUID.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal text: %v", err)
	}
	if string(text) != id.String() {
		t.Fatalf("expected text %s, got %s", id.String(), string(text))
	}

	var textUnmarshaled hasura.GQLUUID
	if err := textUnmarshaled.UnmarshalText([]byte(id.String())); err != nil {
		t.Fatalf("failed to unmarshal text: %v", err)
	}
	if uuid.UUID(textUnmarshaled) != id {
		t.Fatalf("expected text unmarshaled %s, got %s", id, uuid.UUID(textUnmarshaled))
	}
}

func TestUninitializedClientBehavior(t *testing.T) {
	client := hasura.NewClient(config.HasuraConfig{URL: ""})
	if client.GQL != nil {
		t.Fatalf("expected client.GQL to be nil when URL is empty")
	}

	ctx := context.Background()
	userRepo := hasura.NewUserRepository(client)
	paymentRepo := hasura.NewPaymentRepository(client)
	recipeRepo := hasura.NewRecipeRepository(client)
	notifRepo := hasura.NewNotificationRepository(client)

	t.Run("UserRepository returns ErrServiceUnavailable when uninitialized", func(t *testing.T) {
		testUUID := uuid.New()

		if _, err := userRepo.GetByID(ctx, testUUID); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("GetByID: expected ErrServiceUnavailable, got %v", err)
		}
		if _, err := userRepo.GetByEmail(ctx, "test@example.com"); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("GetByEmail: expected ErrServiceUnavailable, got %v", err)
		}
		if _, err := userRepo.GetByUsername(ctx, "testuser"); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("GetByUsername: expected ErrServiceUnavailable, got %v", err)
		}
		if _, err := userRepo.Create(ctx, &entities.User{Email: "test@example.com"}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("Create: expected ErrServiceUnavailable, got %v", err)
		}
		if err := userRepo.UpdateAvatar(ctx, testUUID, "http://avatar.jpg"); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("UpdateAvatar: expected ErrServiceUnavailable, got %v", err)
		}
	})

	t.Run("PaymentRepository returns ErrServiceUnavailable when uninitialized", func(t *testing.T) {
		if err := paymentRepo.CreateTransaction(ctx, &entities.Transaction{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("CreateTransaction: expected ErrServiceUnavailable, got %v", err)
		}
		if _, err := paymentRepo.GetTransactionByRef(ctx, "ref-123"); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("GetTransactionByRef: expected ErrServiceUnavailable, got %v", err)
		}
		if err := paymentRepo.CreatePayment(ctx, &entities.Payment{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("CreatePayment: expected ErrServiceUnavailable, got %v", err)
		}
	})

	t.Run("RecipeRepository returns ErrServiceUnavailable when uninitialized", func(t *testing.T) {
		testUUID := uuid.New()
		if _, _, err := recipeRepo.GetCreatorFromLikeID(ctx, testUUID); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("GetCreatorFromLikeID: expected ErrServiceUnavailable, got %v", err)
		}
		if err := recipeRepo.AddRecipeImage(ctx, &entities.RecipeImage{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("AddRecipeImage: expected ErrServiceUnavailable, got %v", err)
		}
	})

	t.Run("NotificationRepository returns ErrServiceUnavailable when uninitialized", func(t *testing.T) {
		if err := notifRepo.Create(ctx, &entities.Notification{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("Create: expected ErrServiceUnavailable, got %v", err)
		}
	})

	t.Run("Nil client handling", func(t *testing.T) {
		uRepo := hasura.NewUserRepository(nil)
		if _, err := uRepo.GetByID(ctx, uuid.New()); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("expected ErrServiceUnavailable for nil client, got %v", err)
		}
		pRepo := hasura.NewPaymentRepository(nil)
		if err := pRepo.CreateTransaction(ctx, &entities.Transaction{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("expected ErrServiceUnavailable for nil client, got %v", err)
		}
		rRepo := hasura.NewRecipeRepository(nil)
		if _, _, err := rRepo.GetCreatorFromLikeID(ctx, uuid.New()); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("expected ErrServiceUnavailable for nil client, got %v", err)
		}
		nRepo := hasura.NewNotificationRepository(nil)
		if err := nRepo.Create(ctx, &entities.Notification{}); !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("expected ErrServiceUnavailable for nil client, got %v", err)
		}
	})
}

func TestClient_AdminSecretHeader(t *testing.T) {
	var receivedSecret string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSecret = r.Header.Get("x-hasura-admin-secret")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"user":[]}}`))
	}))
	defer server.Close()

	client := hasura.NewClient(config.HasuraConfig{
		URL:         server.URL,
		AdminSecret: "super-admin-secret",
	})

	userRepo := hasura.NewUserRepository(client)
	_, _ = userRepo.GetByEmail(context.Background(), "test@example.com")

	if receivedSecret != "super-admin-secret" {
		t.Fatalf("expected admin secret 'super-admin-secret', got '%s'", receivedSecret)
	}
}

type gqlRequestBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func TestUserRepository_MockServer(t *testing.T) {
	testID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequestBody
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "user_by_pk"):
			if id, ok := req.Variables["id"].(string); ok && id == testID.String() {
				w.Write([]byte(`{
					"data": {
						"user_by_pk": {
							"id": "11111111-1111-1111-1111-111111111111",
							"email": "john@example.com",
							"password": "hashed_secret_password",
							"username": "johndoe",
							"FirstName": "John",
							"LastName": "Doe",
							"avatar_url": "https://img.com/avatar.png"
						}
					}
				}`))
			} else {
				w.Write([]byte(`{"data": {"user_by_pk": null}}`))
			}

		case strings.Contains(req.Query, "user(where: {email: {_eq: $email}}"):
			if email, ok := req.Variables["email"].(string); ok && email == "john@example.com" {
				w.Write([]byte(`{
					"data": {
						"user": [{
							"id": "11111111-1111-1111-1111-111111111111",
							"email": "john@example.com",
							"password": "hashed_secret_password",
							"username": "johndoe",
							"FirstName": "John",
							"LastName": "Doe",
							"avatar_url": "https://img.com/avatar.png"
						}]
					}
				}`))
			} else {
				w.Write([]byte(`{"data": {"user": []}}`))
			}

		case strings.Contains(req.Query, "user(where: {username: {_eq: $username}}"):
			if username, ok := req.Variables["username"].(string); ok && username == "johndoe" {
				w.Write([]byte(`{
					"data": {
						"user": [{
							"id": "11111111-1111-1111-1111-111111111111",
							"email": "john@example.com",
							"password": "hashed_secret_password",
							"username": "johndoe",
							"FirstName": "John",
							"LastName": "Doe",
							"avatar_url": "https://img.com/avatar.png"
						}]
					}
				}`))
			} else {
				w.Write([]byte(`{"data": {"user": []}}`))
			}

		case strings.Contains(req.Query, "insert_user"):
			// Verify user.PasswordHash was passed to $password
			pwd := req.Variables["password"]
			if pwd != "hashed_secret_password" {
				w.Write([]byte(`{"errors": [{"message": "expected password param"}]}`))
				return
			}
			w.Write([]byte(`{
				"data": {
					"insert_user": {
						"returning": [{
							"id": "11111111-1111-1111-1111-111111111111",
							"email": "john@example.com",
							"password": "hashed_secret_password",
							"username": "johndoe",
							"FirstName": "John",
							"LastName": "Doe",
							"avatar_url": ""
						}]
					}
				}
			}`))

		case strings.Contains(req.Query, "update_user"):
			if id, ok := req.Variables["id"].(string); ok && id == testID.String() {
				w.Write([]byte(`{
					"data": {
						"update_user": {
							"returning": [{"id": "11111111-1111-1111-1111-111111111111"}]
						}
					}
				}`))
			} else {
				w.Write([]byte(`{
					"data": {
						"update_user": {
							"returning": []
						}
					}
				}`))
			}
		}
	}))
	defer server.Close()

	client := hasura.NewClient(config.HasuraConfig{URL: server.URL})
	repo := hasura.NewUserRepository(client)
	ctx := context.Background()

	t.Run("GetByID found", func(t *testing.T) {
		user, err := repo.GetByID(ctx, testID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != testID {
			t.Errorf("expected ID %s, got %s", testID, user.ID)
		}
		if user.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", user.Email)
		}
		// Confirm password query result mapped to user.PasswordHash
		if user.PasswordHash != "hashed_secret_password" {
			t.Errorf("expected PasswordHash 'hashed_secret_password', got '%s'", user.PasswordHash)
		}
		if user.FirstName != "John" || user.LastName != "Doe" {
			t.Errorf("expected John Doe, got %s %s", user.FirstName, user.LastName)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("GetByEmail found", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "john@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", user.Email)
		}
		if user.PasswordHash != "hashed_secret_password" {
			t.Errorf("expected PasswordHash 'hashed_secret_password', got '%s'", user.PasswordHash)
		}
	})

	t.Run("GetByEmail not found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("GetByUsername found", func(t *testing.T) {
		user, err := repo.GetByUsername(ctx, "johndoe")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Username != "johndoe" {
			t.Errorf("expected username johndoe, got %s", user.Username)
		}
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		_, err := repo.GetByUsername(ctx, "nobody")
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("Create user maps PasswordHash to $password", func(t *testing.T) {
		newUser := &entities.User{
			Email:        "john@example.com",
			Username:     "johndoe",
			FirstName:    "John",
			LastName:     "Doe",
			PasswordHash: "hashed_secret_password",
		}

		created, err := repo.Create(ctx, newUser)
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}
		if created.ID != testID {
			t.Errorf("expected created ID %s, got %s", testID, created.ID)
		}
		if created.PasswordHash != "hashed_secret_password" {
			t.Errorf("expected PasswordHash to be preserved/mapped, got %s", created.PasswordHash)
		}
	})

	t.Run("Create user nil input", func(t *testing.T) {
		_, err := repo.Create(ctx, nil)
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("UpdateAvatar success", func(t *testing.T) {
		err := repo.UpdateAvatar(ctx, testID, "https://new-avatar.png")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("UpdateAvatar not found", func(t *testing.T) {
		err := repo.UpdateAvatar(ctx, uuid.New(), "https://new-avatar.png")
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestPaymentRepository_MockServer(t *testing.T) {
	testUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	txID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequestBody
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "insert_transaction"):
			if ref, ok := req.Variables["trx_ref"].(string); ok && ref == "TX-EMPTY" {
				w.Write([]byte(`{
					"data": {
						"insert_transaction": {
							"returning": []
						}
					}
				}`))
				return
			}
			w.Write([]byte(`{
				"data": {
					"insert_transaction": {
						"returning": [{"id": "33333333-3333-3333-3333-333333333333"}]
					}
				}
			}`))

		case strings.Contains(req.Query, "transaction(where: {trx_ref: {_eq: $trx_ref}}"):
			if ref, ok := req.Variables["trx_ref"].(string); ok && ref == "TX-123456" {
				w.Write([]byte(`{
					"data": {
						"transaction": [{
							"id": "33333333-3333-3333-3333-333333333333",
							"trx_ref": "TX-123456",
							"recipe_id": 42,
							"user_id": "22222222-2222-2222-2222-222222222222",
							"currency": "ETB",
							"checkout_url": "https://chapa.co/pay/123",
							"amount": 250.50
						}]
					}
				}`))
			} else {
				w.Write([]byte(`{"data": {"transaction": []}}`))
			}

		case strings.Contains(req.Query, "insert_payment"):
			if ref, ok := req.Variables["tx_ref"].(string); ok && ref == "TX-EMPTY" {
				w.Write([]byte(`{
					"data": {
						"insert_payment": {
							"returning": []
						}
					}
				}`))
				return
			}
			w.Write([]byte(`{
				"data": {
					"insert_payment": {
						"returning": [{"id": "pay-uuid-789"}]
					}
				}
			}`))
		}
	}))
	defer server.Close()

	client := hasura.NewClient(config.HasuraConfig{URL: server.URL})
	repo := hasura.NewPaymentRepository(client)
	ctx := context.Background()

	t.Run("CreateTransaction success", func(t *testing.T) {
		tx := &entities.Transaction{
			UserID:      testUserID,
			RecipeID:    42,
			TrxRef:      "TX-123456",
			Amount:      250.50,
			Currency:    "ETB",
			CheckoutURL: "https://chapa.co/pay/123",
		}
		if err := repo.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.ID != txID {
			t.Errorf("expected tx ID %s, got %s", txID, tx.ID)
		}
	})

	t.Run("CreateTransaction nil input", func(t *testing.T) {
		if err := repo.CreateTransaction(ctx, nil); !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("CreateTransaction empty returning returns error", func(t *testing.T) {
		tx := &entities.Transaction{
			UserID:      testUserID,
			RecipeID:    42,
			TrxRef:      "TX-EMPTY",
			Amount:      250.50,
			Currency:    "ETB",
			CheckoutURL: "https://chapa.co/pay/123",
		}
		err := repo.CreateTransaction(ctx, tx)
		if err == nil {
			t.Fatalf("expected error on empty returning, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create transaction") {
			t.Fatalf("expected 'failed to create transaction' error, got %v", err)
		}
	})

	t.Run("GetTransactionByRef found", func(t *testing.T) {
		tx, err := repo.GetTransactionByRef(ctx, "TX-123456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.TrxRef != "TX-123456" {
			t.Errorf("expected trx_ref TX-123456, got %s", tx.TrxRef)
		}
		if tx.RecipeID != 42 {
			t.Errorf("expected recipe_id 42, got %d", tx.RecipeID)
		}
		if tx.Amount != 250.50 {
			t.Errorf("expected amount 250.50, got %f", tx.Amount)
		}
	})

	t.Run("GetTransactionByRef not found", func(t *testing.T) {
		_, err := repo.GetTransactionByRef(ctx, "nonexistent-ref")
		if !errors.Is(err, domainerrors.ErrTransactionNotFound) {
			t.Fatalf("expected ErrTransactionNotFound, got %v", err)
		}
	})

	t.Run("CreatePayment success", func(t *testing.T) {
		payment := &entities.Payment{
			UserID:      testUserID,
			RecipeID:    42,
			TxRef:       "TX-123456",
			Amount:      250.50,
			CheckoutURL: "https://chapa.co/pay/123",
			Currency:    "ETB",
			Status:      "success",
		}
		if err := repo.CreatePayment(ctx, payment); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.ID != "pay-uuid-789" {
			t.Errorf("expected payment ID 'pay-uuid-789', got '%s'", payment.ID)
		}
	})

	t.Run("CreatePayment nil input", func(t *testing.T) {
		if err := repo.CreatePayment(ctx, nil); !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("CreatePayment empty returning returns error", func(t *testing.T) {
		payment := &entities.Payment{
			UserID:      testUserID,
			RecipeID:    42,
			TxRef:       "TX-EMPTY",
			Amount:      250.50,
			CheckoutURL: "https://chapa.co/pay/123",
			Currency:    "ETB",
			Status:      "success",
		}
		err := repo.CreatePayment(ctx, payment)
		if err == nil {
			t.Fatalf("expected error on empty returning, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create payment") {
			t.Fatalf("expected 'failed to create payment' error, got %v", err)
		}
	})
}

func TestRecipeAndNotificationRepository_MockServer(t *testing.T) {
	likeID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ownerID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	notifID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlRequestBody
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(req.Query, "recipe_like_by_pk"):
			if id, ok := req.Variables["id"].(string); ok && id == likeID.String() {
				w.Write([]byte(`{
					"data": {
						"recipe_like_by_pk": {
							"id": "44444444-4444-4444-4444-444444444444",
							"recipe": {
								"user_id": "55555555-5555-5555-5555-555555555555"
							},
							"user": {
								"id": "77777777-7777-7777-7777-777777777777",
								"username": "chef_fan"
							}
						}
					}
				}`))
			} else {
				w.Write([]byte(`{"data": {"recipe_like_by_pk": null}}`))
			}

		case strings.Contains(req.Query, "insert_recipe_image"):
			w.Write([]byte(`{
				"data": {
					"insert_recipe_image": {
						"returning": [{"id": "img-abc-123"}]
					}
				}
			}`))

		case strings.Contains(req.Query, "insert_notifications"):
			w.Write([]byte(`{
				"data": {
					"insert_notifications": {
						"returning": [{"id": "66666666-6666-6666-6666-666666666666"}]
					}
				}
			}`))
		}
	}))
	defer server.Close()

	client := hasura.NewClient(config.HasuraConfig{URL: server.URL})
	recipeRepo := hasura.NewRecipeRepository(client)
	notifRepo := hasura.NewNotificationRepository(client)
	ctx := context.Background()

	t.Run("GetCreatorFromLikeID found", func(t *testing.T) {
		returnedOwner, returnedUsername, err := recipeRepo.GetCreatorFromLikeID(ctx, likeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if returnedOwner != ownerID {
			t.Errorf("expected owner %s, got %s", ownerID, returnedOwner)
		}
		if returnedUsername != "chef_fan" {
			t.Errorf("expected liker chef_fan, got %s", returnedUsername)
		}
	})

	t.Run("GetCreatorFromLikeID not found", func(t *testing.T) {
		_, _, err := recipeRepo.GetCreatorFromLikeID(ctx, uuid.New())
		if err == nil {
			t.Fatalf("expected error for non-existent like ID, got nil")
		}
	})

	t.Run("AddRecipeImage success", func(t *testing.T) {
		img := &entities.RecipeImage{
			RecipeID:    10,
			ImageURL:    "https://img.com/food.jpg",
			IsThumbnail: true,
		}
		if err := recipeRepo.AddRecipeImage(ctx, img); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if img.ID != "img-abc-123" {
			t.Errorf("expected img.ID img-abc-123, got %s", img.ID)
		}
	})

	t.Run("AddRecipeImage nil input", func(t *testing.T) {
		if err := recipeRepo.AddRecipeImage(ctx, nil); !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("CreateNotification success", func(t *testing.T) {
		n := &entities.Notification{
			UserID:      ownerID,
			Message:     "user @chef_fan liked your recipe",
			InitiatorID: uuid.MustParse("77777777-7777-7777-7777-777777777777"),
		}
		if err := notifRepo.Create(ctx, n); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.ID != notifID {
			t.Errorf("expected notification ID %s, got %s", notifID, n.ID)
		}
	})

	t.Run("CreateNotification nil input", func(t *testing.T) {
		if err := notifRepo.Create(ctx, nil); !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})
}
