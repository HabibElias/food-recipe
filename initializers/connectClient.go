package initializers

import (
	"log"

	"go/food-recipe/clients/cloudinary"
	"go/food-recipe/clients/hasura"
)

func ConnectClient() {
	hasura.Init()
	if hasura.Get() == nil {
		log.Println("initializers: Hasura client not initialized")
	}

	cloudinary.Init()
	if cloudinary.Get() == nil {
		log.Println("initializers: Cloudinary client not initialized")
	}
}
