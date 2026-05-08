package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"rag/internal/repository"
	"rag/internal/utils"
	"rag/pkg/floatweaver/floatweaver"
	pb "rag/pkg/rag"
)

type DocumentHistoryUsecase struct {
	repo              repository.DocumentVersionRepository
	documentIndexRepo repository.DocumentIndexRepository
	floatWeaverClient floatweaver.EmbedServiceClient
	getChunkSize      func(ctx context.Context) (int, error)
	getChunkOverlap   func(ctx context.Context) (int, error)
}

func NewDocumentHistoryUsecase(
	repo repository.DocumentVersionRepository,
	documentIndexRepo repository.DocumentIndexRepository,
	floatWeaverClient floatweaver.EmbedServiceClient,
	getChunkSize func(ctx context.Context) (int, error),
	getChunkOverlap func(ctx context.Context) (int, error),
) *DocumentHistoryUsecase {
	return &DocumentHistoryUsecase{
		repo:              repo,
		documentIndexRepo: documentIndexRepo,
		floatWeaverClient: floatWeaverClient,
		getChunkSize:      getChunkSize,
		getChunkOverlap:   getChunkOverlap,
	}
}

type docRepoWithTx interface {
	repository.DocumentVersionRepository
	BeginTx(ctx context.Context) (pgx.Tx, error)
	InsertItemWithTx(ctx context.Context, tx pgx.Tx, item repository.Item) error
}

func (u *DocumentHistoryUsecase) GetDocumentHistory(ctx context.Context, documentID string, limit int) ([]*pb.DocumentVersion, error) {
	if limit <= 0 {
		limit = 20
	}

	versions, err := u.repo.GetDocumentVersions(ctx, documentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get document versions: %w", err)
	}

	result := make([]*pb.DocumentVersion, 0, len(versions))
	for _, v := range versions {
		createdAt := ""
		if v.CreatedAt != nil {
			if t, ok := v.CreatedAt.(time.Time); ok {
				createdAt = t.Format(time.RFC3339)
			}
		}

		result = append(result, &pb.DocumentVersion{
			Id:            int32(v.ID),
			DocumentId:    fmt.Sprintf("%d", v.DocumentID),
			Title:         v.Title,
			Content:       v.Content,
			VersionNumber: int32(v.VersionNumber),
			CreatedAt:     createdAt,
			CreatedBy:     v.CreatedBy,
			Action:        v.Action,
		})
	}
	return result, nil
}

func (u *DocumentHistoryUsecase) RollbackDocument(ctx context.Context, req *utils.RollbackDocumentDomain) (*pb.RollbackDocumentResponse, error) {
	if req.DocumentID == "" {
		return &pb.RollbackDocumentResponse{
			Success: false,
			Message: "document_id is required",
		}, nil
	}
	if req.VersionID <= 0 {
		return &pb.RollbackDocumentResponse{
			Success: false,
			Message: "version_id is required",
		}, nil
	}

	rollbackBy := req.RollbackBy
	if rollbackBy == "" {
		rollbackBy = "admin"
	}

	newVersion, err := u.repo.RollbackToVersion(ctx, req.DocumentID, int(req.VersionID), rollbackBy)
	if err != nil {
		return &pb.RollbackDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("failed to rollback: %v", err),
		}, nil
	}

	return &pb.RollbackDocumentResponse{
		Success:      true,
		Message:      fmt.Sprintf("Successfully rolled back to version %d", req.VersionID),
		NewVersionId: fmt.Sprintf("%d", newVersion),
	}, nil
}

func (u *DocumentHistoryUsecase) GetAllDocuments(ctx context.Context) ([]*pb.DocumentListItem, error) {
	docs, err := u.repo.GetAllDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all documents: %w", err)
	}

	indexes, err := u.documentIndexRepo.GetAllDocumentIndexes(ctx)
	if err != nil {
		indexes = nil
	}

	indexMap := make(map[string]bool)
	for _, idx := range indexes {
		indexMap[idx.DocID] = idx.Indexed
	}

	result := make([]*pb.DocumentListItem, 0, len(docs))
	for _, d := range docs {
		indexed := indexMap[d.ID]
		if !indexed && d.Indexed {
			indexed = true
		}
		result = append(result, &pb.DocumentListItem{
			Id:        d.ID,
			Title:     d.Title,
			UpdatedAt: d.UpdatedAt,
			Indexed:   indexed,
			SizeBytes: d.SizeBytes,
			Chunks:    d.Chunks,
		})
	}
	return result, nil
}

func (u *DocumentHistoryUsecase) UpdateDocument(ctx context.Context, req *utils.UpdateDocumentDomain) (*pb.UpdateDocumentResponse, error) {
	if req.ID == "" {
		return &pb.UpdateDocumentResponse{
			Success: false,
			Message: "id is required",
		}, nil
	}

	updatedBy := req.UpdatedBy
	if updatedBy == "" {
		updatedBy = "admin"
	}

	chunkSize, _ := u.getChunkSize(ctx)
	if chunkSize == 0 {
		chunkSize = 200
	}
	chunkOverlap, _ := u.getChunkOverlap(ctx)
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}

	if req.Content != "" {
		txRepo, ok := u.repo.(docRepoWithTx)
		if !ok {
			return &pb.UpdateDocumentResponse{
				Success: false,
				Message: "repository does not support transactions",
			}, nil
		}
		err := u.updateDocumentContent(ctx, txRepo, req.ID, req.Title, req.Content, chunkSize, chunkOverlap)
		if err != nil {
			return &pb.UpdateDocumentResponse{
				Success: false,
				Message: fmt.Sprintf("failed to update document: %v", err),
			}, nil
		}
	} else if req.Title != "" {
		err := u.repo.UpdateDocumentTitle(ctx, req.ID, req.Title)
		if err != nil {
			return &pb.UpdateDocumentResponse{
				Success: false,
				Message: fmt.Sprintf("failed to update document title: %v", err),
			}, nil
		}
	}

	return &pb.UpdateDocumentResponse{
		Success: true,
		Message: "Document updated successfully",
	}, nil
}

func (u *DocumentHistoryUsecase) updateDocumentContent(ctx context.Context, txRepo docRepoWithTx, docID, title, content string, chunkSize, chunkOverlap int) error {
	chunks := chunkTextByTokens(content, chunkSize, chunkOverlap)

	tx, err := txRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM documents 
		WHERE metadata->>'doc_id' = $1
	`, docID)
	if err != nil {
		return fmt.Errorf("failed to delete old chunks: %w", err)
	}

	metadata := map[string]string{
		"doc_id":      docID,
		"chunk_total": fmt.Sprintf("%d", len(chunks)),
	}

	for i, chunk := range chunks {
		embed, err := u.floatWeaverClient.Embed(ctx, &floatweaver.EmbedRequest{Text: chunk})
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		chunkMeta := make(map[string]string, len(metadata)+2)
		for k, v := range metadata {
			chunkMeta[k] = v
		}
		chunkMeta["chunk_index"] = fmt.Sprintf("%d", i)

		item := repository.Item{
			Title:     fmt.Sprintf("%s [часть %d/%d]", title, i+1, len(chunks)),
			Embedding: embed.Embeddings[0].Values,
			Text:      chunk,
			Metadata:  chunkMeta,
		}
		if err := txRepo.InsertItemWithTx(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to insert chunk: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func (u *DocumentHistoryUsecase) DeleteDocument(ctx context.Context, req *utils.DeleteDocumentDomain) (*pb.DeleteDocumentResponse, error) {
	if req.Id == "" {
		return &pb.DeleteDocumentResponse{
			Success: false,
			Message: "id is required",
		}, nil
	}

	err := u.repo.DeleteDocumentByID(ctx, req.Id)
	if err != nil {
		return &pb.DeleteDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("failed to delete document: %v", err),
		}, nil
	}

	return &pb.DeleteDocumentResponse{
		Success: true,
		Message: "Document deleted successfully",
	}, nil
}
