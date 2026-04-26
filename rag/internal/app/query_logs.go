package app

import (
	"context"

	pb "rag/pkg/rag"
)

func (s *RagServer) GetQueryLogs(ctx context.Context, req *pb.GetQueryLogsRequest) (*pb.GetQueryLogsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.queryLogsUsecase.GetQueryLogs(ctx, page, pageSize)
}
