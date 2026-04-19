-- name: CreateTeam :one
INSERT INTO teams (user_id, name, country)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeamByUserID :one
SELECT * FROM teams
WHERE user_id = $1;

-- name: UpdateTeam :one
UPDATE teams
SET name = $2, country = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTeamBudget :one
UPDATE teams
SET budget = budget + $2, updated_at = NOW()
WHERE id = $1
RETURNING *;