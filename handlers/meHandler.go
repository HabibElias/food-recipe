package handlers

import (
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"go/food-recipe/middlewares"
	"log"
	"net/http"
	"strconv"
)

func HandleMe(w http.ResponseWriter, r *http.Request) {
	userIDStr, _, ok := middlewares.GetUserFromContext(r.Context())
	if !ok || userIDStr == "" {
		helpers.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	client := hasura.Get()
	if client == nil {
		helpers.RespondWithError(w, 500, "hasura client not initialized")
		return
	}

	// query Hasura by primary key (id)
	query := hasura.GetUserByIDQuery

	num, err := strconv.Atoi(userIDStr) // Convert string to int

	if err != nil {
		log.Fatalf("Error converting string to integer: %v", err)
	}

	variables := map[string]interface{}{"id": num}

	if err := client.Query(r.Context(), &query, variables); err != nil {
		log.Printf("Can't get user by id, Message: %v", err)
		helpers.RespondWithError(w, 500, "failed to query user")
		return
	}

	user := query.User

	helpers.RespondWithJson(w, 200, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.UserName,
			"email":    user.Email,
		},
	})
}
