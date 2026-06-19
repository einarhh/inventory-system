package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealthz(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
	}{
		{name: "db up", pingErr: nil, wantStatus: http.StatusOK},
		{name: "db down", pingErr: errors.New("no connection"), wantStatus: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// healthz does not touch the store, so a nil Store is fine here.
			srv := NewServer(nil, fakePinger{err: tt.pingErr})

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}
