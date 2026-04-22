package service

import (
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetPlayer_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewPlayerService(store)

	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), FirstName: "Luca", Value: 1_000_000}, nil)

	result, err := svc.GetPlayer(t.Context(), playerUUID)
	require.NoError(t, err)
	assert.Equal(t, "Luca", result.FirstName)
}

func TestGetPlayer_NotFound(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewPlayerService(store)

	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).Return(repository.Player{}, pgx.ErrNoRows)

	_, err := svc.GetPlayer(t.Context(), playerUUID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdatePlayer_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewPlayerService(store)

	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: pgID(teamUUID)}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().UpdatePlayer(gomock.Any(), repository.UpdatePlayerParams{
		ID: pgID(playerUUID), FirstName: "Marco", LastName: "Rossi", Country: "Italy",
	}).Return(repository.Player{FirstName: "Marco", LastName: "Rossi", Country: "Italy"}, nil)

	result, err := svc.UpdatePlayer(t.Context(), userUUID, playerUUID, "Marco", "Rossi", "Italy")
	require.NoError(t, err)
	assert.Equal(t, "Marco", result.FirstName)
}

func TestUpdatePlayer_NotOwner(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewPlayerService(store)

	otherTeam := pgtype.UUID{Bytes: [16]byte{99}, Valid: true}
	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).
		Return(repository.Player{ID: pgID(playerUUID), TeamID: otherTeam}, nil)
	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)

	_, err := svc.UpdatePlayer(t.Context(), userUUID, playerUUID, "X", "Y", "Z")
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdatePlayer_PlayerNotFound(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewPlayerService(store)

	mq.EXPECT().GetPlayerByID(gomock.Any(), pgID(playerUUID)).Return(repository.Player{}, pgx.ErrNoRows)

	_, err := svc.UpdatePlayer(t.Context(), userUUID, playerUUID, "X", "Y", "Z")
	require.ErrorIs(t, err, domain.ErrNotFound)
}
