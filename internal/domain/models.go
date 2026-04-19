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
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	Country    string
	Budget     int64
	TotalValue int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Player struct {
	ID        uuid.UUID
	TeamID    uuid.UUID
	FirstName string
	LastName  string
	Country   string
	Position  string
	Age       int
	Value     int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transfer struct {
	ID              uuid.UUID
	PlayerID        uuid.UUID
	AskingPrice     int64
	ListedAt        time.Time
	PlayerFirstName string
	PlayerLastName  string
	PlayerPosition  string
	PlayerValue     int64
	TeamName        string
}
