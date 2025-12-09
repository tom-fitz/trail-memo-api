package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tom-fitz/trailmemo-api/config"
	"github.com/tom-fitz/trailmemo-api/internal/database"
	"github.com/tom-fitz/trailmemo-api/internal/handlers"
	"github.com/tom-fitz/trailmemo-api/internal/middleware"
	"github.com/tom-fitz/trailmemo-api/internal/repository"
	"github.com/tom-fitz/trailmemo-api/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Firebase service
	firebaseService, err := services.NewFirebaseService(
		cfg.FirebaseProjectID,
		cfg.FirebaseStorageBucket,
		cfg.FirebaseServiceAccountPath,
		cfg.FirebaseServiceAccountJSON,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	memoRepo := repository.NewMemoRepository(db)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(userRepo, firebaseService)
	memoHandler := handlers.NewMemoHandler(memoRepo, userRepo, firebaseService, cfg.MaxUploadSize)

	// Set up Gin router
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint (no auth required)
	r.GET("/health", healthHandler.Check)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			// Register requires authentication (Firebase token to get user info)
			auth.POST("/register", middleware.AuthMiddleware(firebaseService), authHandler.Register)
			auth.GET("/me", middleware.AuthMiddleware(firebaseService), authHandler.GetMe)
		}

		// Memo routes (all require authentication)
		memos := v1.Group("/memos")
		memos.Use(middleware.AuthMiddleware(firebaseService))
		{
			memos.POST("", memoHandler.Create)
			memos.GET("", memoHandler.List)
			memos.GET("/nearby", memoHandler.GetNearby)
			memos.GET("/search", memoHandler.Search)
			memos.GET("/:id", memoHandler.GetByID)
			memos.PUT("/:id", memoHandler.Update)
			memos.DELETE("/:id", memoHandler.Delete)
		}
	}

	// Start server
	log.Printf("🚀 TrailMemo API server starting on port %s", cfg.Port)
	log.Printf("📍 Environment: %s", cfg.Environment)
	log.Printf("🗄️  Database connected")
	log.Printf("🔥 Firebase initialized")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
