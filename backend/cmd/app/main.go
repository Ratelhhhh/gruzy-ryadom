package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"gruzy-ryadom/config"
	"gruzy-ryadom/internal/api"
	"gruzy-ryadom/internal/bots"
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database connection
	database, err := db.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Service layer
	svc := service.New(database)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait group for all services
	var wg sync.WaitGroup

	// Start REST API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		startAPIServer(ctx, cfg, svc)
	}()

	// Start Driver Bot
	wg.Add(1)
	go func() {
		defer wg.Done()
		startDriverBot(ctx, cfg, svc)
	}()

	// Start Admin Bot
	wg.Add(1)
	go func() {
		defer wg.Done()
		startAdminBot(ctx, cfg, svc)
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down application...")
	cancel()

	// Wait for all services to finish
	wg.Wait()
	log.Println("Application stopped")
}

func startAPIServer(ctx context.Context, cfg *config.Config, svc *service.Service) {
	// API layer
	apiHandler := api.New(svc)

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Mount("/", apiHandler.Routes())

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("API Server forced to shutdown: %v", err)
	} else {
		log.Println("API Server stopped gracefully")
	}
}

func startDriverBot(ctx context.Context, cfg *config.Config, svc *service.Service) {
	bot, err := bots.NewDriverBot(cfg.Bots.DriverBotToken, svc)
	if err != nil {
		log.Fatalf("Failed to create driver bot: %v", err)
	}

	log.Println("Driver Bot starting...")
	
	// Start bot in goroutine
	go func() {
		bot.Start()
	}()

	// Wait for context cancellation
	<-ctx.Done()
	
	// Stop bot
	bot.Stop()
	log.Println("Driver Bot stopped")
}

func startAdminBot(ctx context.Context, cfg *config.Config, svc *service.Service) {
	bot, err := bots.NewAdminBot(cfg.Bots.AdminBotToken, svc)
	if err != nil {
		log.Fatalf("Failed to create admin bot: %v", err)
	}

	log.Println("Admin Bot starting...")
	
	// Start bot in goroutine
	go func() {
		bot.Start()
	}()

	// Wait for context cancellation
	<-ctx.Done()
	
	// Stop bot
	bot.Stop()
	log.Println("Admin Bot stopped")
} 