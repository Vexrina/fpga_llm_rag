package app

import (
	"context"
	pb "float-weaver/pkg/floatweaver"
	"fmt"
)

type (
	FloatWeaver struct {
		pb.UnimplementedEmbedServiceServer

		embedUsecase EmbedUsecase
	}

	EmbedUsecase interface {
		EmbedContent(ctx context.Context, content string) ([][]float32, error)
		SetModel(model string)
		GetModel() string
	}
)

func NewFloatWeaver(embedUsecase EmbedUsecase) *FloatWeaver {
	return &FloatWeaver{embedUsecase: embedUsecase}
}

func (s *FloatWeaver) SetEmbeddingModel(ctx context.Context, req *pb.SetModelRequest) (*pb.SetModelResponse, error) {
	if req.Model == "" {
		return &pb.SetModelResponse{Success: false, Message: "model cannot be empty"}, nil
	}
	s.embedUsecase.SetModel(req.Model)
	return &pb.SetModelResponse{Success: true, Message: fmt.Sprintf("Model changed to: %s", req.Model)}, nil
}

func (s *FloatWeaver) GetEmbeddingModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	return &pb.GetModelResponse{Model: s.embedUsecase.GetModel()}, nil
}
