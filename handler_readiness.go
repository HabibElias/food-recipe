package main

import (
	"go/food-recipe/helpers"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithJson(w, 200, struct{}{})
}
