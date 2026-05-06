package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
)

func chunkTextByTokens(text string, maxTokens, overlap int) []string {
	words := strings.Fields(text)
	var chunks []string
	var currentChunk []string
	currentLen := 0

	for _, word := range words {
		if currentLen+len(word)+1 > maxTokens*4 && len(currentChunk) > 0 {
			chunks = append(chunks, strings.Join(currentChunk, " "))

			if overlap > 0 && len(currentChunk) > overlap {
				currentChunk = currentChunk[len(currentChunk)-overlap:]
				currentLen = 0
				for _, w := range currentChunk {
					currentLen += len(w) + 1
				}
			} else {
				currentChunk = nil
				currentLen = 0
			}
		}
		currentChunk = append(currentChunk, word)
		currentLen += len(word) + 1
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}
	return chunks
}

type DocumentIndexRepository interface {
	CreateDocumentIndex(ctx context.Context, docID, title string, embeddingModel *string, chunkSize, chunkOverlap *int) error
	MarkIndexed(ctx context.Context, docID string) error
}

type CommitDocumentUsecase struct {
	repo              AddDocumentRepository
	documentIndexRepo DocumentIndexRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	getEmbeddingModel func(ctx context.Context) (string, error)
	getChunkSize      func(ctx context.Context) (int, error)
	getChunkOverlap   func(ctx context.Context) (int, error)
}

func NewCommitDocumentUsecase(
	repo AddDocumentRepository,
	documentIndexRepo DocumentIndexRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
	getEmbeddingModel func(ctx context.Context) (string, error),
	getChunkSize func(ctx context.Context) (int, error),
	getChunkOverlap func(ctx context.Context) (int, error),
) *CommitDocumentUsecase {
	return &CommitDocumentUsecase{
		repo:              repo,
		documentIndexRepo: documentIndexRepo,
		floatWeaverClient: floatWeaverClient,
		getEmbeddingModel: getEmbeddingModel,
		getChunkSize:      getChunkSize,
		getChunkOverlap:   getChunkOverlap,
	}
}

func (u *CommitDocumentUsecase) Commit(ctx context.Context, domain *utils.CommitDocumentDomain) (string, error) {
	if domain.Content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}
	if domain.Title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}

	embeddingModel, err := u.getEmbeddingModel(ctx)
	if err != nil {
		embeddingModel = "mxbai-embed-large"
	}
	chunkSize, _ := u.getChunkSize(ctx)
	if chunkSize == 0 {
		chunkSize = 200
	}
	chunkOverlap, _ := u.getChunkOverlap(ctx)
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}

	chunks := chunkTextByTokens(domain.Content, chunkSize, chunkOverlap)
	docID := uuid.New().String()

	err = u.repo.WithTransactional(ctx, func(tx pgx.Tx) error {
		for i, chunk := range chunks {
			embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: chunk})
			if err != nil {
				return fmt.Errorf("floatWeaverClient.Embed got error: %w", err)
			}

			metadata := domain.Metadata
			if len(metadata) == 0 {
				metadata = make(map[string]string)
			}
			metadata["doc_id"] = docID
			metadata["chunk_index"] = fmt.Sprintf("%d", i)
			metadata["chunk_total"] = fmt.Sprintf("%d", len(chunks))

			item := repository.Item{
				Title:     fmt.Sprintf("%s [часть %d/%d]", domain.Title, i+1, len(chunks)),
				Embedding: embed.Embeddings[0].Values,
				Text:      chunk,
				Metadata:  metadata,
			}
			if err := u.repo.InsertItemWithTx(ctx, tx, item); err != nil {
				return fmt.Errorf("failed to insert document chunk: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	chunkSizePtr := &chunkSize
	chunkOverlapPtr := &chunkOverlap
	if err := u.documentIndexRepo.CreateDocumentIndex(ctx, docID, domain.Title, &embeddingModel, chunkSizePtr, chunkOverlapPtr); err != nil {
		return "", fmt.Errorf("failed to create document index: %w", err)
	}

	if err := u.documentIndexRepo.MarkIndexed(ctx, docID); err != nil {
		return "", fmt.Errorf("failed to mark document as indexed: %w", err)
	}

	return docID, nil
}
