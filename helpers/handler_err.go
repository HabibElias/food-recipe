package helpers

import "net/http"

func HandleErr(w http.ResponseWriter, _ *http.Request) {
	RespondWithError(w, 400, "Something went wrong")
}
