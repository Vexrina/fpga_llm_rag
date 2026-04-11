package app

import (
	"context"
	llm_gateway "llm-gateway/pkg/llm-gateway"
)

type AskUsecase interface {
	Ask(ctx context.Context, question string) (string, error)
}

type BasePromptUpdater interface {
	UpdateBasePrompt(prompt string)
	GetBasePrompt() string
}

type Service struct {
	llm_gateway.UnimplementedGatewayServiceServer
	usecase           AskUsecase
	basePromptUpdater BasePromptUpdater
}

func NewService(usecase AskUsecase) *Service {
	return &Service{
		usecase: usecase,
	}
}

func NewServiceWithUpdater(usecase AskUsecase, updater BasePromptUpdater) *Service {
	return &Service{
		usecase:           usecase,
		basePromptUpdater: updater,
	}
}

func (s *Service) Ask(ctx context.Context, req *llm_gateway.AskRequest) (*llm_gateway.AskResponse, error) {
	answer, err := s.usecase.Ask(ctx, req.Question)
	if err != nil {
		return nil, err
	}

	return &llm_gateway.AskResponse{Answer: answer}, nil
}

func (s *Service) UpdateBasePrompt(ctx context.Context, req *llm_gateway.UpdateBasePromptRequest) (*llm_gateway.UpdateBasePromptResponse, error) {
	if s.basePromptUpdater != nil {
		s.basePromptUpdater.UpdateBasePrompt(req.Prompt)
	}
	return &llm_gateway.UpdateBasePromptResponse{Success: true}, nil
}

func (s *Service) GetBasePrompt(ctx context.Context, req *llm_gateway.GetBasePromptRequest) (*llm_gateway.GetBasePromptResponse, error) {
	if s.basePromptUpdater != nil {
		return &llm_gateway.GetBasePromptResponse{Prompt: s.basePromptUpdater.GetBasePrompt()}, nil
	}
	return &llm_gateway.GetBasePromptResponse{Prompt: ""}, nil
}
