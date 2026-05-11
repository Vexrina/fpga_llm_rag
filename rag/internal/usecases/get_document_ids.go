package usecases

import (
	"context"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
	pb "rag/pkg/rag"
)

type GetDocumentIdsRepository interface {
	SearchSimilar(ctx context.Context, tx pgx.Tx, queryEmbedding []float32, limit int, threshold float32, method utils.ComparisonMethod) ([]repository.SearchResult, error)
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type GetDocumentIdsSettingsProvider interface {
	GetComparisonMethod(ctx context.Context) (utils.ComparisonMethod, error)
}

type GetDocumentIdsUsecase struct {
	repository        GetDocumentIdsRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	settingsProvider  GetDocumentIdsSettingsProvider
}

func NewGetDocumentIdsUsecase(
	repository GetDocumentIdsRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
	settingsProvider GetDocumentIdsSettingsProvider,
) *GetDocumentIdsUsecase {
	return &GetDocumentIdsUsecase{
		repository:        repository,
		floatWeaverClient: floatWeaverClient,
		settingsProvider:  settingsProvider,
	}
}

func (u *GetDocumentIdsUsecase) GetDocumentIds(ctx context.Context, domain *utils.GetDocumentIdsDomain) (*pb.GetDocumentIdsResponse, error) {
	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: domain.Query})
	if err != nil {
		return nil, err
	}
	if len(embed.Embeddings) == 0 {
		return nil, err
	}
	queryEmbedding := embed.Embeddings[0].Values

	method := domain.ComparisonMethod
	if method == utils.ComparisonMethodUnspecified && u.settingsProvider != nil {
		method, _ = u.settingsProvider.GetComparisonMethod(ctx)
	}
	if method == utils.ComparisonMethodUnspecified {
		method = utils.ComparisonMethodCosine
	}

	var ids []*pb.DocumentIdEntry
	var totalFound int32

	err = u.repository.WithTransactional(ctx, func(tx pgx.Tx) error {
		items, err := u.repository.SearchSimilar(ctx, tx, queryEmbedding, int(domain.Limit), domain.SimilarityThs, method)
		if err != nil {
			return err
		}
		totalFound = int32(len(items))
		for _, item := range items {
			ids = append(ids, &pb.DocumentIdEntry{
				Id:              item.DocID,
				SimilarityScore: item.Score,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetDocumentIdsResponse{
		DocumentIds: ids,
		TotalFound:  totalFound,
	}, nil
}
