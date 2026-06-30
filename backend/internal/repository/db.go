package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type Store struct {
	*db.Queries
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, connString string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return &Store{
		Queries: db.New(pool),
		pool:    pool,
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}
