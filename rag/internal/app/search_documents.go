package app

import (
	"context"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

// SearchDocuments выполняет поиск документов
func (s *RagServer) SearchDocuments(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if err := s.validateSearch(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %s", err)
	}
	domainReq := utils.SearchDocumentFromGRPCToDomain(req)
	resp, err := s.searchDocumentUsecase.SearchDocuments(ctx, domainReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска документов: %w", err)
	}
	return resp, nil
}

func (s *RagServer) validateSearch(req *pb.SearchRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Query, validation.Required),
		validation.Field(&req.Limit, validation.Required, validation.Min(1), validation.Max(100)),
		validation.Field(&req.SimilarityThreshold, validation.Required, validation.Min(0.0), validation.Max(1.0)),
	)
}
