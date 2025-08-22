package session

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements Session interface for PostgreSQL
type Postgres struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
	ctx  context.Context
}

// NewPostgres creates a new root session for PostgreSQL
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool: pool,
		ctx:  context.Background(),
	}
}

// Begin starts a new transaction and returns a session with that transaction
func (s *Postgres) Begin(ctx context.Context) (Session, error) {
	// If there's already a transaction in the context, use it
	tx, err := GetOrCreateTx(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		pool: s.pool,
		tx:   tx,
		ctx:  context.WithValue(ctx, dbKey{}, tx),
	}, nil
}

// Transaction executes a function within a transaction
func (s *Postgres) Transaction(ctx context.Context, f func(context.Context) error) error {
	// If there's already a transaction in the context, use it
	tx, err := GetOrCreateTx(ctx, s.pool)
	if err != nil {
		return err
	}

	c := context.WithValue(ctx, dbKey{}, tx)
	err = f(c)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (s *Postgres) Rollback() error {
	if s.tx == nil {
		return nil
	}
	return s.tx.Rollback(s.ctx)
}

// Commit commits the transaction
func (s *Postgres) Commit() error {
	if s.tx == nil {
		return nil
	}
	return s.tx.Commit(s.ctx)
}

// Context returns the session's context
func (s *Postgres) Context() context.Context {
	return s.ctx
}
