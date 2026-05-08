package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rag/internal/utils"
	pb "rag/pkg/rag"
)

type GetDocumentIdsUsecase interface {
	GetDocumentIds(ctx context.Context, req *utils.GetDocumentIdsDomain) (*pb.GetDocumentIdsResponse, error)
}

func (s *RagServer) GetDocumentIdsByQuery(ctx context.Context, req *pb.GetDocumentIdsRequest) (*pb.GetDocumentIdsResponse, error) {
	if req.Query == "" {
		return nil, status.Errorf(codes.InvalidArgument, "query is required")
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	domainReq := &utils.GetDocumentIdsDomain{
		Query:            req.Query,
		Limit:            req.Limit,
		SimilarityThs:    req.SimilarityThreshold,
		ComparisonMethod: utils.ComparisonMethod(req.ComparisonMethod.String()),
	}

	return s.getDocumentIdsUsecase.GetDocumentIds(ctx, domainReq)
}
