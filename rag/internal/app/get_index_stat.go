package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "rag/pkg/rag"
)

// GetIndexStats получает статистику индекса
func (s *RagServer) GetIndexStats(ctx context.Context, req *pb.GetIndexStatsRequest) (*pb.GetIndexStatsResponse, error) {
	var totalDocs, totalChunks int
	var lastUpdated, indexSize string

	err := s.db.Pool().QueryRow(ctx, `
		SELECT 
			COUNT(DISTINCT metadata->>'doc_id') as total_docs,
			COUNT(*) as total_chunks,
			COALESCE(SUM(char_length(content)), 0) as index_size,
			MAX(updated_at)::text as last_updated
		FROM documents
	`).Scan(&totalDocs, &totalChunks, &indexSize, &lastUpdated)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get index stats: %v", err)
	}

	return &pb.GetIndexStatsResponse{
		TotalDocuments: int32(totalDocs),
		IndexSizeBytes: 0,
		LastUpdated:    lastUpdated,
	}, nil
}
