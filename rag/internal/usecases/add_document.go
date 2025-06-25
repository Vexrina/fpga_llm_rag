package usecases

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
)

//go:generate mockgen -source=rag/internal/usecases/add_document.go -destination=rag/internal/usecases/mocks/mock_add_document_repository.go -package=mocks AddDocumentRepository
type AddDocumentRepository interface {
	InsertItemWithTx(ctx context.Context, tx pgx.Tx, item repository.Item) error
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type AddDocumentUsecase struct {
	addDocRepository  AddDocumentRepository
	floatWeaverClient floatweaver.EmbedServiceClient
}

func NewAddDocumentUsecase(
	repository AddDocumentRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
) *AddDocumentUsecase {
	return &AddDocumentUsecase{
		addDocRepository:  repository,
		floatWeaverClient: floatWeaverClient,
	}

}

func (u *AddDocumentUsecase) AddDocument(ctx context.Context, domain *utils.AddDocumentDomain) error {
	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: domain.Content})
	if err != nil {
		return fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
	}
	return u.addDocRepository.WithTransactional(ctx, func(tx pgx.Tx) error {
		item := repository.Item{
			Title:     domain.Title,
			Embedding: embed.Embeddings[0].Values,
			Text:      domain.Content,
			Metadata:  domain.Metadata,
		}
		return u.addDocRepository.InsertItemWithTx(ctx, tx, item)
	})
}
