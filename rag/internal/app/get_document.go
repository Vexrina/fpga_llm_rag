package app

import (
	"fmt"
	pb "rag/pkg/rag"
)

// GetDocument получает документ по ID
func (s *RagServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	// TODO: Реализовать получение документа по ID
	return &pb.GetDocumentResponse{
		Document: nil,
		Found:    false,
	}, fmt.Errorf("метод GetDocument не реализован")
}
