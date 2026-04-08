package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type SearchResult struct {
	DocID     string
	Title     string
	Text      string
	Metadata  map[string]string
	Score     float32
	ChunkInfo string
}

func (r *VecDb) SearchSimilar(ctx context.Context, tx pgx.Tx, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	vectorValue := VectorFromFloat32(queryEmbedding)

	query := `
		SELECT 
			metadata->>'doc_id' as doc_id,
			title,
			content,
			metadata,
			(embedding <-> $1)::float8 as distance
		FROM documents
		WHERE embedding IS NOT NULL AND metadata ? 'doc_id'
		ORDER BY embedding <-> $1
		LIMIT $2
	`

	rows, err := tx.Query(ctx, query, vectorValue, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar items: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var distance float64
		readErr := rows.Scan(&result.DocID, &result.Title, &result.Text, &result.Metadata, &distance)
		if readErr != nil {
			return nil, fmt.Errorf("не удалось считать строку: %w", readErr)
		}
		result.Score = float32(1 - distance)
		if idx, ok := result.Metadata["chunk_index"]; ok {
			total := result.Metadata["chunk_total"]
			result.ChunkInfo = fmt.Sprintf("часть %s/%s", idx, total)
		}
		results = append(results, result)
	}

	return results, nil
}
