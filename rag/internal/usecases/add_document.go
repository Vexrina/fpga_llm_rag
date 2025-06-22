package usecases

import (
	"context"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
)

type AddDocumentRepository interface {
	InsertItemWithTx(ctx context.Context, tx pgx.Tx, item repository.Item) error
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type AddDocumentUsecase struct {
	addDocRepository AddDocumentRepository
}

func NewAddDocumentUsecase(
	repository AddDocumentRepository,
) *AddDocumentUsecase {
	return &AddDocumentUsecase{
		addDocRepository: repository,
	}

}

func (u *AddDocumentUsecase) AddDocument(ctx context.Context, domain *utils.AddDocumentDomain) error {
	return u.addDocRepository.WithTransactional(ctx, func(tx pgx.Tx) error {
		item := repository.Item{
			Title:     domain.Title,
			Embedding: domain.Embedding,
			Text:      domain.Content,
			Metadata:  domain.Metadata,
		}
		return u.addDocRepository.InsertItemWithTx(ctx, tx, item)
	})
}
