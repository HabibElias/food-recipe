package handlers

import (
	"encoding/json"
	"fmt"
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"log"
	"net/http"
	"os"
	"time"

	chapa "github.com/Chapa-Et/chapa-go"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type HasuraInput struct {
	Input struct {
		Amount   float64 `json:"amount"`
		RecipeID int     `json:"recipe_id"`
	} `json:"input"`
	Session struct {
		HasuraUserID string `json:"x-hasura-user-id"`
	} `json:"session_variables"`
}

type ChapaRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	TxRef       string `json:"tx_ref"`
	CallbackURL string `json:"callback_url"`
	ReturnURL   string `json:"return_url"`
}

type ChapaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

type InitializePaymentOutput struct {
	CheckoutURL string `json:"checkout_url"`
	TxRef       string `json:"tx_ref"`
	Status      string `json:"status"`
}

// InitializePaymentHandler handles Hasura Action
func InitializePaymentHandler(w http.ResponseWriter, r *http.Request) {
	var payload HasuraInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// Validate input
	if payload.Input.Amount <= 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	if payload.Input.RecipeID <= 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid recipe ID")
		return
	}

	userID := payload.Session.HasuraUserID
	if userID == "" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	client := hasura.Get()
	if client == nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "service temporarily unavailable")
		return
	}

	// query Hasura by primary key (id)
	query := hasura.GetUserByIDQuery

	uid, err := uuid.Parse(payload.Session.HasuraUserID)
	if err != nil {
		log.Printf("invalid user id in session: %v", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	variables := map[string]interface{}{"id": hasura.UUID(uid)}

	if err := client.Query(r.Context(), &query, variables); err != nil {
		log.Printf("failed to query user by id: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to retrieve user information")
		return
	}

	user := query.User
	txRef := fmt.Sprintf("TX-%d", time.Now().Unix())

	chapaAPI := chapa.API(chapa.New())

	req := &chapa.PaymentRequest{
		Amount:         decimal.NewFromFloat(payload.Input.Amount),
		Currency:       "ETB",
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		TransactionRef: txRef,
		Phone:          "0900123456",
		TestMode:       true,
		CallbackURL:    os.Getenv("CALLBACK_URL") + "/v1/payment/verify",
	}

	res, err := chapaAPI.PaymentRequest(req)

	if err != nil {
		log.Printf("Chapa payment request error: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to initialize payment")
		return
	}

	if res.Status != "success" {
		log.Printf("Chapa payment request failed: %s", res.Message)
		helpers.RespondWithError(w, http.StatusBadRequest, "payment initialization failed")
		return
	}

	var InsertTransaction struct {
		InsertTransaction struct {
			Returning []struct {
				Id uuid.UUID
			}
		} `graphql:"insert_transaction(objects: {trx_ref: $trx_ref, user_id: $user_id, recipe_id: $recipe_id, currency: $currency, amount: $amount, checkout_url: $checkout_url})"`
	}

	variables = map[string]interface{}{
		"user_id":      hasura.UUID(uid),
		"recipe_id":    payload.Input.RecipeID,
		"trx_ref":      txRef,
		"checkout_url": res.Data.CheckoutURL,
		"amount":       payload.Input.Amount,
		"currency":     "ETB",
	}

	if err := client.Mutate(r.Context(), &InsertTransaction, variables); err != nil {
		log.Printf("failed to insert transaction: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to record transaction")
		return
	}

	// Return response to Hasura
	output := InitializePaymentOutput{
		CheckoutURL: res.Data.CheckoutURL,
		TxRef:       txRef,
		Status:      res.Status,
	}
	helpers.RespondWithJson(w, http.StatusOK, output)
}
