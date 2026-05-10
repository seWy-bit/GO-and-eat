package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxAdapter struct {
	pool *pgxpool.Pool
}

func NewTxAdapter(pool *pgxpool.Pool) *TxAdapter {
	return &TxAdapter{pool: pool}
}

func (a *TxAdapter) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

func (a *TxAdapter) Commit(tx pgx.Tx) error {
	return tx.Commit(context.Background())
}

func (a *TxAdapter) Rollback(tx pgx.Tx) error {
	return tx.Rollback(context.Background())
}
