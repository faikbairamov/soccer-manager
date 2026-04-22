package service

import (
	"context"
	"errors"
	"strings"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TeamService struct {
	store repository.Storer
}

func NewTeamService(store repository.Storer) *TeamService {
	return &TeamService{store: store}
}

func (s *TeamService) GetTeam(ctx context.Context, userID uuid.UUID) (domain.Team, error) {
	team, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, domain.ErrNotFound
		}
		return domain.Team{}, err
	}

	players, err := s.store.GetPlayersByTeamID(ctx, team.ID)
	if err != nil {
		return domain.Team{}, err
	}
	var total int64
	for _, p := range players {
		total += p.Value
	}
	return domain.Team{
		ID:         team.ID.Bytes,
		UserID:     team.UserID.Bytes,
		Name:       team.Name,
		Country:    team.Country,
		Budget:     team.Budget,
		TotalValue: total,
	}, nil
}

func (s *TeamService) UpdateTeam(ctx context.Context, userID uuid.UUID, name, country string) (domain.Team, error) {
	name = strings.TrimSpace(name)
	country = strings.TrimSpace(country)
	if name == "" || country == "" {
		return domain.Team{}, domain.ErrInvalidInput
	}

	team, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, domain.ErrNotFound
		}
		return domain.Team{}, err
	}
	updated, err := s.store.UpdateTeam(ctx, repository.UpdateTeamParams{
		ID:      team.ID,
		Name:    name,
		Country: country,
	})

	if err != nil {
		return domain.Team{}, err
	}

	return domain.Team{
		ID:      updated.ID.Bytes,
		UserID:  updated.UserID.Bytes,
		Name:    updated.Name,
		Country: updated.Country,
		Budget:  updated.Budget,
	}, nil
}
