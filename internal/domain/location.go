// Package domain holds the core types and the server-side validation rules.
// It has no knowledge of HTTP or the database — callers parse transport-level
// values (UUIDs, JSON) and the store persists them; domain only decides what is
// a valid value.
package domain

import (
	"fmt"
	"strings"
)

// ValidationError is returned when input fails a business rule. The HTTP layer
// maps it to a 400 with the message surfaced to the client.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

// LocationType mirrors the location_type enum in the database.
type LocationType string

const (
	LocationArea        LocationType = "area"
	LocationBuilding    LocationType = "building"
	LocationFloor       LocationType = "floor"
	LocationRoom        LocationType = "room"
	LocationRack        LocationType = "rack"
	LocationCabinet     LocationType = "cabinet"
	LocationBox         LocationType = "box"
	LocationCompartment LocationType = "compartment"
	LocationVehicle     LocationType = "vehicle"
	LocationOther       LocationType = "other"
)

var validLocationTypes = map[LocationType]bool{
	LocationArea: true, LocationBuilding: true, LocationFloor: true,
	LocationRoom: true, LocationRack: true, LocationCabinet: true,
	LocationBox: true, LocationCompartment: true, LocationVehicle: true,
	LocationOther: true,
}

// Valid reports whether t is one of the known location types.
func (t LocationType) Valid() bool { return validLocationTypes[t] }

// MaxLocationNameLen bounds a location name to keep the column sane.
const MaxLocationNameLen = 200

// CreateLocation is the validated input for creating a location. Validate
// normalises it in place (trimming whitespace, applying the default type), so
// callers should pass the result to the store rather than the raw request.
type CreateLocation struct {
	Name  string
	Type  LocationType
	Notes *string
}

// Validate checks and normalises the input, returning a ValidationError on the
// first problem found.
func (in *CreateLocation) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return ValidationError{"name is required"}
	}
	if len(in.Name) > MaxLocationNameLen {
		return ValidationError{fmt.Sprintf("name must be at most %d characters", MaxLocationNameLen)}
	}

	if in.Type == "" {
		in.Type = LocationOther // matches the schema default
	}
	if !in.Type.Valid() {
		return ValidationError{fmt.Sprintf("invalid location type %q", in.Type)}
	}

	if in.Notes != nil {
		trimmed := strings.TrimSpace(*in.Notes)
		if trimmed == "" {
			in.Notes = nil
		} else {
			in.Notes = &trimmed
		}
	}
	return nil
}
