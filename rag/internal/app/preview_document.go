package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

func (s *RagServer) PreviewDocument(ctx context.Context, req *pb.PreviewDocumentRequest) (*pb.PreviewDocumentResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}
	if req.GetSourceType() == pb.DocumentSourceType_SOURCE_TYPE_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "source_type is required")
	}

	domain := utils.PreviewDocumentFromGRPCToDomain(req)
	result, err := s.previewDocumentUsecase.Preview(ctx, domain)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "previewDocumentUsecase.Preview error: %s", err)
	}

	return utils.PreviewResultToDomainToGRPC(result), nil
}
