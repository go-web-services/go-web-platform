package session

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements Session interface for PostgreSQL
type Postgres struct {
	pool   *pgxpool.Pool
	tx     pgx.Tx
	ctx    context.Context
	nested bool // true when participating in an outer transaction; Commit/Rollback are no-ops
}

// NewPostgres creates a new root session for PostgreSQL
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool: pool,
		ctx:  context.Background(),
	}
}

// Begin starts a transaction and returns a session that owns it.
// If ctx already carries a transaction, returns a nested session that
// participates in the outer transaction — Commit and Rollback are no-ops
// for nested sessions; the outer owner controls the transaction lifecycle.
func (s *Postgres) Begin(ctx context.Context) (Session, error) {
	if existingTx := GetTxFromContext(ctx); existingTx != nil {
		return &Postgres{
			pool:   s.pool,
			tx:     existingTx,
			ctx:    ctx,
			nested: true,
		}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		pool: s.pool,
		tx:   tx,
		ctx:  context.WithValue(ctx, dbKey{}, tx),
	}, nil
}

// Transaction executes f within a transaction.
// If ctx already carries a transaction, f runs inside it without starting a
// new one — commit/rollback is left to the outer owner.
// Otherwise a new transaction is started, committed on success, and rolled
// back on error.
func (s *Postgres) Transaction(ctx context.Context, f func(context.Context) error) error {
	if existingTx := GetTxFromContext(ctx); existingTx != nil {
		return f(ctx)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	c := context.WithValue(ctx, dbKey{}, tx)
	if err = f(c); err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction. No-op for nested sessions.
func (s *Postgres) Rollback() error {
	if s.tx == nil || s.nested {
		return nil
	}
	return s.tx.Rollback(s.ctx)
}

// Commit commits the transaction. No-op for nested sessions.
func (s *Postgres) Commit() error {
	if s.tx == nil || s.nested {
		return nil
	}
	return s.tx.Commit(s.ctx)
}

// Context returns the session's context
func (s *Postgres) Context() context.Context {
	return s.ctx
}
