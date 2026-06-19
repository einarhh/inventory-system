package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/einarhh/inventory/internal/domain"
	"github.com/einarhh/inventory/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type createLocationRequest struct {
	CustomerID string  `json:"customer_id"`
	ParentID   *string `json:"parent_id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Notes      *string `json:"notes"`
}

type locationResponse struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	ParentID   *uuid.UUID `json:"parent_id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Notes      *string    `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func toLocationResponse(l store.Location) locationResponse {
	return locationResponse{
		ID:         l.ID,
		CustomerID: l.CustomerID,
		ParentID:   l.ParentID,
		Type:       string(l.Type),
		Name:       l.Name,
		Notes:      l.Notes,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request) {
	var req createLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "customer_id must be a valid UUID")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "parent_id must be a valid UUID")
			return
		}
		parentID = &pid
	}

	in := domain.CreateLocation{
		Name:  req.Name,
		Type:  domain.LocationType(req.Type),
		Notes: req.Notes,
	}
	if err := in.Validate(); err != nil {
		var verr domain.ValidationError
		if errors.As(err, &verr) {
			writeError(w, http.StatusBadRequest, verr.Msg)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// NOTE: this enforces that parent_id references an existing location (via the
	// FK) but not yet that the parent belongs to the same customer. That check
	// lands with the location read/update endpoints in the next increment.
	loc, err := s.store.CreateLocation(r.Context(), store.CreateLocationParams{
		CustomerID: customerID,
		ParentID:   parentID,
		Type:       store.LocationType(in.Type),
		Name:       in.Name,
		Notes:      in.Notes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			writeError(w, http.StatusBadRequest, "customer_id or parent_id does not reference an existing row")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create location")
		return
	}

	writeJSON(w, http.StatusCreated, toLocationResponse(loc))
}

func (s *Server) handleListLocations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	customerID, err := uuid.Parse(q.Get("customer_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "customer_id query parameter is required and must be a valid UUID")
		return
	}

	limit, err := parseLimit(q.Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	offset, err := parseOffset(q.Get("offset"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	locs, err := s.store.ListLocations(r.Context(), store.ListLocationsParams{
		CustomerID: customerID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list locations")
		return
	}

	out := make([]locationResponse, 0, len(locs))
	for _, l := range locs {
		out = append(out, toLocationResponse(l))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"locations": out,
		"limit":     limit,
		"offset":    offset,
	})
}

func parseLimit(s string) (int32, error) {
	if s == "" {
		return defaultLimit, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, errors.New("limit must be a positive integer")
	}
	if n > maxLimit {
		n = maxLimit
	}
	return int32(n), nil
}

func parseOffset(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, errors.New("offset must be a non-negative integer")
	}
	return int32(n), nil
}
