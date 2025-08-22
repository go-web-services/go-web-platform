package session

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Session aims at facilitating business transactions while abstracting the underlying mechanism
type Session interface {
	// Begin returns a new session with the given context and a started transaction
	Begin(ctx context.Context) (Session, error)

	// Transaction executes a transaction. If the given function returns an error, the transaction
	// is rolled back. Otherwise it is automatically committed before Transaction() returns
	Transaction(ctx context.Context, f func(context.Context) error) error

	// Rollback the changes in the transaction
	Rollback() error

	// Commit the changes in the transaction
	Commit() error

	// Context returns the session's context
	Context() context.Context
}

// dbKey is the key used to store the database transaction in the context
type dbKey struct{}

// GetOrCreateTx returns the pgx.Tx instance stored in the given context
// Returns a new transaction from the fallback pool if no transaction is found in the context
func GetOrCreateTx(ctx context.Context, fallback *pgxpool.Pool) (pgx.Tx, error) {
	db := ctx.Value(dbKey{})
	if db == nil {
		return fallback.Begin(ctx)
	}
	return db.(pgx.Tx), nil
}

// GetTxFromContext returns the transaction stored in the context if one exists, or nil if none exists.
// This function doesn't attempt to create a new transaction when none exists.
func GetTxFromContext(ctx context.Context) pgx.Tx {
	db := ctx.Value(dbKey{})
	if db == nil {
		return nil
	}
	return db.(pgx.Tx)
}
