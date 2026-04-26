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

		db                     *repository.VecDb
		addDocumentUsecase     AddDocumentUsecase
		previewDocumentUsecase PreviewDocumentUsecase
		commitDocumentUsecase  CommitDocumentUsecase
		getDocumentUsecase     GetDocumentUsecase
		deleteDocumentUsecase  DeleteDocumentsUsecase
		searchDocumentUsecase  SearchDocumentsUsecase
		getIndexStatUsecase    GetIndexStatUsecase
		settingsUsecase        SettingsUsecase
		documentHistoryUsecase DocumentHistoryUsecase
		queryLogsUsecase       QueryLogsUsecase
	}

	AddDocumentUsecase interface {
		AddDocument(ctx context.Context, req *utils.AddDocumentDomain) error
	}
	PreviewDocumentUsecase interface {
		Preview(ctx context.Context, req *utils.PreviewDocumentDomain) (*utils.PreviewResult, error)
	}
	CommitDocumentUsecase interface {
		Commit(ctx context.Context, req *utils.CommitDocumentDomain) (string, error)
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
	SettingsUsecase interface {
		GetRagSettings(ctx context.Context) (map[string]string, error)
		UpdateRagSetting(ctx context.Context, key, value, changedBy string) error
		GetSettingsHistory(ctx context.Context, limit int) ([]*pb.SettingsHistoryEntry, error)
		GetComparisonMethod(ctx context.Context) (utils.ComparisonMethod, error)
	}
	DocumentHistoryUsecase interface {
		GetDocumentHistory(ctx context.Context, documentID string, limit int) ([]*pb.DocumentVersion, error)
		RollbackDocument(ctx context.Context, req *utils.RollbackDocumentDomain) (*pb.RollbackDocumentResponse, error)
		GetAllDocuments(ctx context.Context) ([]*pb.DocumentListItem, error)
	}
	QueryLogsUsecase interface {
		GetQueryLogs(ctx context.Context, page, pageSize int) (*pb.GetQueryLogsResponse, error)
	}
)

// NewRagServer создает новый экземпляр RagServer
func NewRagServer(
	database *repository.VecDb,
	addDocumentUsecase AddDocumentUsecase,
	previewDocumentUsecase PreviewDocumentUsecase,
	commitDocumentUsecase CommitDocumentUsecase,
	searchDocumentUsecase SearchDocumentsUsecase,
	settingsUsecase SettingsUsecase,
	documentHistoryUsecase DocumentHistoryUsecase,
	queryLogsUsecase QueryLogsUsecase,
) *RagServer {
	return &RagServer{
		db:                     database,
		addDocumentUsecase:     addDocumentUsecase,
		previewDocumentUsecase: previewDocumentUsecase,
		commitDocumentUsecase:  commitDocumentUsecase,
		searchDocumentUsecase:  searchDocumentUsecase,
		settingsUsecase:        settingsUsecase,
		documentHistoryUsecase: documentHistoryUsecase,
		queryLogsUsecase:       queryLogsUsecase,
	}
}
