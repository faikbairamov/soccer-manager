CREATE TABLE transfer_list (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id    UUID NOT NULL UNIQUE REFERENCES players(id) ON DELETE CASCADE,
    asking_price BIGINT NOT NULL,
    listed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transfer_list_player_id ON transfer_list(player_id);

