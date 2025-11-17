package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"

	"golang.org/x/crypto/bcrypt"
)

type SignUpActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            signupArgs             `json:"input"`
}

type signupArgs struct {
	Input signupInput
}
type signupOutput struct {
	Token   string `json:"token"`
	User_id string `json:"user_id"`
}

type signupInput struct {
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {

	// set the response header as JSON
	w.Header().Set("Content-Type", "application/json")

	// read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// parse the body as action payload
	var actionPayload SignUpActionPayload
	if err := json.Unmarshal(reqBody, &actionPayload); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := signup(actionPayload.Input)

	// throw if an error happens
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Write the response as JSON
	helpers.RespondWithJson(w, http.StatusOK, result)

}

// Auto-generated function that takes the Action parameters and must return it's response type
func signup(args signupArgs) (response signupOutput, err error) {
	in := args.Input

	// validation errors collected here
	errs := map[string]string{}

	// Validate first name
	in.First_name = strings.TrimSpace(in.First_name)
	if in.First_name == "" {
		errs["first_name"] = "first name is required"
	} else if len(in.First_name) < 2 {
		errs["first_name"] = "first name must be at least 2 characters"
	}

	// Validate last name
	in.Last_name = strings.TrimSpace(in.Last_name)
	if in.Last_name == "" {
		errs["last_name"] = "last name is required"
	} else if len(in.Last_name) < 2 {
		errs["last_name"] = "last name must be at least 2 characters"
	}

	// Validate password
	if in.Password == "" {
		errs["password"] = "password is required"
	} else if len(in.Password) < 6 {
		errs["password"] = "password must be at least 6 characters"
	}

	// Validate username
	in.Username = strings.TrimSpace(in.Username)
	if in.Username == "" {
		errs["username"] = "username is required"
	} else if len(in.Username) < 5 {
		errs["username"] = "username must be at least 5 characters"
	}

	// Validate email
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" {
		errs["email"] = "email is required"
	} else if !strings.Contains(in.Email, "@") {
		errs["email"] = "email must be a valid email address"
	}

	if len(errs) > 0 {
		errJSON, _ := json.Marshal(errs)
		return signupOutput{}, fmt.Errorf("validation failed: %s", string(errJSON))
	}

	client := hasura.Get()
	if client == nil {
		return signupOutput{}, fmt.Errorf("service temporarily unavailable")
	}

	// check email uniqueness
	emailQuery := hasura.GetUserByEmailQuery
	if err := client.Query(context.Background(), &emailQuery, map[string]interface{}{"email": in.Email}); err != nil {
		log.Printf("failed to query user by email: %v", err)
		return signupOutput{}, fmt.Errorf("failed to verify email availability")
	}
	if len(emailQuery.User) > 0 {
		return signupOutput{}, fmt.Errorf("email already registered")
	}

	// check username uniqueness
	usernameQuery := hasura.GetUserByUsernameQuery
	if err := client.Query(context.Background(), &usernameQuery, map[string]interface{}{"username": in.Username}); err != nil {
		log.Printf("failed to query user by username: %v", err)
		return signupOutput{}, fmt.Errorf("failed to verify username availability")
	}
	if len(usernameQuery.User) > 0 {
		return signupOutput{}, fmt.Errorf("username already taken")
	}

	// Hash password with cost 12 for better security
	const bcryptCost = 12
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return signupOutput{}, fmt.Errorf("failed to process password")
	}

	// Run insert mutation
	insertVariables := map[string]interface{}{
		"email":      in.Email,
		"password":   string(hash),
		"username":   in.Username,
		"first_name": in.First_name,
		"last_name":  in.Last_name,
	}

	insertMutation := hasura.InsertUser
	if err := client.Mutate(context.Background(), &insertMutation, insertVariables); err != nil {
		log.Printf("failed to insert user: %v", err)
		return signupOutput{}, fmt.Errorf("failed to create user account")
	}

	if len(insertMutation.InsertUsers.Returning) == 0 {
		return signupOutput{}, fmt.Errorf("failed to create user account")
	}

	createdUser := insertMutation.InsertUsers.Returning[0]

	token, err := helpers.GenerateHasuraJWT(createdUser.ID, createdUser.Email)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		return signupOutput{}, fmt.Errorf("failed to generate authentication token")
	}

	return signupOutput{Token: token, User_id: createdUser.ID}, nil
}
