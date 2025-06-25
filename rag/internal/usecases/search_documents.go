package usecases

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
	pb "rag/pkg/rag"
)

//go:generate mockgen -source=rag/internal/usecases/search_documents.go -destination=rag/internal/usecases/mocks/mock_search_documents_repository.go -package=mocks SearchDocumentsRepository

type SearchDocumentsRepository interface {
	SearchSimilar(ctx context.Context, tx pgx.Tx, queryEmbedding []float32, limit int) ([]repository.Item, error)
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type SearchDocumentsUsecase struct {
	repository        SearchDocumentsRepository
	floatWeaverClient floatweaver.EmbedServiceClient
}

func NewSearchDocumentsUsecase(
	repository SearchDocumentsRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
) *SearchDocumentsUsecase {
	return &SearchDocumentsUsecase{
		repository:        repository,
		floatWeaverClient: floatWeaverClient,
	}
}

func (u *SearchDocumentsUsecase) SearchDocuments(ctx context.Context, domain *utils.SearchDocumentDomain) (*pb.SearchResponse, error) {
	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: domain.Query})
	if err != nil {
		return nil, fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
	}
	if len(embed.Embeddings) == 0 {
		return nil, fmt.Errorf("embedding not returned from floatweaver")
	}
	queryEmbedding := embed.Embeddings[0].Values

	var results []*pb.DocumentResult
	var totalFound int32

	err = u.repository.WithTransactional(ctx, func(tx pgx.Tx) error {
		items, err := u.repository.SearchSimilar(ctx, tx, queryEmbedding, int(domain.Limit))
		if err != nil {
			return err
		}
		totalFound = int32(len(items))
		for _, item := range items {
			results = append(results, &pb.DocumentResult{
				Id:       fmt.Sprintf("%d", item.ID),
				Title:    item.Title,
				Content:  item.Text,
				Metadata: item.Metadata,
				// SimilarityScore: // Можно добавить, если считать дистанцию
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.SearchResponse{
		Results:    results,
		TotalFound: totalFound,
	}, nil
}
