CREATE TYPE player_position AS ENUM ('goalkeeper', 'defender', 'midfielder', 'attacker');

CREATE TABLE players (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    country    TEXT NOT NULL,
    position   player_position NOT NULL,
    age        INT NOT NULL,
    value      BIGINT NOT NULL DEFAULT 1000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_team_id ON players(team_id);