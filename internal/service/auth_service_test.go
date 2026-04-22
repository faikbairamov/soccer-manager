package service

import (
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/auth"
	"github.com/faikbairamov/soccer-manager/internal/domain"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegister_DuplicateEmail(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	mq.EXPECT().GetUserByEmail(gomock.Any(), "dup@example.com").
		Return(repository.User{ID: pgID(userUUID)}, nil)

	_, err := svc.Register(t.Context(), "dup@example.com", "password123")
	require.ErrorIs(t, err, domain.ErrConflict)
}

func TestRegister_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	mq.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(repository.User{}, pgx.ErrNoRows)
	mq.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{ID: pgID(userUUID)}, nil)
	mq.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().CreatePlayer(gomock.Any(), gomock.Any()).Return(repository.Player{}, nil).Times(20)

	token, err := svc.Register(t.Context(), "new@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRegister_Creates20Players(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	mq.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(repository.User{}, pgx.ErrNoRows)
	mq.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{ID: pgID(userUUID)}, nil)
	mq.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().CreatePlayer(gomock.Any(), gomock.Any()).Return(repository.Player{}, nil).Times(20)

	_, err := svc.Register(t.Context(), "new@example.com", "password123")
	require.NoError(t, err)
}

func TestRegister_PositionDistribution(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	var captured []repository.CreatePlayerParams
	mq.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(repository.User{}, pgx.ErrNoRows)
	mq.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{ID: pgID(userUUID)}, nil)
	mq.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().CreatePlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, p repository.CreatePlayerParams) (repository.Player, error) {
			captured = append(captured, p)
			return repository.Player{}, nil
		}).Times(20)

	_, err := svc.Register(t.Context(), "new@example.com", "password123")
	require.NoError(t, err)

	counts := map[repository.PlayerPosition]int{}
	for _, p := range captured {
		counts[p.Position]++
	}
	assert.Equal(t, 3, counts[repository.PlayerPositionGoalkeeper], "3 goalkeepers")
	assert.Equal(t, 6, counts[repository.PlayerPositionDefender], "6 defenders")
	assert.Equal(t, 6, counts[repository.PlayerPositionMidfielder], "6 midfielders")
	assert.Equal(t, 5, counts[repository.PlayerPositionAttacker], "5 attackers")
}

func TestRegister_PlayerInitialValues(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	var captured []repository.CreatePlayerParams
	mq.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(repository.User{}, pgx.ErrNoRows)
	mq.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repository.User{ID: pgID(userUUID)}, nil)
	mq.EXPECT().CreateTeam(gomock.Any(), gomock.Any()).Return(repository.Team{ID: pgID(teamUUID)}, nil)
	mq.EXPECT().CreatePlayer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, p repository.CreatePlayerParams) (repository.Player, error) {
			captured = append(captured, p)
			return repository.Player{}, nil
		}).Times(20)

	_, err := svc.Register(t.Context(), "new@example.com", "password123")
	require.NoError(t, err)

	for i, p := range captured {
		assert.Equal(t, int64(1_000_000), p.Value, "player %d must start at $1,000,000", i)
		assert.GreaterOrEqual(t, p.Age, int32(18), "player %d age >= 18", i)
		assert.LessOrEqual(t, p.Age, int32(40), "player %d age <= 40", i)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	mq.EXPECT().GetUserByEmail(gomock.Any(), "ghost@example.com").Return(repository.User{}, pgx.ErrNoRows)

	_, err := svc.Login(t.Context(), "ghost@example.com", "password123")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestLogin_WrongPassword(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	hash, _ := auth.HashPassword("correctpassword")
	mq.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
		Return(repository.User{ID: pgID(userUUID), Password: hash}, nil)

	_, err := svc.Login(t.Context(), "test@example.com", "wrongpassword")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestLogin_Success(t *testing.T) {
	store, mq := newMockStore(t)
	svc := NewAuthService(store, testConfig())

	hash, _ := auth.HashPassword("password123")
	mq.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
		Return(repository.User{ID: pgID(userUUID), Password: hash}, nil)

	token, err := svc.Login(t.Context(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}
