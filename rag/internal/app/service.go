package app

import (
	"context"

	"rag/internal/repository"
	"rag/internal/utils"
	pb "rag/pkg/rag"
)

// RagServer реализует RagService
type (
	RagServer struct {
		pb.UnimplementedRagServiceServer

		db                    *repository.VecDb
		addDocumentUsecase    AddDocumentUsecase
		getDocumentUsecase    GetDocumentUsecase
		deleteDocumentUsecase DeleteDocumentsUsecase
		searchDocumentUsecase SearchDocumentsUsecase
		getIndexStatUsecase   GetIndexStatUsecase
	}

	AddDocumentUsecase interface {
		AddDocument(ctx context.Context, req *utils.AddDocumentDomain) error
	}
	GetDocumentUsecase interface {
		GetDocument(ctx context.Context, req *utils.GetDocumentDomain) (*pb.GetDocumentResponse, error)
	}
	DeleteDocumentsUsecase interface {
		DeleteDocument(ctx context.Context, req *utils.DeleteDocumentDomain) (*pb.DeleteDocumentResponse, error)
	}
	SearchDocumentsUsecase interface {
		SearchDocuments(ctx context.Context, req *utils.SearchDocumentDomain) (*pb.SearchResponse, error)
	}
	GetIndexStatUsecase interface {
		GetIndexStat(ctx context.Context, req *pb.GetIndexStatsRequest) (*pb.GetIndexStatsResponse, error)
	}
)

// NewRagServer создает новый экземпляр RagServer
func NewRagServer(
	database *repository.VecDb,
	addDocumentUsecase AddDocumentUsecase,
	searchDocumentUsecase SearchDocumentsUsecase,
) *RagServer {
	return &RagServer{
		db:                    database,
		addDocumentUsecase:    addDocumentUsecase,
		searchDocumentUsecase: searchDocumentUsecase,
	}
}
