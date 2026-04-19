-- name: CreatePlayer :one
INSERT INTO players (team_id, first_name, last_name, country, position, age, value)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = $1;

-- name: GetPlayersByTeamID :many
SELECT * FROM players WHERE team_id = $1 ORDER BY position, last_name;

-- name: UpdatePlayer :one
UPDATE players
SET first_name = $2, last_name = $3, country = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: TransferPlayer :one
UPDATE players
SET team_id = $2, value = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;
