package repository

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"

	"rag/generated/rag_db/public/table"
)

func (r *VecDb) DeleteItem(ctx context.Context, tx pgx.Tx, id int) error {
	deleteQuery := table.Documents.
		DELETE().
		WHERE(table.Documents.ID.EQ(postgres.Int(int64(id))))

	sql, args := deleteQuery.Sql()

	_, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}
