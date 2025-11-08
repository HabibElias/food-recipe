package main

import (
	"fmt"
	"go/food-recipe/handlers"
	"go/food-recipe/helpers"
	"go/food-recipe/initializers"
	"log"
	"net/http"
	"os"

	"go/food-recipe/middlewares"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectClient()
}

func main() {
	portString := os.Getenv("PORT")

	if portString == "" {
		log.Fatal("PORT is not found in the environment")
		return
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", helpers.HandleErr)
	v1Router.Post("/login", handlers.HandleUserLogin)
	v1Router.Post("/signup", handlers.SignupHandler)
	// protected route returning current user
	v1Router.With(middlewares.RequireAuth).Get("/me", handlers.HandleMe)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:" + portString,
	}

	log.Printf("Server starting on port %v", portString)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PORT: ", portString)
}
