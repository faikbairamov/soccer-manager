package service

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransferService struct {
	store repository.Storer
}

func NewTransferService(store repository.Storer) *TransferService {
	return &TransferService{store: store}
}
func (s *TransferService) ListTransfer(ctx context.Context, userID, playerID uuid.UUID, askingPrice int64) (domain.Transfer, error) {
	if askingPrice <= 0 {
		return domain.Transfer{}, domain.ErrInvalidInput
	}
	player, err := s.store.GetPlayerByID(ctx, pgtype.UUID{Bytes: playerID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transfer{}, domain.ErrNotFound
		}
		return domain.Transfer{}, err
	}
	team, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return domain.Transfer{}, err
	}
	if player.TeamID.Bytes != team.ID.Bytes {
		return domain.Transfer{}, domain.ErrForbidden
	}
	_, err = s.store.GetTransferByPlayerID(ctx, player.ID)
	if err == nil {
		return domain.Transfer{}, domain.ErrConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Transfer{}, err
	}
	transfer, err := s.store.CreateTransfer(ctx, repository.CreateTransferParams{
		PlayerID:    player.ID,
		AskingPrice: askingPrice,
	})
	if err != nil {
		return domain.Transfer{}, err
	}
	return domain.Transfer{
		ID:              transfer.ID.Bytes,
		PlayerID:        transfer.PlayerID.Bytes,
		AskingPrice:     transfer.AskingPrice,
		ListedAt:        transfer.ListedAt.Time,
		PlayerFirstName: player.FirstName,
		PlayerLastName:  player.LastName,
		PlayerPosition:  string(player.Position),
		PlayerValue:     player.Value,
		TeamName:        team.Name,
	}, nil
}
func (s *TransferService) GetTransfers(ctx context.Context, page, limit int) ([]domain.Transfer, int64, error) {
	offset := (page - 1) * limit
	rows, err := s.store.ListTransfers(ctx, repository.ListTransfersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.store.CountTransfers(ctx)
	if err != nil {
		return nil, 0, err
	}
	transfers := make([]domain.Transfer, len(rows))
	for i, r := range rows {
		transfers[i] = domain.Transfer{
			ID:              r.ID.Bytes,
			PlayerID:        r.PlayerID.Bytes,
			AskingPrice:     r.AskingPrice,
			ListedAt:        r.ListedAt.Time,
			PlayerFirstName: r.FirstName,
			PlayerLastName:  r.LastName,
			PlayerPosition:  string(r.Position),
			PlayerValue:     r.Value,
			TeamName:        r.TeamName,
		}
	}
	return transfers, total, nil
}
func (s *TransferService) DelistTransfer(ctx context.Context, userID, transferID uuid.UUID) error {
	transfer, err := s.store.GetTransferByID(ctx, pgtype.UUID{Bytes: transferID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	player, err := s.store.GetPlayerByID(ctx, transfer.PlayerID)
	if err != nil {
		return err
	}
	team, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return err
	}
	if player.TeamID.Bytes != team.ID.Bytes {
		return domain.ErrForbidden
	}
	return s.store.DeleteTransfer(ctx, transfer.ID)
}
func (s *TransferService) BuyPlayer(ctx context.Context, userID, transferID uuid.UUID) error {
	transfer, err := s.store.GetTransferByID(ctx, pgtype.UUID{Bytes: transferID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	player, err := s.store.GetPlayerByID(ctx, transfer.PlayerID)
	if err != nil {
		return err
	}
	buyerTeam, err := s.store.GetTeamByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return err
	}
	if buyerTeam.ID.Bytes == player.TeamID.Bytes {
		return domain.ErrOwnPlayer
	}
	if buyerTeam.Budget < transfer.AskingPrice {
		return domain.ErrInsufficientFunds
	}
	return s.store.WithTx(ctx, func(q repository.Querier) error {
		if _, err := q.UpdateTeamBudget(ctx, repository.UpdateTeamBudgetParams{
			ID:     player.TeamID,
			Budget: transfer.AskingPrice,
		}); err != nil {
			return err
		}
		if _, err := q.UpdateTeamBudget(ctx, repository.UpdateTeamBudgetParams{
			ID:     buyerTeam.ID,
			Budget: -transfer.AskingPrice,
		}); err != nil {
			return err
		}
		multiplier := 1.10 + rand.Float64()*0.90
		newValue := int64(float64(player.Value) * multiplier)
		if _, err := q.TransferPlayer(ctx, repository.TransferPlayerParams{
			ID:     player.ID,
			TeamID: buyerTeam.ID,
			Value:  newValue,
		}); err != nil {
			return err
		}
		return q.DeleteTransfer(ctx, transfer.ID)
	})
}
