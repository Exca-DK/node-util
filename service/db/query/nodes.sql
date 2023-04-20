-- name: GetNode :one
SELECT * FROM "nodes"
WHERE "id" = $1;

-- name: ListNodesByCreation :many
SELECT * FROM "nodes"
ORDER BY "added_at"
LIMIT $1
OFFSET $2;

-- name: ListNodesByActivity :many
SELECT * FROM "nodes"
ORDER BY "seen_at"
LIMIT $1
OFFSET $2;

-- name: CreateNode :one
INSERT INTO "nodes" (
  "id"
) VALUES (
  $1
)
RETURNING *;

-- name: DeleteNode :exec
DELETE FROM "nodes"
WHERE "id" = $1;

-- name: AddQueries :one
UPDATE "nodes" 
SET "queries" = "queries" + $1
WHERE "id" = $2
RETURNING *;

-- name: IncrementQuery :one
UPDATE "nodes" 
SET "queries" = "queries" + 1
WHERE "id" = $1
RETURNING *;