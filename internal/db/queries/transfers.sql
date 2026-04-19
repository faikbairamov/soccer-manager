-- name: CreateTransfer :one
INSERT INTO transfer_list (player_id, asking_price)
VALUES ($1, $2)
RETURNING *;

-- name: GetTransferByID :one
SELECT * FROM transfer_list WHERE id = $1;

-- name: GetTransferByPlayerID :one
SELECT * FROM transfer_list WHERE player_id = $1;

-- name: ListTransfers :many
SELECT
  tl.id,
  tl.player_id,
  tl.asking_price,
  tl.listed_at,
  p.first_name,
  p.last_name,
  p.position,
  p.value,
  t.name AS team_name
FROM transfer_list tl
JOIN players p ON p.id = tl.player_id
JOIN teams   t ON t.id = p.team_id
ORDER BY tl.listed_at DESC
LIMIT $1 OFFSET $2;

-- name: CountTransfers :one
SELECT COUNT(*) FROM transfer_list;

-- name: DeleteTransfer :exec
DELETE FROM transfer_list WHERE id = $1;