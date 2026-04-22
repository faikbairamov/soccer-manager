package service

import (
	"slices"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// buyPlayerSetup sets up the three pre-transaction lookups that every BuyPlayer test needs.
// sellerID is the team that owns the player being sold.
func buyPlayerSetup(mq interface {
	EXPECT() interface {
		GetTransferByID(any, any) any
		GetPlayerByID(any, any) any
		GetTeamByUserID(any, any) any
	}
}, askingPrice int64, buyerBudget int64, sellerTeamID pgtype.UUID) {
	// Use the mock directly in each test instead — see individual tests below.
}
func TestBuyPlayer_NotFound(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).Return(repository.TransferList{}, pgx.ErrNoRows)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
func TestBuyPlayer_OwnPlayer(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	// buyer owns the player being sold
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: 1_000_000}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(teamUUID)}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 5_000_000}, nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.ErrorIs(t, err, domain.ErrOwnPlayer)
}
func TestBuyPlayer_InsufficientFunds(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: 3_000_000}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(sellerUUID)}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 1_000_000}, nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}
func TestBuyPlayer_Success_AtomicBudgetUpdate(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	const askingPrice = int64(2_000_000)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: askingPrice}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(sellerUUID), Value: 1_000_000}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 5_000_000}, nil)
	var budgetUpdates []repository.UpdateTeamBudgetParams
	mq.EXPECT().UpdateTeamBudget(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, p repository.UpdateTeamBudgetParams) (repository.Team, error) {
			budgetUpdates = append(budgetUpdates, p)
			return repository.Team{}, nil
		}).Times(2)
	mq.EXPECT().TransferPlayer(gomock.Any(), gomock.Any()).Return(repository.Player{}, nil)
	mq.EXPECT().DeleteTransfer(gomock.Any(), pgID(transferUUID)).Return(nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.NoError(t, err)
	sellerCredited := slices.ContainsFunc(budgetUpdates, func(p repository.UpdateTeamBudgetParams) bool {
		return p.ID == pgID(sellerUUID) && p.Budget == askingPrice
	})
	assert.True(t, sellerCredited, "seller must be credited with the asking price in the same transaction")
	buyerDebited := slices.ContainsFunc(budgetUpdates, func(p repository.UpdateTeamBudgetParams) bool {
		return p.ID == pgID(teamUUID) && p.Budget == -askingPrice
	})
	assert.True(t, buyerDebited, "buyer must be debited with the asking price in the same transaction")
}
func TestBuyPlayer_Success_PlayerTransferredToBuyer(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: 1_000_000}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(sellerUUID), Value: 1_000_000}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 5_000_000}, nil)
	mq.EXPECT().UpdateTeamBudget(gomock.Any(), gomock.Any()).Return(repository.Team{}, nil).Times(2)
	var transferParams repository.TransferPlayerParams
	mq.EXPECT().TransferPlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, p repository.TransferPlayerParams) (repository.Player, error) {
			transferParams = p
			return repository.Player{}, nil
		})
	mq.EXPECT().DeleteTransfer(gomock.Any(), pgID(transferUUID)).Return(nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.NoError(t, err)
	assert.Equal(t, pgID(teamUUID), transferParams.TeamID, "player must be transferred to buyer's team")
}
func TestBuyPlayer_Success_PlayerValueIncreased(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	const originalValue = int64(1_000_000)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: 1_000_000}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(sellerUUID), Value: originalValue}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 5_000_000}, nil)
	mq.EXPECT().UpdateTeamBudget(gomock.Any(), gomock.Any()).Return(repository.Team{}, nil).Times(2)
	var newValue int64
	mq.EXPECT().TransferPlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, p repository.TransferPlayerParams) (repository.Player, error) {
			newValue = p.Value
			return repository.Player{}, nil
		})
	mq.EXPECT().DeleteTransfer(gomock.Any(), pgID(transferUUID)).Return(nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, newValue, int64(float64(originalValue)*1.10), "value must increase by at least 10%%")
	assert.LessOrEqual(t, newValue, int64(float64(originalValue)*2.00), "value must increase by at most 100%%")
}
func TestBuyPlayer_Success_ListingDeleted(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID), AskingPrice: 1_000_000}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(sellerUUID), Value: 1_000_000}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Budget: 5_000_000}, nil)
	mq.EXPECT().UpdateTeamBudget(gomock.Any(), gomock.Any()).Return(repository.Team{}, nil).Times(2)
	mq.EXPECT().TransferPlayer(gomock.Any(), gomock.Any()).Return(repository.Player{}, nil)
	// DeleteTransfer must be called exactly once with the correct transfer ID
	mq.EXPECT().DeleteTransfer(gomock.Any(), pgID(transferUUID)).Return(nil)
	err := svc.BuyPlayer(t.Context(), userUUID, transferUUID)
	require.NoError(t, err)
}
func TestListTransfer_ZeroAskingPrice(t *testing.T) {
	store, _ := newMockStore(t)
	svc := NewTransferService(store)
	_, err := svc.ListTransfer(t.Context(), userUUID, playerUUID, 0)
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}
func TestListTransfer_NotOwner(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	otherTeam := pgtype.UUID{Bytes: [16]byte{99}, Valid: true}
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: otherTeam}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)
	_, err := svc.ListTransfer(t.Context(), userUUID, playerUUID, 1_000_000)
	require.ErrorIs(t, err, domain.ErrForbidden)
}
func TestListTransfer_AlreadyListed(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(teamUUID)}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)
	// Player is already on the transfer list
	mq.EXPECT().GetTransferByPlayerID(gomock.Any(), pgID(playerUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID)}, nil)
	_, err := svc.ListTransfer(t.Context(), userUUID, playerUUID, 1_000_000)
	require.ErrorIs(t, err, domain.ErrConflict)
}
func TestDelistTransfer_NotOwner(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTransferService(store)
	otherTeam := pgtype.UUID{Bytes: [16]byte{99}, Valid: true}
	mq.EXPECT().GetTransferByID(gomock.Any(), pgID(transferUUID)).
		Return(repository.TransferList{ID: pgID(transferUUID), PlayerID: pgID(playerUUID)}, nil)
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: otherTeam}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)
	err := svc.DelistTransfer(t.Context(), userUUID, transferUUID)
	require.ErrorIs(t, err, domain.ErrForbidden)
}
