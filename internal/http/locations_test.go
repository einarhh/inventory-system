package httpapi

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestCreateAndListLocations(t *testing.T) {
	srv := testServer(t)
	customerID := insertCustomer(t, "Create+List Customer")

	// Create a location.
	rec := doJSON(t, srv, http.MethodPost, "/locations", map[string]any{
		"customer_id": customerID.String(),
		"name":        "Garage",
		"type":        "building",
		"notes":       "  detached  ",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body %s", rec.Code, rec.Body.String())
	}

	var created locationResponse
	mustDecode(t, rec, &created)
	if created.ID == uuid.Nil {
		t.Errorf("expected a generated id")
	}
	if created.CustomerID != customerID {
		t.Errorf("customer_id = %v, want %v", created.CustomerID, customerID)
	}
	if created.Name != "Garage" || created.Type != "building" {
		t.Errorf("unexpected location: name=%q type=%q", created.Name, created.Type)
	}
	if created.Notes == nil || *created.Notes != "detached" {
		t.Errorf("notes = %v, want trimmed %q", created.Notes, "detached")
	}

	// List should return exactly the location we just created for this customer.
	rec = doJSON(t, srv, http.MethodGet, "/locations?customer_id="+customerID.String(), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body %s", rec.Code, rec.Body.String())
	}
	var list struct {
		Locations []locationResponse `json:"locations"`
	}
	mustDecode(t, rec, &list)
	if len(list.Locations) != 1 {
		t.Fatalf("expected 1 location, got %d", len(list.Locations))
	}
	if list.Locations[0].ID != created.ID {
		t.Errorf("listed id = %v, want %v", list.Locations[0].ID, created.ID)
	}
}

func TestCreateLocationRejectsBadInput(t *testing.T) {
	srv := testServer(t)
	customerID := insertCustomer(t, "Validation Customer")

	tests := []struct {
		name string
		body map[string]any
		want int
	}{
		{
			name: "blank name",
			body: map[string]any{"customer_id": customerID.String(), "name": "   "},
			want: http.StatusBadRequest,
		},
		{
			name: "unknown type",
			body: map[string]any{"customer_id": customerID.String(), "name": "X", "type": "spaceship"},
			want: http.StatusBadRequest,
		},
		{
			name: "malformed customer_id",
			body: map[string]any{"customer_id": "not-a-uuid", "name": "X"},
			want: http.StatusBadRequest,
		},
		{
			name: "nonexistent customer",
			body: map[string]any{"customer_id": uuid.New().String(), "name": "X"},
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodPost, "/locations", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestListLocationsRequiresCustomerID(t *testing.T) {
	srv := testServer(t)

	rec := doJSON(t, srv, http.MethodGet, "/locations", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
