package app

import (
	"fmt"
	pb "rag/pkg/rag"
)

// GetIndexStats получает статистику индекса
func (s *RagServer) GetIndexStats(ctx context.Context, req *pb.GetIndexStatsRequest) (*pb.GetIndexStatsResponse, error) {
	// TODO: Реализовать получение статистики индекса
	return &pb.GetIndexStatsResponse{
		TotalDocuments: 0,
		IndexSizeBytes: 0,
		LastUpdated:    "",
	}, fmt.Errorf("метод GetIndexStats не реализован")
}
