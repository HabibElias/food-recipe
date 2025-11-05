package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type loginInput struct {
	Email    string
	Password string
}

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            loginArgs              `json:"input"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type LoginOutput struct {
	User_id string `json:"user_id"`
	Token   string `json:"token"`
}

type Mutation struct {
	Login *LoginOutput
}

type loginArgs struct {
	Input loginInput
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleUserLogin(w http.ResponseWriter, r *http.Request) {

	// set the response header as JSON
	w.Header().Set("Content-Type", "application/json")

	// read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// parse the body as action payload
	var actionPayload LoginActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := login(actionPayload.Input)

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

func login(args loginArgs) (LoginOutput, error) {
	client := hasura.Get()
	if client == nil {
		return LoginOutput{}, fmt.Errorf("hasura client not initialized")
	}

	query := hasura.GetUserByEmailQuery
	variables := map[string]interface{}{"email": args.Input.Email}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		return LoginOutput{}, fmt.Errorf("failed to query user by email: %w", err)
	}

	if len(query.User) == 0 {
		return LoginOutput{}, fmt.Errorf("invalid credentials")
	}

	user := query.User[0]

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(args.Input.Password)); err != nil {
		return LoginOutput{}, fmt.Errorf("invalid credentials")
	}

	token, err := helpers.GenerateHasuraJWT(user.ID, user.Email)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return LoginOutput{
		User_id: user.ID,
		Token:   token,
	}, nil
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.RespondWithError(w, 405, "method not allowed")
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("invalid login payload: %v", err)
		helpers.RespondWithError(w, 400, "invalid request body")
		return
	}

	client := hasura.Get()
	if client == nil {
		helpers.RespondWithError(w, 500, "hasura client not initialized")
		return
	}

	query := hasura.GetUserByEmailQuery
	variables := map[string]interface{}{"email": req.Email}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		log.Printf("Can't get users, Message: %v", err)
		helpers.RespondWithError(w, 500, "failed to query user")
		return
	}
	if len(query.User) != 0 {
		helpers.RespondWithError(w, 401, "email already used")
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Can't hash the password: %v", err)
		helpers.RespondWithError(w, 500, "failed to hash password")
		return
	}

	// Prepare variables for insert mutation
	insertVariables := map[string]interface{}{
		"email":    req.Email,
		"password": string(hash),
		"name":     req.Name,
	}

	// Run insert mutation
	insertMutation := hasura.InsertUser
	if err := client.Mutate(context.Background(), &insertMutation, insertVariables); err != nil {
		log.Printf("Can't insert user, Message: %v", err)
		helpers.RespondWithError(w, 500, "failed to create user")
		return
	}

	if len(insertMutation.InsertUsers.Returning) == 0 {
		helpers.RespondWithError(w, 500, "failed to create user")
		return
	}

	createdUser := insertMutation.InsertUsers.Returning[0]

	token, err := helpers.GenerateHasuraJWT(createdUser.ID, createdUser.Email)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		helpers.RespondWithError(w, 500, "failed to generate token")
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")
	helpers.RespondWithJson(
		w,
		200,
		LoginOutput{
			User_id: createdUser.ID,
			Token:   token,
		},
	)
}
