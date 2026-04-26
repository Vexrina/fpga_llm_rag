package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
	pb "rag/pkg/rag"
)

//go:generate mockgen -source=rag/internal/usecases/search_documents.go -destination=rag/internal/usecases/mocks/mock_search_documents_repository.go -package=mocks SearchDocumentsRepository

type SearchDocumentsRepository interface {
	SearchSimilar(ctx context.Context, tx pgx.Tx, queryEmbedding []float32, limit int, method utils.ComparisonMethod) ([]repository.SearchResult, error)
	WithTransactional(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type SettingsProvider interface {
	GetComparisonMethod(ctx context.Context) (utils.ComparisonMethod, error)
	GetEmbeddingModel(ctx context.Context) (string, error)
}

type QueryLogger interface {
	InsertQueryLog(ctx context.Context, params repository.QueryLogParams) error
}

type SearchDocumentsUsecase struct {
	repository        SearchDocumentsRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	settingsProvider  SettingsProvider
	queryLogger       QueryLogger
}

func NewSearchDocumentsUsecase(
	repository SearchDocumentsRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
	settingsProvider SettingsProvider,
	queryLogger QueryLogger,
) *SearchDocumentsUsecase {
	return &SearchDocumentsUsecase{
		repository:        repository,
		floatWeaverClient: floatWeaverClient,
		settingsProvider:  settingsProvider,
		queryLogger:       queryLogger,
	}
}

func (u *SearchDocumentsUsecase) SearchDocuments(ctx context.Context, domain *utils.SearchDocumentDomain) (*pb.SearchResponse, error) {
	startTime := time.Now()

	embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: domain.Query})
	if err != nil {
		return nil, fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
	}
	if len(embed.Embeddings) == 0 {
		return nil, fmt.Errorf("embedding not returned from floatweaver")
	}
	queryEmbedding := embed.Embeddings[0].Values

	method := domain.ComparisonMethod
	if method == utils.ComparisonMethodUnspecified && u.settingsProvider != nil {
		method, _ = u.settingsProvider.GetComparisonMethod(ctx)
	}
	if method == utils.ComparisonMethodUnspecified {
		method = utils.ComparisonMethodCosine
	}

	var results []*pb.DocumentResult
	var totalFound int32

	err = u.repository.WithTransactional(ctx, func(tx pgx.Tx) error {
		items, err := u.repository.SearchSimilar(ctx, tx, queryEmbedding, int(domain.Limit), method)
		if err != nil {
			return err
		}
		totalFound = int32(len(items))
		for _, item := range items {
			results = append(results, &pb.DocumentResult{
				Id:              item.DocID,
				Title:           item.Title,
				Content:         item.Text,
				Metadata:        item.Metadata,
				SimilarityScore: item.Score,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	responseTimeMs := int(time.Since(startTime).Milliseconds())
	found := totalFound > 0

	embeddingModel := "unknown"
	if u.settingsProvider != nil {
		if model, err := u.settingsProvider.GetEmbeddingModel(ctx); err == nil {
			embeddingModel = model
		}
	}

	if u.queryLogger != nil {
		go func() {
			logCtx := context.Background()
			if err := u.queryLogger.InsertQueryLog(logCtx, repository.QueryLogParams{
				QueryText:      domain.Query,
				EmbeddingModel: embeddingModel,
				ResponseTimeMs: responseTimeMs,
				Found:          found,
				ResultsCount:   int(totalFound),
			}); err != nil {
				log.Printf("ERROR: failed to insert query log: %v", err)
			}
		}()
	}

	return &pb.SearchResponse{
		Results:    results,
		TotalFound: totalFound,
	}, nil
}
