package main

import (
	"context"
	"fmt"
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
	"gruzy-ryadom/internal/api"
	"gruzy-ryadom/internal/bots"
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/service"
)

type Application struct {
	server     *http.Server
	adminBot   *bots.AdminBot
	driverBot  *bots.DriverBot
	service    *service.Service
	database   *db.DB
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewApplication() (*Application, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	database, err := db.New(dbURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Service layer
	svc := service.New(database)

	// Create bots
	adminBotToken := os.Getenv("ADMIN_BOT_TOKEN")
	if adminBotToken == "" {
		cancel()
		database.Close()
		return nil, fmt.Errorf("ADMIN_BOT_TOKEN is required")
	}

	driverBotToken := os.Getenv("DRIVER_BOT_TOKEN")
	if driverBotToken == "" {
		cancel()
		database.Close()
		return nil, fmt.Errorf("DRIVER_BOT_TOKEN is required")
	}

	adminBot, err := bots.NewAdminBot(adminBotToken, svc)
	if err != nil {
		cancel()
		database.Close()
		return nil, fmt.Errorf("failed to create admin bot: %w", err)
	}

	driverBot, err := bots.NewDriverBot(driverBotToken, svc)
	if err != nil {
		cancel()
		database.Close()
		return nil, fmt.Errorf("failed to create driver bot: %w", err)
	}

	// Create HTTP server
	apiHandler := api.New(svc)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		server:    server,
		adminBot:  adminBot,
		driverBot: driverBot,
		service:   svc,
		database:  database,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (app *Application) Start() error {
	log.Println("Starting application...")

	// Start bots in goroutines
	app.wg.Add(2)
	go func() {
		defer app.wg.Done()
		log.Println("Starting admin bot...")
		app.adminBot.Start()
	}()

	go func() {
		defer app.wg.Done()
		log.Println("Starting driver bot...")
		app.driverBot.Start()
	}()

	// Start HTTP server in goroutine
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		log.Printf("Starting HTTP server on port %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (app *Application) Stop() error {
	log.Println("Stopping application...")

	// Cancel context to signal shutdown
	app.cancel()

	// Stop bots
	if app.adminBot != nil {
		app.adminBot.Stop()
	}
	if app.driverBot != nil {
		app.driverBot.Stop()
	}

	// Shutdown HTTP server
	if app.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := app.server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}

	// Close database
	if app.database != nil {
		app.database.Close()
	}

	// Wait for all goroutines to finish
	app.wg.Wait()

	log.Println("Application stopped")
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create application
	app, err := NewApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal...")

	// Stop application
	if err := app.Stop(); err != nil {
		log.Printf("Error stopping application: %v", err)
	}
} 