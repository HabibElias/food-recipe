package main

import (
	"go/food-recipe/handlers"
	"go/food-recipe/helpers"
	"go/food-recipe/initializers"
	"log"
	"net/http"
	"os"

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

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = []string{allowedOrigins}
	} else {
		origins = []string{"*"}
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", helpers.HandleErr)
	v1Router.Post("/login", handlers.HandleUserLogin)
	v1Router.Post("/signup", handlers.SignupHandler)

	v1Router.Post("/me", handlers.HandleMe)

	v1Router.Post("/payment", handlers.InitializePaymentHandler)
	v1Router.Get("/payment/verify", handlers.VerifyPaymentHandler)
	v1Router.Post("/upload-recipe-images", handlers.UploadRecipeImagesHandler)
	v1Router.Post("/upload-user-image", handlers.UploadUserImageHandler)
	v1Router.Post("/like-notification", handlers.LikeNotificationHandler)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:" + portString,
	}

	log.Printf("Server starting on port %v", portString)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed to start: ", err)
	}
}
