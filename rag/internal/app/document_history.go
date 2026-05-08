package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

func (s *RagServer) GetDocumentHistory(ctx context.Context, req *pb.GetDocumentHistoryRequest) (*pb.GetDocumentHistoryResponse, error) {
	if req.DocumentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "document_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	versions, err := s.documentHistoryUsecase.GetDocumentHistory(ctx, req.DocumentId, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get document history: %v", err)
	}

	return &pb.GetDocumentHistoryResponse{Versions: versions}, nil
}

func (s *RagServer) RollbackDocument(ctx context.Context, req *pb.RollbackDocumentRequest) (*pb.RollbackDocumentResponse, error) {
	if req.DocumentId == "" {
		return &pb.RollbackDocumentResponse{
			Success: false,
			Message: "document_id is required",
		}, nil
	}
	if req.VersionId <= 0 {
		return &pb.RollbackDocumentResponse{
			Success: false,
			Message: "version_id is required",
		}, nil
	}

	rollbackBy := req.RollbackBy
	if rollbackBy == "" {
		rollbackBy = "admin"
	}

	domain := &utils.RollbackDocumentDomain{
		DocumentID: req.DocumentId,
		VersionID:  req.VersionId,
		RollbackBy: rollbackBy,
	}

	return s.documentHistoryUsecase.RollbackDocument(ctx, domain)
}

func (s *RagServer) GetAllDocuments(ctx context.Context, req *pb.GetAllDocumentsRequest) (*pb.GetAllDocumentsResponse, error) {
	docs, err := s.documentHistoryUsecase.GetAllDocuments(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get all documents: %v", err)
	}
	return &pb.GetAllDocumentsResponse{Documents: docs}, nil
}

func (s *RagServer) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error) {
	if req.Id == "" {
		return &pb.UpdateDocumentResponse{
			Success: false,
			Message: "id is required",
		}, nil
	}

	updatedBy := req.UpdatedBy
	if updatedBy == "" {
		updatedBy = "admin"
	}

	domain := &utils.UpdateDocumentDomain{
		ID:        req.Id,
		Title:     req.Title,
		Content:   req.Content,
		UpdatedBy: updatedBy,
	}

	result, err := s.documentHistoryUsecase.UpdateDocument(ctx, domain)
	if err != nil {
		return &pb.UpdateDocumentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return result, nil
}
