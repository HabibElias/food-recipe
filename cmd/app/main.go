package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/infrastructure/payment/chapa"
	"go/food-recipe/internal/infrastructure/persistence/hasura"
	"go/food-recipe/internal/infrastructure/storage/cloudinary"
	"go/food-recipe/internal/presentation/http/handlers"
	"go/food-recipe/internal/presentation/http/middleware"
	"go/food-recipe/internal/presentation/http/router"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Infrastructure Adapters
	hasuraClient := hasura.NewClient(cfg.Hasura)
	userRepo := hasura.NewUserRepository(hasuraClient)
	paymentRepo := hasura.NewPaymentRepository(hasuraClient)
	recipeRepo := hasura.NewRecipeRepository(hasuraClient)
	notificationRepo := hasura.NewNotificationRepository(hasuraClient)

	cloudinaryStorage, err := cloudinary.NewCloudinaryStorage(cfg.Cloudinary)
	if err != nil {
		log.Printf("Warning: failed to init cloudinary: %v", err)
	}

	chapaGateway := chapa.NewChapaGateway(cfg.Chapa)

	// Initialize Application Services
	authService := services.NewAuthService(userRepo, cfg)
	paymentService := services.NewPaymentService(paymentRepo, userRepo, chapaGateway, cfg)
	mediaService := services.NewMediaService(cloudinaryStorage, recipeRepo, userRepo)
	notificationService := services.NewNotificationService(recipeRepo, notificationRepo)

	// Initialize Presentation Handlers
	authHandler := handlers.NewAuthHandler(authService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	healthHandler := handlers.NewHealthHandler()

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 200 * 1024 * 1024,
	})

	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.RecoverMiddleware())
	app.Use(middleware.CORSMiddleware(cfg.App.AllowedOrigins))

	// Setup Routes
	router.SetupRoutes(router.RouterParams{
		App:                 app,
		AuthHandler:         authHandler,
		PaymentHandler:      paymentHandler,
		MediaHandler:        mediaHandler,
		NotificationHandler: notificationHandler,
		HealthHandler:       healthHandler,
		ActionSecret:        cfg.Hasura.ActionSecret,
	})

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", cfg.App.Port)
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			log.Printf("Server listen error: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down server gracefully...")

	shutdownTimeout := 5 * time.Second
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
