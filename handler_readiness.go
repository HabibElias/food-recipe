package main

import (
	"go/food-recipe/helpers"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	helpers.RespondWithJson(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
