package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// parse the body as action payload
	var actionPayload SignUpActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := signup(actionPayload.Input)

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
func signup(args signupArgs) (response signupOutput, err error) {
	in := args.Input

	// validation errors collected here
	errs := map[string]string{}

	if len(in.First_name) <= 2 {
		errs["first_name"] = "first name must be greater than 3 characters"
	}
	if len(in.Last_name) <= 2 {
		errs["last_name"] = "last name must be greater than 3 characters"
	}
	if len(in.Password) <= 5 {
		errs["password"] = "password must be greater than 6 characters"
	}
	if len(in.Username) <= 4 {
		errs["username"] = "username must be greater than 5 characters"
	}
	if in.Username == "" {
		errs["username"] = "username must not be empty"
	}
	if in.Email == "" {
		errs["email"] = "email must not be empty"
	}

	if len(errs) > 0 {
		b, _ := json.Marshal(errs)
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	client := hasura.Get()
	if client == nil {
		b, _ := json.Marshal(map[string]string{"server": "hasura client not initialized"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	// check email uniqueness
	emailQuery := hasura.GetUserByEmailQuery
	if err := client.Query(context.Background(), &emailQuery, map[string]interface{}{"email": in.Email}); err != nil {
		log.Printf("failed to query user by email: %v", err)
		b, _ := json.Marshal(map[string]string{"server": "failed to query user by email"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}
	if len(emailQuery.User) > 0 {
		b, _ := json.Marshal(map[string]string{"email": "email already used"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	// check username uniqueness
	usernameQuery := hasura.GetUserByUsernameQuery
	if err := client.Query(context.Background(), &usernameQuery, map[string]interface{}{"username": in.Username}); err != nil {
		log.Printf("failed to query user by username: %v", err)
		b, _ := json.Marshal(map[string]string{"server": "failed to query user by username"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}
	if len(usernameQuery.User) > 0 {
		b, _ := json.Marshal(map[string]string{"username": "username already used"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Can't hash the password: %v", err)
		b, _ := json.Marshal(map[string]string{"server": "failed to hash password"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
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
		log.Printf("Can't insert user, Message: %v", err)
		b, _ := json.Marshal(map[string]string{"server": "failed to create user"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	if len(insertMutation.InsertUsers.Returning) == 0 {
		b, _ := json.Marshal(map[string]string{"server": "failed to create user"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	createdUser := insertMutation.InsertUsers.Returning[0]

	token, err := helpers.GenerateHasuraJWT(createdUser.ID, createdUser.Email)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		b, _ := json.Marshal(map[string]string{"server": "failed to generate token"})
		return signupOutput{}, fmt.Errorf("%s", string(b))
	}

	return signupOutput{Token: token, User_id: createdUser.ID}, nil
}
