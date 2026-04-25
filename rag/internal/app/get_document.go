package app

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/repository"
	pb "rag/pkg/rag"
)

// GetDocument получает документ по ID (объединяет все чанки)
func (s *RagServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "document id is required")
	}

	docs, err := s.db.GetDocumentChunks(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get document: %v", err)
	}

	if len(docs) == 0 {
		return &pb.GetDocumentResponse{
			Document: nil,
			Found:    false,
		}, nil
	}

	// Сортируем чанки по индексу и объединяем
	chunkMap := make(map[int]repository.DocumentChunk)
	var maxIndex int
	for _, doc := range docs {
		idx := 0
		if doc.ChunkIndex != nil {
			idx = *doc.ChunkIndex
		}
		chunkMap[idx] = doc
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	var sb strings.Builder
	for i := 0; i <= maxIndex; i++ {
		if chunk, ok := chunkMap[i]; ok {
			sb.WriteString(chunk.Content)
			if i < maxIndex {
				sb.WriteString("\n\n")
			}
		}
	}

	title := docs[0].Title
	if idx := strings.Index(title, " [часть"); idx > 0 {
		title = title[:idx]
	}

	return &pb.GetDocumentResponse{
		Document: &pb.DocumentResult{
			Id:      req.Id,
			Title:   title,
			Content: sb.String(),
		},
		Found: true,
	}, nil
}
