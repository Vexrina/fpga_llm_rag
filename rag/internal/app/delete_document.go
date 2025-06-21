package app

import (
	"fmt"
	pb "rag/pkg/rag"
)

// DeleteDocument удаляет документ по ID
func (s *RagServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	// TODO: Реализовать удаление документа по ID
	return &pb.DeleteDocumentResponse{
		Success: false,
		Message: "Метод DeleteDocument не реализован",
	}, fmt.Errorf("метод DeleteDocument не реализован")
}
