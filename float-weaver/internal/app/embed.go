package app

import (
	"context"
	pb "float-weaver/pkg/floatweaver"

	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fw *FloatWeaver) Embed(ctx context.Context, req *pb.EmbedRequest) (*pb.EmbedResponse, error) {
	if err := fw.validateEmbed(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	_, err := fw.embedUsecase.EmbedContent(ctx, req.GetText())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "embed error: %v", err)
	}

	return &pb.EmbedResponse{}, nil
}

func (fw *FloatWeaver) validateEmbed(req *pb.EmbedRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Text, validation.Required),
	)
}
