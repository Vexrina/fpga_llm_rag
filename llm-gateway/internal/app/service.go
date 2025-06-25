package app

import (
	"context"
	llm_gateway "llm-gateway/pkg/llm-gateway"
)

type AskUsecase interface {
	Ask(ctx context.Context, question string) (string, error)
}

type Service struct {
	llm_gateway.UnimplementedGatewayServiceServer
	usecase AskUsecase
}

func NewService(usecase AskUsecase) *Service {
	return &Service{
		usecase: usecase,
	}
}

func (s *Service) Ask(ctx context.Context, req *llm_gateway.AskRequest) (*llm_gateway.AskResponse, error) {
	answer, err := s.usecase.Ask(ctx, req.Question)
	if err != nil {
		return nil, err
	}

	return &llm_gateway.AskResponse{Answer: answer}, nil
}
