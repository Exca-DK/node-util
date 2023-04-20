-- name: GetNodeScanRecords :one
SELECT * FROM "scan_records"
WHERE "node" = $1;

-- name: CreateScanRecord :one
INSERT INTO "scan_records" (
  "node", "ip", "ports"
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: DeleteScanRecord :exec
DELETE FROM "scan_records"
WHERE "id" = $1;

-- name: DeleteNodeScanRecords :exec
DELETE FROM "scan_records"
WHERE "node" = $1;