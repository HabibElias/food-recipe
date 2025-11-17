package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"io/ioutil"
	"net/http"
)

type MeActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            MeArgs                 `json:"input"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type MeOutput struct {
	Token   string      `json:"token"`
	User_id hasura.UUID `json:"user_id"`
}

type MeInput struct {
	Token string
}

type MeArgs struct {
	Input MeInput
}

func HandleMe(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// parse the body as action payload
	var meActionPayload MeActionPayload
	err = json.Unmarshal(reqBody, &meActionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := Me(meActionPayload.Input)

	// throw if an error happens
	if err != nil {
		errorObject := GraphQLError{
			Message: err.Error(),
		}
		errorBody, _ := json.Marshal(errorObject)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorBody)
		return
	}

	// Write the response as JSON
	data, _ := json.Marshal(result)
	w.Write(data)

}

// Auto-generated function that takes the Action parameters and must return it's response type
func Me(args MeArgs) (response MeOutput, err error) {
	claims, err := helpers.VerifyHasuraJWT(args.Input.Token)

	if err != nil {
		return MeOutput{}, fmt.Errorf("error occurred: %s", err)
	}

	client := hasura.Get()
	if client == nil {
		return MeOutput{}, fmt.Errorf("service temporarily unavailable")
	}

	query := hasura.GetUserByEmailQuery
	variables := map[string]interface{}{"email": claims["email"]}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		return MeOutput{}, fmt.Errorf("failed to query user by email: %w", err)
	}

	newToken, err := helpers.GenerateHasuraJWT(query.User[0].ID.String(), query.User[0].Email)

	if err != nil {
		return MeOutput{}, fmt.Errorf("failed to generate a new token: %w", err)
	}

	return MeOutput{
		Token:   newToken,
		User_id: hasura.UUID(query.User[0].ID),
	}, nil
}
