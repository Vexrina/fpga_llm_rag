package app

import (
	"context"
	"fmt"
	pb "rag/pkg/rag"
)

// AddDocument добавляет документ в индекс
func (s *RagServer) AddDocument(ctx context.Context, req *pb.AddDocumentRequest) (*pb.AddDocumentResponse, error) {
	// TODO: Реализовать добавление документа в индекс
	return &pb.AddDocumentResponse{
		Success: false,
		Message: "Метод AddDocument не реализован",
	}, fmt.Errorf("метод AddDocument не реализован")
}
