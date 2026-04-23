package service

import (
	"context"
	"errors"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PlayerService struct {
	store repository.Storer
}

func NewPlayerService(store repository.Storer) *PlayerService {
	return &PlayerService{store: store}
}

func (s *PlayerService) GetPlayer(ctx context.Context, id uuid.UUID) (domain.Player, error) {
	p, err := s.store.GetPlayerByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Player{}, domain.ErrNotFound
		}
		return domain.Player{}, err
	}
	return toDomainPlayer(p), nil
}

func (s *PlayerService) UpdatePlayer(ctx context.Context, userID, playerID uuid.UUID, firstName, lastName, country string) (domain.Player, error) {
	player, err := s.store.GetPlayerByID(ctx, pgtype.UUID{Bytes: playerID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Player{}, domain.ErrNotFound
		}
		return domain.Player{}, err
	}

	team, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return domain.Player{}, err
	}

	if player.TeamID.Bytes != team.ID.Bytes {
		return domain.Player{}, domain.ErrForbidden
	}

	updated, err := s.store.UpdatePlayer(ctx, repository.UpdatePlayerParams{
		ID:        player.ID,
		FirstName: firstName,
		LastName:  lastName,
		Country:   country,
	})
	if err != nil {
		return domain.Player{}, err
	}
	return toDomainPlayer(updated), nil
}

func toDomainPlayer(p repository.Player) domain.Player {
	return domain.Player{
		ID:        p.ID.Bytes,
		TeamID:    p.TeamID.Bytes,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Country:   p.Country,
		Position:  string(p.Position),
		Age:       int(p.Age),
		Value:     p.Value,
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}
