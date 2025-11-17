package handlers

import (
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"log"
	"net/http"

	chapa "github.com/Chapa-Et/chapa-go"
	"github.com/google/uuid"
)

func VerifyPaymentHandler(w http.ResponseWriter, r *http.Request) {
	txRef := r.URL.Query().Get("trx_ref")
	if txRef == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "transaction reference is required")
		return
	}

	chapaAPI := chapa.New()
	res, err := chapaAPI.Verify(txRef)
	if err != nil {
		log.Printf("Chapa verification error: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "payment verification failed")
		return
	}
	if res.Status != "success" {
		log.Printf("Payment verification failed for tx_ref=%s: %s", txRef, res.Message)
		helpers.RespondWithError(w, http.StatusBadRequest, "payment verification failed")
		return
	}

	var InsertPaymentMutation = hasura.InsertPayment

	client := hasura.Get()
	if client == nil {
		log.Printf("hasura client not initialized")
		helpers.RespondWithError(w, http.StatusInternalServerError, "service temporarily unavailable")
		return
	}

	var query struct {
		Transaction []struct {
			TrxRef      string    `graphql:"trx_ref"`
			RecipeId    int       `graphql:"recipe_id"`
			UserId      uuid.UUID `graphql:"user_id"`
			Currency    string    `graphql:"currency"`
			CheckOutUrl string    `graphql:"checkout_url"`
			Amount      float32   `graphql:"amount"`
		} `graphql:"transaction(where: { trx_ref: { _eq: $trx_ref }})"`
	}

	variables := map[string]interface{}{"trx_ref": txRef}
	if err := client.Query(r.Context(), &query, variables); err != nil {
		log.Printf("failed to query transaction by reference: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to query transaction")
		return
	}

	if len(query.Transaction) == 0 {
		log.Printf("transaction not found for tx_ref=%s", txRef)
		helpers.RespondWithError(w, http.StatusNotFound, "transaction not found")
		return
	}

	vars := map[string]interface{}{
		"recipe_id":    query.Transaction[0].RecipeId,
		"user_id":      hasura.UUID(query.Transaction[0].UserId),
		"tx_ref":       txRef,
		"amount":       query.Transaction[0].Amount,
		"checkout_url": query.Transaction[0].CheckOutUrl,
		"currency":     query.Transaction[0].Currency,
		"status":       "success",
	}

	if err := client.Mutate(r.Context(), &InsertPaymentMutation, vars); err != nil {
		log.Printf("failed to insert payment: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to record payment")
		return
	}

	log.Printf("payment verified and recorded successfully tx_ref=%s", txRef)
	helpers.RespondWithJson(w, http.StatusOK, map[string]string{
		"message": "Payment verified successfully",
		"tx_ref":  txRef,
	})
}
