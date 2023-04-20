-- name: GetEnrRecord :one
SELECT * FROM "enr_records"
WHERE "id" = $1;

-- name: CreateEnrRecord :one
INSERT INTO "enr_records" (
  "id", "fields"
) VALUES (
  $1, $2
) RETURNING *;

-- name: DeleteEnrRecord :exec
DELETE FROM "enr_records"
WHERE "id" = $1;

-- name: MarkEnrSeen :exec
UPDATE "enr_records" 
SET "edited_at" = $1
WHERE "id" = $2;

-- name: UpdateEnrFields :exec
UPDATE "enr_records" 
SET 
  "fields" = $1,
  "edited_at" = now()
WHERE "id" = $2;