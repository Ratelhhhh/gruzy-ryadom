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
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/service"
)

type APIOnlyApplication struct {
	server   *http.Server
	service  *service.Service
	database *db.DB
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewAPIOnlyApplication() (*APIOnlyApplication, error) {
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

	return &APIOnlyApplication{
		server:   server,
		service:  svc,
		database: database,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (app *APIOnlyApplication) Start() error {
	log.Println("Starting API-only application...")

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

func (app *APIOnlyApplication) Stop() error {
	log.Println("Stopping API-only application...")

	// Cancel context to signal shutdown
	app.cancel()

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

	log.Println("API-only application stopped")
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create API-only application
	app, err := NewAPIOnlyApplication()
	if err != nil {
		log.Fatalf("Failed to create API-only application: %v", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start API-only application: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal...")

	// Stop application
	if err := app.Stop(); err != nil {
		log.Printf("Error stopping API-only application: %v", err)
	}
} 