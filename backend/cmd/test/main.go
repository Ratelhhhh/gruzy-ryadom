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

	"gruzy-ryadom/config"
	"gruzy-ryadom/internal/api"
	"gruzy-ryadom/internal/bots"
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

type TestApplication struct {
	server    *http.Server
	adminBot  *bots.AdminBot
	driverBot *bots.DriverBot
	service   *service.Service
	database  *db.DB
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	hasBots   bool
}

func NewTestApplication() (*TestApplication, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Database connection
	if cfg.Database.URL == "" {
		cancel()
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	database, err := db.New(cfg.Database.URL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Service layer
	svc := service.New(database)

	// Try to create bots, but don't fail if they can't be created
	var adminBot *bots.AdminBot
	var driverBot *bots.DriverBot
	hasBots := false

	if cfg.Bots.AdminBotToken != "" && cfg.Bots.DriverBotToken != "" {
		log.Println("Attempting to create Telegram bots...")

		adminBot, err = bots.NewAdminBot(cfg.Bots.AdminBotToken, svc)
		if err != nil {
			log.Printf("Warning: Failed to create admin bot: %v", err)
			log.Println("Application will run without admin bot")
		} else {
			driverBot, err = bots.NewDriverBot(cfg.Bots.DriverBotToken, svc)
			if err != nil {
				log.Printf("Warning: Failed to create driver bot: %v", err)
				log.Println("Application will run without driver bot")
			} else {
				hasBots = true
				log.Println("Both bots created successfully")
			}
		}
	} else {
		log.Println("Bot tokens not provided, running without bots")
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
		status := map[string]interface{}{
			"status": "OK",
			"bots":   hasBots,
			"time":   time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON response
		response := fmt.Sprintf(`{"status":"%s","bots":%t,"time":"%s"}`,
			status["status"], status["bots"], status["time"])
		w.Write([]byte(response))
	})

	// API routes
	r.Mount("/", apiHandler.Routes())

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &TestApplication{
		server:    server,
		adminBot:  adminBot,
		driverBot: driverBot,
		service:   svc,
		database:  database,
		ctx:       ctx,
		cancel:    cancel,
		hasBots:   hasBots,
	}, nil
}

func (app *TestApplication) Start() error {
	log.Println("Starting test application...")

	// Start bots in goroutines if they exist
	if app.hasBots {
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
	} else {
		log.Println("Running without Telegram bots")
	}

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

func (app *TestApplication) Stop() error {
	log.Println("Stopping test application...")

	// Cancel context to signal shutdown
	app.cancel()

	// Stop bots if they exist
	if app.hasBots {
		if app.adminBot != nil {
			app.adminBot.Stop()
		}
		if app.driverBot != nil {
			app.driverBot.Stop()
		}
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

	log.Println("Test application stopped")
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create application
	app, err := NewTestApplication()
	if err != nil {
		log.Fatalf("Failed to create test application: %v", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start test application: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal...")

	// Stop application
	if err := app.Stop(); err != nil {
		log.Printf("Error stopping test application: %v", err)
	}
}
