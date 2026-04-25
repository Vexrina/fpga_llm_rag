package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Item struct {
	ID        int
	Title     string
	Embedding []float32
	Text      string
	Metadata  map[string]string
}

type VecDb struct {
	conn *pgxpool.Pool
}

type VectorRepository interface {
	InsertItem(ctx context.Context, item Item) error
	SearchSimilar(ctx context.Context, queryEmbedding []float32, limit int) ([]Item, error)
	DeleteItem(ctx context.Context, id int) error
}

type DocumentVersionRepository interface {
	GetDocumentVersions(ctx context.Context, documentID string, limit int) ([]DocumentVersion, error)
	GetDocumentVersion(ctx context.Context, versionID int) (*DocumentVersion, error)
	RollbackToVersion(ctx context.Context, documentID string, versionID int, rollbackBy string) (int, error)
	GetAllDocuments(ctx context.Context) ([]AllDocumentItem, error)
}
