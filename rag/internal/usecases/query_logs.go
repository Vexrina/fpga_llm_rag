package usecases

import (
	"context"
	"time"

	"rag/internal/repository"
	pb "rag/pkg/rag"
)

type QueryLogsRepository interface {
	GetQueryLogs(ctx context.Context, page, pageSize int) (*repository.QueryLogsResult, error)
}

type QueryLogsUsecase struct {
	repository QueryLogsRepository
}

func NewQueryLogsUsecase(repository QueryLogsRepository) *QueryLogsUsecase {
	return &QueryLogsUsecase{repository: repository}
}

func (u *QueryLogsUsecase) GetQueryLogs(ctx context.Context, page, pageSize int) (*pb.GetQueryLogsResponse, error) {
	result, err := u.repository.GetQueryLogs(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	var entries []*pb.QueryLogEntry
	for _, log := range result.Logs {
		entries = append(entries, &pb.QueryLogEntry{
			Id:             int32(log.ID),
			QueryText:      log.QueryText,
			EmbeddingModel: log.EmbeddingModel,
			ResponseTimeMs: int32(log.ResponseTimeMs),
			Found:          log.Found,
			ResultsCount:   int32(log.ResultsCount),
			CreatedAt:      formatTime(log.CreatedAt),
		})
	}

	return &pb.GetQueryLogsResponse{
		Logs:     entries,
		Total:    int32(result.Total),
		Page:     int32(result.Page),
		PageSize: int32(result.PageSize),
	}, nil
}

func formatTime(t interface{}) string {
	if t == nil {
		return ""
	}
	switch v := t.(type) {
	case string:
		return v
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	default:
		return ""
	}
}
