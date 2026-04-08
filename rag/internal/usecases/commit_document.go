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

const maxTokens = 200

func chunkText(text string) []string {
	words := strings.Fields(text)
	var chunks []string
	var currentChunk []string
	currentLen := 0

	for _, word := range words {
		if currentLen+len(word)+1 > maxTokens*4 && len(currentChunk) > 0 {
			chunks = append(chunks, strings.Join(currentChunk, " "))
			currentChunk = nil
			currentLen = 0
		}
		currentChunk = append(currentChunk, word)
		currentLen += len(word) + 1
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}
	return chunks
}

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

	chunks := chunkText(domain.Content)
	docID := uuid.New().String()

	err := u.repo.WithTransactional(ctx, func(tx pgx.Tx) error {
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

	return docID, nil
}
