package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

var (
	connPool *pgxpool.Pool
	once     sync.Once
	initErr  error
)

func ConnectToPostgres(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	once.Do(func() {
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			initErr = fmt.Errorf("failed to parse connection string: %v", err)
			return
		}

		config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			return pgxvec.RegisterTypes(ctx, conn)
		}

		connPool, initErr = pgxpool.NewWithConfig(ctx, config)
		if initErr != nil {
			initErr = fmt.Errorf("failed to connect to database: %v", initErr)
			return
		}
	})
	return connPool, initErr
}

// NewVecDb создает новый экземпляр VecDb
func NewVecDb(ctx context.Context, connStr string) *VecDb {
	connection, err := ConnectToPostgres(ctx, connStr)
	if err != nil {
		panic(err)
	}
	return &VecDb{
		conn: connection,
	}
}

func (r *VecDb) WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
