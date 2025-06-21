package app

import (
	"context"
	"rag/internal/repository"
	pb "rag/pkg/rag"
)

// RagServer реализует RagService
type RagServer struct {
	pb.UnimplementedRagServiceServer

	db *repository.VecDb
}

// NewRagServer создает новый экземпляр RagServer
func NewRagServer(ctx context.Context, connStr string) *RagServer {
	return &RagServer{
		db: repository.NewVecDb(ctx, connStr),
	}
}
