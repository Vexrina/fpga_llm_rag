package app

import (
	"context"
	pb "float-weaver/pkg/floatweaver"
)

type (
	FloatWeaver struct {
		pb.UnimplementedEmbedServiceServer

		embedUsecase EmbedUsecase
	}

	EmbedUsecase interface {
		EmbedContent(ctx context.Context, content string) ([][]float32, error)
	}
)

func NewFloatWeaver(embedUsecase EmbedUsecase) *FloatWeaver {
	return &FloatWeaver{embedUsecase: embedUsecase}
}
