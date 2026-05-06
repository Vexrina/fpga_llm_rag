package repository

import (
	"context"
	"fmt"
)

type DocumentIndex struct {
	DocID              string
	Title              string
	Indexed            bool
	EmbeddingModel     *string
	ChunkSize          *int
	ChunkOverlap       *int
	ChunkTotal         *int
	SizeBytes          *int
	CreatedAt          interface{}
	UpdatedAt          interface{}
	IndexingStartedAt  interface{}
	IndexingFinishedAt interface{}
	IndexingError      *string
}

type DocumentIndexRepository interface {
	CreateDocumentIndex(ctx context.Context, docID, title string, embeddingModel *string, chunkSize, chunkOverlap *int) error
	MarkIndexed(ctx context.Context, docID string) error
	MarkNotIndexed(ctx context.Context, docID string) error
	GetDocumentIndex(ctx context.Context, docID string) (*DocumentIndex, error)
	GetAllDocumentIndexes(ctx context.Context) ([]DocumentIndex, error)
	UpdateIndexingStatus(ctx context.Context, docID string, started bool, finished bool, err *string) error
	GetDocumentsToReindex(ctx context.Context) ([]DocumentIndex, error)
	DeleteDocumentIndex(ctx context.Context, docID string) error
}

func (r *VecDb) CreateDocumentIndex(ctx context.Context, docID, title string, embeddingModel *string, chunkSize, chunkOverlap *int) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO document_index (doc_id, title, indexed, embedding_model, chunk_size, chunk_overlap)
		VALUES ($1, $2, FALSE, $3, $4, $5)
		ON CONFLICT (doc_id) DO UPDATE SET 
			title = EXCLUDED.title,
			embedding_model = EXCLUDED.embedding_model,
			chunk_size = EXCLUDED.chunk_size,
			chunk_overlap = EXCLUDED.chunk_overlap,
			indexed = FALSE,
			updated_at = CURRENT_TIMESTAMP
	`, docID, title, embeddingModel, chunkSize, chunkOverlap)
	if err != nil {
		return fmt.Errorf("failed to create document index: %w", err)
	}
	return nil
}

func (r *VecDb) MarkIndexed(ctx context.Context, docID string) error {
	_, err := r.conn.Exec(ctx, `
		UPDATE document_index 
		SET indexed = TRUE, 
			indexing_finished_at = CURRENT_TIMESTAMP,
			indexing_error = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE doc_id = $1
	`, docID)
	if err != nil {
		return fmt.Errorf("failed to mark document as indexed: %w", err)
	}
	return nil
}

func (r *VecDb) MarkNotIndexed(ctx context.Context, docID string) error {
	_, err := r.conn.Exec(ctx, `
		UPDATE document_index 
		SET indexed = FALSE,
			indexing_started_at = NULL,
			indexing_finished_at = NULL,
			indexing_error = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE doc_id = $1
	`, docID)
	if err != nil {
		return fmt.Errorf("failed to mark document as not indexed: %w", err)
	}
	return nil
}

func (r *VecDb) GetDocumentIndex(ctx context.Context, docID string) (*DocumentIndex, error) {
	var di DocumentIndex
	err := r.conn.QueryRow(ctx, `
		SELECT doc_id, title, indexed, embedding_model, chunk_size, chunk_total, size_bytes,
			created_at, updated_at, indexing_started_at, indexing_finished_at, indexing_error
		FROM document_index
		WHERE doc_id = $1
	`, docID).Scan(
		&di.DocID, &di.Title, &di.Indexed, &di.EmbeddingModel, &di.ChunkSize, &di.ChunkTotal, &di.SizeBytes,
		&di.CreatedAt, &di.UpdatedAt, &di.IndexingStartedAt, &di.IndexingFinishedAt, &di.IndexingError,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document index: %w", err)
	}
	return &di, nil
}

func (r *VecDb) GetAllDocumentIndexes(ctx context.Context) ([]DocumentIndex, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT doc_id, title, indexed, embedding_model, chunk_size, chunk_total, size_bytes,
			created_at, updated_at, indexing_started_at, indexing_finished_at, indexing_error
		FROM document_index
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all document indexes: %w", err)
	}
	defer rows.Close()

	var indexes []DocumentIndex
	for rows.Next() {
		var di DocumentIndex
		err := rows.Scan(
			&di.DocID, &di.Title, &di.Indexed, &di.EmbeddingModel, &di.ChunkSize, &di.ChunkTotal, &di.SizeBytes,
			&di.CreatedAt, &di.UpdatedAt, &di.IndexingStartedAt, &di.IndexingFinishedAt, &di.IndexingError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document index: %w", err)
		}
		indexes = append(indexes, di)
	}
	return indexes, nil
}

func (r *VecDb) UpdateIndexingStatus(ctx context.Context, docID string, started bool, finished bool, err *string) error {
	if started && !finished {
		_, err := r.conn.Exec(ctx, `
			UPDATE document_index 
			SET indexing_started_at = CURRENT_TIMESTAMP,
				indexing_finished_at = NULL,
				indexing_error = NULL,
				updated_at = CURRENT_TIMESTAMP
			WHERE doc_id = $1
		`, docID)
		if err != nil {
			return fmt.Errorf("failed to update indexing started status: %w", err)
		}
	} else if finished {
		_, err := r.conn.Exec(ctx, `
			UPDATE document_index 
			SET indexing_finished_at = CURRENT_TIMESTAMP,
				indexing_error = $2,
				updated_at = CURRENT_TIMESTAMP
			WHERE doc_id = $1
		`, docID, err)
		if err != nil {
			return fmt.Errorf("failed to update indexing finished status: %w", err)
		}
	}
	return nil
}

func (r *VecDb) GetDocumentsToReindex(ctx context.Context) ([]DocumentIndex, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT doc_id, title, indexed, embedding_model, chunk_size, chunk_total, size_bytes,
			created_at, updated_at, indexing_started_at, indexing_finished_at, indexing_error
		FROM document_index
		WHERE indexed = TRUE
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents to reindex: %w", err)
	}
	defer rows.Close()

	var indexes []DocumentIndex
	for rows.Next() {
		var di DocumentIndex
		err := rows.Scan(
			&di.DocID, &di.Title, &di.Indexed, &di.EmbeddingModel, &di.ChunkSize, &di.ChunkTotal, &di.SizeBytes,
			&di.CreatedAt, &di.UpdatedAt, &di.IndexingStartedAt, &di.IndexingFinishedAt, &di.IndexingError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document index: %w", err)
		}
		indexes = append(indexes, di)
	}
	return indexes, nil
}

func (r *VecDb) DeleteDocumentIndex(ctx context.Context, docID string) error {
	_, err := r.conn.Exec(ctx, `DELETE FROM document_index WHERE doc_id = $1`, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document index: %w", err)
	}
	return nil
}
