package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{Queries: New(pool), pool: pool}
}

func (s *Store) WithTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type Storer interface {
	Querier
	WithTx(ctx context.Context, fn func(q Querier) error) error
}
