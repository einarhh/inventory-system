-- name: CreateLocation :one
INSERT INTO locations (customer_id, parent_id, type, name, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListLocations :many
SELECT * FROM locations
WHERE customer_id = $1
ORDER BY created_at, id
LIMIT $2 OFFSET $3;
