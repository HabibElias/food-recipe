package hasura

import (
	"log"
	"net/http"
	"os"

	graphql "github.com/hasura/go-graphql-client"
)

var client *graphql.Client

func Init() {
	url := os.Getenv("HASURA_URL")
	if url == "" {
		log.Println("HASURA_URL not set; Hasura client will not be initialized")
		return
	}

	hasuraKey := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	c := graphql.NewClient(url, &http.Client{}).
		WithRequestModifier(func(req *http.Request) {
			if hasuraKey != "" {
				req.Header.Set("x-hasura-admin-secret", hasuraKey)
			}
		})

	client = c
	log.Println("hasura: client initialized")
}

func Get() *graphql.Client {
	return client
}
