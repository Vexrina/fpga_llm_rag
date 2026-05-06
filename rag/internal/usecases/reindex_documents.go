package usecases

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"rag/internal/repository"
	"rag/pkg/floatweaver/floatweaver"
)

type ReindexDocumentsUsecase struct {
	documentIndexRepo repository.DocumentIndexRepository
	db                *repository.VecDb
	settingsRepo      SettingsRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	chunkSize         int
	chunkOverlap      int
	embeddingModel    string
	mu                sync.Mutex
	isReindexing      bool
}

func NewReindexDocumentsUsecase(
	documentIndexRepo repository.DocumentIndexRepository,
	db *repository.VecDb,
	settingsRepo SettingsRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
) *ReindexDocumentsUsecase {
	return &ReindexDocumentsUsecase{
		documentIndexRepo: documentIndexRepo,
		db:                db,
		settingsRepo:      settingsRepo,
		floatWeaverClient: floatWeaverClient,
	}
}

func (u *ReindexDocumentsUsecase) StartReindex(ctx context.Context) error {
	u.mu.Lock()
	if u.isReindexing {
		u.mu.Unlock()
		return fmt.Errorf("reindexing already in progress")
	}
	u.isReindexing = true
	u.mu.Unlock()

	fmt.Println("Starting reindex job...")
	go func() {
		if err := u.reindexAllDocuments(context.Background()); err != nil {
			fmt.Printf("reindex error: %v\n", err)
		}
		u.mu.Lock()
		u.isReindexing = false
		u.mu.Unlock()
		fmt.Println("Reindex job finished")
	}()

	return nil
}

func (u *ReindexDocumentsUsecase) IsReindexing() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.isReindexing
}

func (u *ReindexDocumentsUsecase) reindexAllDocuments(ctx context.Context) error {
	fmt.Println("Starting reindex of all documents...")
	docs, err := u.documentIndexRepo.GetDocumentsToReindex(ctx)
	if err != nil {
		return fmt.Errorf("failed to get documents to reindex: %w", err)
	}

	if len(docs) == 0 {
		fmt.Println("No documents to reindex")
		return nil
	}

	fmt.Printf("Found %d documents to reindex\n", len(docs))

	chunkSize, err := u.getChunkSize(ctx)
	if err != nil || chunkSize == 0 {
		chunkSize = 200
	}
	u.chunkSize = chunkSize

	chunkOverlap, err := u.getChunkOverlap(ctx)
	if err != nil || chunkOverlap < 0 {
		chunkOverlap = 0
	}
	u.chunkOverlap = chunkOverlap

	embeddingModel, err := u.getEmbeddingModel(ctx)
	if err != nil || embeddingModel == "" {
		embeddingModel = "mxbai-embed-large"
	}
	u.embeddingModel = embeddingModel

	fmt.Printf("Reindexing with chunkSize=%d, chunkOverlap=%d, embeddingModel=%s\n", chunkSize, chunkOverlap, embeddingModel)

	for _, doc := range docs {
		if err := u.reindexDocument(ctx, doc); err != nil {
			fmt.Printf("Failed to reindex document %s: %v\n", doc.DocID, err)
			errMsg := err.Error()
			u.documentIndexRepo.UpdateIndexingStatus(ctx, doc.DocID, true, true, &errMsg)
			continue
		}
	}

	return nil
}

func (u *ReindexDocumentsUsecase) reindexDocument(ctx context.Context, doc repository.DocumentIndex) error {
	if err := u.documentIndexRepo.UpdateIndexingStatus(ctx, doc.DocID, true, false, nil); err != nil {
		return fmt.Errorf("failed to update indexing status: %w", err)
	}

	allDocs, err := u.db.GetAllDocumentsRaw(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all documents: %w", err)
	}

	var docContent string
	var docTitle string
	for _, d := range allDocs {
		if d.ID == doc.DocID {
			docTitle = d.Title
			break
		}
	}

	rows, err := u.db.Pool().Query(ctx, `
		SELECT title, content FROM documents 
		WHERE metadata->>'doc_id' = $1
		ORDER BY (metadata->>'chunk_index')::int
	`, doc.DocID)
	if err != nil {
		return fmt.Errorf("failed to get document chunks: %w", err)
	}
	defer rows.Close()

	var contents []string
	for rows.Next() {
		var title, content string
		if err := rows.Scan(&title, &content); err != nil {
			return fmt.Errorf("failed to scan chunk: %w", err)
		}
		contents = append(contents, content)
	}
	docContent = strings.Join(contents, " ")

	if docContent == "" {
		return fmt.Errorf("document content is empty")
	}

	chunks := chunkTextByTokens(docContent, u.chunkSize, u.chunkOverlap)

	tx, err := u.db.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM documents WHERE metadata->>'doc_id' = $1`, doc.DocID)
	if err != nil {
		return fmt.Errorf("failed to delete old chunks: %w", err)
	}

	for i, chunk := range chunks {
		embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: chunk})
		if err != nil {
			return fmt.Errorf("failed to embed chunk: %w", err)
		}

		metadata := map[string]string{
			"doc_id":      doc.DocID,
			"chunk_index": fmt.Sprintf("%d", i),
			"chunk_total": fmt.Sprintf("%d", len(chunks)),
		}

		vectorValue := repository.VectorFromFloat32(embed.Embeddings[0].Values)

		_, err = tx.Exec(ctx, `
			INSERT INTO documents (embedding, title, content, metadata)
			VALUES ($1, $2, $3, $4)
		`, vectorValue, fmt.Sprintf("%s [часть %d/%d]", docTitle, i+1, len(chunks)), chunk, metadata)
		if err != nil {
			return fmt.Errorf("failed to insert chunk: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	embeddingModelPtr := u.embeddingModel
	chunkSizePtr := u.chunkSize
	chunkOverlapPtr := u.chunkOverlap
	if err := u.documentIndexRepo.CreateDocumentIndex(ctx, doc.DocID, docTitle, &embeddingModelPtr, &chunkSizePtr, &chunkOverlapPtr); err != nil {
		return fmt.Errorf("failed to update document index: %w", err)
	}

	if err := u.documentIndexRepo.MarkIndexed(ctx, doc.DocID); err != nil {
		return fmt.Errorf("failed to mark document as indexed: %w", err)
	}

	return nil
}

func (u *ReindexDocumentsUsecase) getChunkSize(ctx context.Context) (int, error) {
	value, err := u.settingsRepo.GetSetting(ctx, "chunkSize")
	if err != nil {
		return 200, nil
	}
	var chunkSize int
	fmt.Sscanf(value, "%d", &chunkSize)
	if chunkSize == 0 {
		chunkSize = 200
	}
	return chunkSize, nil
}

func (u *ReindexDocumentsUsecase) getChunkOverlap(ctx context.Context) (int, error) {
	value, err := u.settingsRepo.GetSetting(ctx, "chunkOverlap")
	if err != nil {
		return 0, nil
	}
	var chunkOverlap int
	fmt.Sscanf(value, "%d", &chunkOverlap)
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	return chunkOverlap, nil
}

func (u *ReindexDocumentsUsecase) getEmbeddingModel(ctx context.Context) (string, error) {
	value, err := u.settingsRepo.GetSetting(ctx, "model")
	if err != nil {
		return "mxbai-embed-large", nil
	}
	if value == "" {
		return "mxbai-embed-large", nil
	}
	return value, nil
}
