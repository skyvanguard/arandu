package main

import (
	"context"
	"database/sql"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arandu-ai/arandu/assets"
	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/executor"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/router"
	"github.com/arandu-ai/arandu/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed templates/prompts/*.tmpl
var promptTemplates embed.FS

//go:embed templates/scripts/*.js
var scriptTemplates embed.FS

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	config.Init()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize database with connection pooling
	db, err := initDatabase()
	if err != nil {
		logging.Error("Failed to initialize database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	queries := database.New(db)

	// Run migrations
	if err := runMigrations(db); err != nil {
		logging.Error("Failed to run migrations", "error", err.Error())
		os.Exit(1)
	}

	// Initialize assets
	assets.Init(promptTemplates, scriptTemplates)

	// Initialize Docker client
	if err := executor.InitClient(); err != nil {
		logging.Error("Failed to initialize Docker client", "error", err.Error())
		os.Exit(1)
	}

	// Initialize browser container
	if err := executor.InitBrowser(queries); err != nil {
		logging.Error("Failed to initialize browser container", "error", err.Error())
		os.Exit(1)
	}

	// Setup HTTP server
	port := strconv.Itoa(config.Config.Port)
	r := router.New(queries)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		logging.Info("Server starting", "url", "http://localhost:"+port)
		logging.Info("GraphQL playground available", "url", "http://localhost:"+port+"/playground")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("HTTP server error", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Wait for termination signal
	<-sigChan
	logging.Info("Shutdown signal received")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logging.Error("HTTP server shutdown error", "error", err.Error())
	}

	// Close all WebSocket connections
	websocket.CloseAll()

	// Cleanup Docker resources
	if err := executor.Cleanup(queries); err != nil {
		logging.Error("Error during cleanup", "error", err.Error())
	}

	logging.Info("Shutdown complete")
}

// initDatabase configura la conexión a SQLite con pooling optimizado
func initDatabase() (*sql.DB, error) {
	// SQLite con WAL mode para mejor concurrencia
	dsn := config.Config.DatabaseURL + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=1000000000&_foreign_keys=true"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Configurar connection pool
	// SQLite es single-writer, pero múltiples lectores
	db.SetMaxOpenConns(1) // SQLite solo permite una conexión de escritura
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// runMigrations ejecuta las migraciones de base de datos
func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	logging.Info("Database migrations completed successfully")
	return nil
}
