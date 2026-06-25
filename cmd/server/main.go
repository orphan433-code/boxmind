package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"pet-link/internal/config"
	"pet-link/internal/handler"
	"pet-link/internal/middleware"
	"pet-link/internal/pkg/gemini"
	"pet-link/internal/pkg/jwt"
	"pet-link/internal/pkg/pagemeta"
	"pet-link/internal/repository/postgres"
	"pet-link/internal/service"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	healthService := service.NewHealthService("boxmind-api", db)
	healthHandler := handler.NewHealthHandler(healthService)

	userRepo := postgres.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	bookmarkRepo := postgres.NewBookmarkRepository(db)
	geminiEnricher, err := gemini.NewEnricher(ctx, cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatalf("gemini enricher init failed: %v", err)
	}
	imageFetcher := pagemeta.NewImageFetcher(pagemeta.NewImageHTTPExtractor())
	metaFallback := pagemeta.NewMetaFallback(pagemeta.NewHTTPExtractor())
	bookmarkService := service.NewBookmarkService(bookmarkRepo, geminiEnricher, imageFetcher, metaFallback)
	bookmarkHandler := handler.NewBookmarkHandler(bookmarkService)

	otpRepo := postgres.NewOTPRepository(db)
	tokenProvider := jwt.NewProvider(cfg.JWTSecret, cfg.JWTTTL)
	emailSender := service.NewEmailSender(cfg.SMTP)
	authService := service.NewAuthService(
		otpRepo,
		userService,
		tokenProvider,
		emailSender,
		cfg.JWTSecret,
		cfg.OTPTTL,
	)
	authHandler := handler.NewAuthHandler(authService)

	app := fiber.New(fiber.Config{
		AppName:      "boxmind-api",
		ServerHeader: "boxmind",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == "" {
				return true
			}
			if strings.HasPrefix(origin, "chrome-extension://") {
				return true
			}
			return origin == "http://localhost:5173" ||
				origin == "http://127.0.0.1:5173"
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(requestid.New())
	app.Use("/api", logger.New())

	api := app.Group("/api/v1")

	// публичные
	api.Get("/health", healthHandler.Check)
	api.Post("/auth/login", authHandler.RequestLogin)
	api.Post("/auth/verify", authHandler.VerifyLogin)

	// защищённые — middleware только на эту группу
	protected := api.Group("", middleware.Auth(tokenProvider))
	protected.Get("/me", userHandler.Me)
	protected.Post("/bookmarks", bookmarkHandler.Create)
	protected.Get("/bookmarks", bookmarkHandler.List)
	protected.Get("/bookmarks/:id", bookmarkHandler.GetByID)
	protected.Delete("/bookmarks/:id", bookmarkHandler.Delete)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down server...")
		if err := app.Shutdown(); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("server starting on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}

	log.Println("server stopped")
}
