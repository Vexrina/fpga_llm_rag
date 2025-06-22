package app

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

// AddDocument добавляет документ в индекс
func (s *RagServer) AddDocument(ctx context.Context, req *pb.AddDocumentRequest) (*pb.AddDocumentResponse, error) {
	if err := s.validateAdd(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %s", err)
	}
	add := utils.AddDocumentFromGRPCToDomain(req)

	err := s.addDocumentUsecase.AddDocument(ctx, add)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "addDocumentUsecase.AddDocument error: %s", err)
	}

	return &pb.AddDocumentResponse{}, nil
}

func (s *RagServer) validateAdd(req *pb.AddDocumentRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Title, validation.Required),
		validation.Field(&req.Content, validation.Required),
		validation.Field(&req.Metadata, validation.Required),
	)
}
