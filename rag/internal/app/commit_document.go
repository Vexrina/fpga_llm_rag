package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

func (s *RagServer) CommitDocument(ctx context.Context, req *pb.CommitDocumentRequest) (*pb.CommitDocumentResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}
	if req.GetSourceType() == pb.DocumentSourceType_SOURCE_TYPE_PDF {
		if req.GetContentBase64() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "content_base64 is required for PDF")
		}
	} else {
		if req.GetContent() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "content is required")
		}
	}

	domain := utils.CommitDocumentFromGRPCToDomain(req)
	id, err := s.commitDocumentUsecase.Commit(ctx, domain)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "commitDocumentUsecase.Commit error: %s", err)
	}

	return utils.CommitResultToGRPC(id), nil
}
