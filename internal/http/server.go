// Package httpapi wires the HTTP routes to the store. It owns request/response
// shapes and status codes; business validation lives in internal/domain and
// persistence in internal/store.
package httpapi

import (
	"context"
	"net/http"

	"github.com/einarhh/inventory/internal/store"
)

// Store is the slice of the generated *store.Queries that the handlers use.
// Defining it here (rather than importing a concrete type) keeps the HTTP layer
// decoupled and testable.
type Store interface {
	CreateLocation(context.Context, store.CreateLocationParams) (store.Location, error)
	ListLocations(context.Context, store.ListLocationsParams) ([]store.Location, error)
}

// Pinger is what /healthz needs from the database pool.
type Pinger interface {
	Ping(context.Context) error
}

// Server is the application's http.Handler.
type Server struct {
	store Store
	db    Pinger
	mux   *http.ServeMux
}

// NewServer builds the router.
func NewServer(s Store, db Pinger) *Server {
	srv := &Server{store: s, db: db, mux: http.NewServeMux()}
	srv.routes()
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("POST /locations", s.handleCreateLocation)
	s.mux.HandleFunc("GET /locations", s.handleListLocations)
}
