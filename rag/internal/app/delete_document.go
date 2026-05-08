package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

func (s *RagServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	domain := &utils.DeleteDocumentDomain{
		Id: req.Id,
	}

	return s.documentHistoryUsecase.DeleteDocument(ctx, domain)
}
