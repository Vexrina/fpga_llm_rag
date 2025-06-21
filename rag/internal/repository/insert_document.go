package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *VecDb) InsertItemWithTx(ctx context.Context, tx pgx.Tx, item Item) error {
	// Используем нашу векторную обертку
	vectorValue := VectorFromFloat32(item.Embedding)

	// Используем обычный SQL для вставки с векторами
	query := `
		INSERT INTO documents (embedding, title, content, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := tx.QueryRow(ctx, query, vectorValue, item.Title, item.Text, item.Metadata).Scan(&item.ID)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	return nil
}
