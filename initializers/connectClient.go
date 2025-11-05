package initializers

import (
	"log"

	"go/food-recipe/clients/hasura"
)

func ConnectClient() {
	hasura.Init()
	if hasura.Get() == nil {
		log.Println("initializers: Hasura client not initialized")
	}
}
