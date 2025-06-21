package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r *VecDb) SearchSimilar(ctx context.Context, tx pgx.Tx, queryEmbedding []float32, limit int) ([]Item, error) {
	// Используем нашу векторную обертку
	vectorValue := VectorFromFloat32(queryEmbedding)

	// Используем обычный SQL для векторных операций
	query := `
		SELECT id, embedding, content, metadata
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <-> $1
		LIMIT $2
	`

	rows, err := tx.Query(ctx, query, vectorValue, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var (
			item       Item
			vectorScan VectorScan
		)
		readErr := rows.Scan(&item.ID, &vectorScan, &item.Text, &item.Metadata)
		if readErr != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		item.Embedding = VectorToFloat32(vectorScan)
		items = append(items, item)
	}

	return items, nil
}
