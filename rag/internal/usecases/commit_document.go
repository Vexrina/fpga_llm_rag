package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
)

type CommitDocumentUsecase struct {
	repo              AddDocumentRepository
	floatWeaverClient floatweaver.EmbedServiceClient
}

func NewCommitDocumentUsecase(
	repo AddDocumentRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
) *CommitDocumentUsecase {
	return &CommitDocumentUsecase{
		repo:              repo,
		floatWeaverClient: floatWeaverClient,
	}
}

func (u *CommitDocumentUsecase) Commit(ctx context.Context, domain *utils.CommitDocumentDomain) (string, error) {
	if domain.Content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}
	if domain.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}

	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: domain.Content})
	if err != nil {
		return "", fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
	}

	docID := uuid.New().String()

	err = u.repo.WithTransactional(ctx, func(tx pgx.Tx) error {
		item := repository.Item{
			Title:     domain.Title,
			Embedding: embed.Embeddings[0].Values,
			Text:      domain.Content,
			Metadata:  domain.Metadata,
		}
		return u.repo.InsertItemWithTx(ctx, tx, item)
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return docID, nil
}
