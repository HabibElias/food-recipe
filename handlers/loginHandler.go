package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"io"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type loginInput struct {
	Email    string
	Password string
}

type LoginActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            loginArgs              `json:"input"`
}

type LoginOutput struct {
	User_id string `json:"user_id"`
	Token   string `json:"token"`
}

type loginArgs struct {
	Input loginInput
}


func HandleUserLogin(w http.ResponseWriter, r *http.Request) {

	// set the response header as JSON
	w.Header().Set("Content-Type", "application/json")

	// read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// parse the body as action payload
	var actionPayload LoginActionPayload
	if err := json.Unmarshal(reqBody, &actionPayload); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := login(actionPayload.Input)

	// throw if an error happens
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Write the response as JSON
	helpers.RespondWithJson(w, http.StatusOK, result)

}

func login(args loginArgs) (LoginOutput, error) {
	client := hasura.Get()
	if client == nil {
		return LoginOutput{}, fmt.Errorf("service temporarily unavailable")
	}

	// Normalize email input
	email := strings.TrimSpace(strings.ToLower(args.Input.Email))
	if email == "" {
		return LoginOutput{}, fmt.Errorf("email is required")
	}

	query := hasura.GetUserByEmailQuery
	variables := map[string]interface{}{"email": email}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		return LoginOutput{}, fmt.Errorf("failed to query user by email: %w", err)
	}

	if len(query.User) == 0 {
		return LoginOutput{}, fmt.Errorf("invalid credentials")
	}

	user := query.User[0]

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(args.Input.Password)); err != nil {
		// Use generic error message to prevent user enumeration
		return LoginOutput{}, fmt.Errorf("invalid credentials")
	}

	token, err := helpers.GenerateHasuraJWT(user.ID.String(), user.Email)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return LoginOutput{
		User_id: user.ID.String(),
		Token:   token,
	}, nil
}

