package main

import (
	"log"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"postgres-server/internal/config"
	"postgres-server/internal/db"
	"postgres-server/internal/handler"
	"postgres-server/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env (if present)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment")
	}

	// 2. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// 3. Connect to Postgres
	sqlDB, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer sqlDB.Close()

	// 4. Run migrations
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("migrate driver init: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("migrate initialization failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrations apply failed: %v", err)
	}
	log.Println("database migrations applied successfully")

	// 5. Prepare HTTP server
	mux := http.NewServeMux()

	// Allowed tables/fields for generic queries
	allowed := map[string][]string{
		"items":  {"id", "name", "category", "price_ugx", "available"},
		"users":  {"id", "email", "role"},
		"orders": {"id", "user_id", "status", "total_cost"},
	}
	// Public query endpoint
	mux.HandleFunc("/query", handler.MakeQueryHandler(sqlDB, allowed))

	// Admin sub-router (protected by JWT)
	adminMux := http.NewServeMux()
	adminMux.Handle("/admin/models/register",
		middleware.RequireJWT(handler.RegisterModelHandler(sqlDB)))
	adminMux.Handle("/admin/models/get",
		middleware.RequireJWT(handler.ReadModelHandler(sqlDB))) // GET
	adminMux.Handle("/admin/models/list",
		middleware.RequireJWT(handler.ListModelsHandler(sqlDB)))
	adminMux.Handle("/admin/models/delete",
		middleware.RequireJWT(handler.DeleteModelHandler(sqlDB))) // DELETE

	mux.Handle("/admin/", adminMux)

	// 6. Start HTTP server with timeouts
	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("starting MCP server on %s", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
