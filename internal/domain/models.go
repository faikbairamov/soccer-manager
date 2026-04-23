package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
}

type Team struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	Budget     int64     `json:"budget"`
	TotalValue int64     `json:"total_value"`
	Players    []Player  `json:"players"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Player struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Country   string    `json:"country"`
	Position  string    `json:"position"`
	Age       int       `json:"age"`
	Value     int64     `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transfer struct {
	ID              uuid.UUID `json:"id"`
	PlayerID        uuid.UUID `json:"player_id"`
	AskingPrice     int64     `json:"asking_price"`
	ListedAt        time.Time `json:"listed_at"`
	PlayerFirstName string    `json:"player_first_name"`
	PlayerLastName  string    `json:"player_last_name"`
	PlayerPosition  string    `json:"player_position"`
	PlayerValue     int64     `json:"player_value"`
	TeamName        string    `json:"team_name"`
}
