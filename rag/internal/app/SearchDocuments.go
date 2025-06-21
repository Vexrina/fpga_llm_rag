package app

import (
	"fmt"
	pb "rag/pkg/rag"
)

// SearchDocuments выполняет поиск документов
func (s *RagServer) SearchDocuments(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	// TODO: Реализовать поиск документов
	return &pb.SearchResponse{
		Results:    []*pb.DocumentResult{},
		TotalFound: 0,
	}, fmt.Errorf("метод SearchDocuments не реализован")
}
