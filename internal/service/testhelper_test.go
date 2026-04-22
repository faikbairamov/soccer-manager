package service

import (
	"context"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/config"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/faikbairamov/soccer-manager/internal/repository/mock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/mock/gomock"
)

type mockStore struct {
	*mock.MockQuerier
}

func (m *mockStore) WithTx(_ context.Context, fn func(q repository.Querier) error) error {
	return fn(m.MockQuerier)
}

func newMockStore(t *testing.T) (*mockStore, *mock.MockQuerier) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mq := mock.NewMockQuerier(ctrl)
	return &mockStore{mq}, mq
}

func testConfig() *config.Config {
	return &config.Config{JWTSecret: "test-secret", JWTExpiryHours: 1}
}

var (
	userUUID     = uuid.UUID{1}
	teamUUID     = uuid.UUID{2}
	playerUUID   = uuid.UUID{3}
	transferUUID = uuid.UUID{4}
	sellerUUID   = uuid.UUID{5}
)

func pgID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
