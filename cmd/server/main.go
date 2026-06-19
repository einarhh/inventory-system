// Command server is the inventory backend. main.go is wiring only: read config,
// open the database pool, and start the HTTP server. Routes and handlers live in
// internal/http (see AGENTS.md).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	httpapi "github.com/einarhh/inventory/internal/http"
	"github.com/einarhh/inventory/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(store.New(pool), pool),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", addr)
	return srv.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
