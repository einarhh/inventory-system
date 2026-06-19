package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/einarhh/inventory/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const migrationsDir = "../../db/migrations"

// Shared across the package's DB-backed tests: one Postgres container is started
// once in TestMain and reused. sharedErr is set (and the DB tests skipped) when
// Docker is unavailable, so the non-DB tests in this package still run.
var (
	sharedPool *pgxpool.Pool
	sharedErr  error
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("inventory"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		sharedErr = fmt.Errorf("start postgres container: %w", err)
	} else {
		sharedErr = bringUp(ctx, container)
	}

	code := m.Run()

	if sharedPool != nil {
		sharedPool.Close()
	}
	if container != nil {
		_ = container.Terminate(ctx)
	}
	os.Exit(code)
}

// bringUp runs migrations and opens the shared pool against the container.
func bringUp(ctx context.Context, container *tcpostgres.PostgresContainer) error {
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("connection string: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	sharedPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pgxpool: %w", err)
	}
	return nil
}

// testServer returns a Server wired to the shared test database, skipping the
// test if Docker (and therefore the container) was unavailable.
func testServer(t *testing.T) *Server {
	t.Helper()
	if sharedErr != nil {
		t.Skipf("postgres test container unavailable: %v", sharedErr)
	}
	return NewServer(store.New(sharedPool), sharedPool)
}

// insertCustomer creates a customer row directly (there is no customer endpoint
// yet) and returns its id, so locations have a valid owner to reference.
func insertCustomer(t *testing.T, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := sharedPool.QueryRow(context.Background(),
		"INSERT INTO customers (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		t.Fatalf("insert customer: %v", err)
	}
	return id
}

// doJSON marshals body (if non-nil), performs the request against h, and returns
// the recorder.
func doJSON(t *testing.T, h http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, target, r)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func mustDecode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
}
