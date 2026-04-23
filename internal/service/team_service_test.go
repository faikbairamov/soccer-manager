package service

import (
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetTeam_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTeamService(store)

	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Name: "Test FC", Country: "Spain", Budget: 5_000_000}, nil)
	mq.EXPECT().GetPlayersByTeamID(gomock.Any(), pgID(teamUUID)).
		Return([]repository.Player{{Value: 1_000_000}, {Value: 2_000_000}}, nil)

	result, err := svc.GetTeam(t.Context(), userUUID)
	require.NoError(t, err)
	assert.Equal(t, "Test FC", result.Name)
	assert.Equal(t, int64(3_000_000), result.TotalValue)
}

func TestGetTeam_NotFound(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTeamService(store)

	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).Return(repository.Team{}, pgx.ErrNoRows)

	_, err := svc.GetTeam(t.Context(), userUUID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetTeam_TotalValueIsSum(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTeamService(store)

	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().GetPlayersByTeamID(gomock.Any(), pgID(teamUUID)).
		Return([]repository.Player{{Value: 500_000}, {Value: 1_500_000}, {Value: 750_000}}, nil)

	result, err := svc.GetTeam(t.Context(), userUUID)
	require.NoError(t, err)
	assert.Equal(t, int64(2_750_000), result.TotalValue)
}

func TestUpdateTeam_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewTeamService(store)

	mq.EXPECT().GetTeamByUserID(gomock.Any(), pgID(userUUID)).
		Return(repository.Team{ID: pgID(teamUUID), Name: "New Name", Country: "France", Budget: 5_000_000}, nil).Times(2)
	mq.EXPECT().UpdateTeam(gomock.Any(), repository.UpdateTeamParams{
		ID: pgID(teamUUID), Name: "New Name", Country: "France",
	}).Return(repository.Team{}, nil)
	mq.EXPECT().GetPlayersByTeamID(gomock.Any(), pgID(teamUUID)).
		Return([]repository.Player{{Value: 1_000_000}}, nil)

	result, err := svc.UpdateTeam(t.Context(), userUUID, "New Name", "France")
	require.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
}

func TestUpdateTeam_EmptyName(t *testing.T) {
	store, _ := newMockStore(t)
	svc := NewTeamService(store)

	_, err := svc.UpdateTeam(t.Context(), userUUID, "   ", "France")
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestUpdateTeam_EmptyCountry(t *testing.T) {
	store, _ := newMockStore(t)
	svc := NewTeamService(store)

	_, err := svc.UpdateTeam(t.Context(), userUUID, "Good Name", "")
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}
