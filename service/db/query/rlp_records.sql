-- name: GetRlpRecord :one
SELECT * FROM "rlp_records"
WHERE "id" = $1;

-- name: CreateRlpRecord :one
INSERT INTO "rlp_records" (
  "id", "raw_name", "protocol_version", "caps"
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: UpdateRlpRecord :exec
UPDATE "rlp_records"
SET 
 raw_name = coalesce(sqlc.narg('raw_name'), raw_name),
 protocol_version = coalesce(sqlc.narg('protocol_version'), protocol_version),
 caps = coalesce(sqlc.narg('caps'), caps),
 edited_at = now()
WHERE id = sqlc.arg('id');

-- name: DeleteRlpRecord :exec
DELETE FROM "rlp_records"
WHERE "id" = $1;
