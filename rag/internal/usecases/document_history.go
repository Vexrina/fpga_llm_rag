package usecases

import (
	"context"
	"fmt"
	"time"

	"rag/internal/repository"
	"rag/internal/utils"
	pb "rag/pkg/rag"
)

type DocumentHistoryUsecase struct {
	repo repository.DocumentVersionRepository
}

type DocumentVersionRepository interface {
	GetDocumentVersions(ctx context.Context, documentID string, limit int) ([]repository.DocumentVersion, error)
	GetDocumentVersion(ctx context.Context, versionID int) (*repository.DocumentVersion, error)
	RollbackToVersion(ctx context.Context, documentID string, versionID int, rollbackBy string) (int, error)
	GetAllDocuments(ctx context.Context) ([]repository.AllDocumentItem, error)
}

func NewDocumentHistoryUsecase(repo DocumentVersionRepository) *DocumentHistoryUsecase {
	return &DocumentHistoryUsecase{repo: repo}
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

	result := make([]*pb.DocumentListItem, 0, len(docs))
	for _, d := range docs {
		result = append(result, &pb.DocumentListItem{
			Id:        d.ID,
			Title:     d.Title,
			UpdatedAt: d.UpdatedAt,
			Indexed:   d.Indexed,
			SizeBytes: d.SizeBytes,
			Chunks:    d.Chunks,
		})
	}
	return result, nil
}
