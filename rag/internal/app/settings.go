package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "rag/pkg/rag"
)

func (s *RagServer) GetRagSettings(ctx context.Context, req *pb.GetRagSettingsRequest) (*pb.GetRagSettingsResponse, error) {
	settings, err := s.settingsUsecase.GetRagSettings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get settings: %v", err)
	}
	return &pb.GetRagSettingsResponse{Settings: settings}, nil
}

func (s *RagServer) UpdateRagSettings(ctx context.Context, req *pb.UpdateRagSettingsRequest) (*pb.UpdateRagSettingsResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key is required")
	}
	if req.Value == "" {
		return nil, status.Errorf(codes.InvalidArgument, "value is required")
	}
	changedBy := req.ChangedBy
	if changedBy == "" {
		changedBy = "admin"
	}
	err := s.settingsUsecase.UpdateRagSetting(ctx, req.Key, req.Value, changedBy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update setting: %v", err)
	}
	return &pb.UpdateRagSettingsResponse{
		Success: true,
		Message: "Setting updated successfully",
	}, nil
}

func (s *RagServer) GetRagSettingsHistory(ctx context.Context, req *pb.GetRagSettingsHistoryRequest) (*pb.GetRagSettingsHistoryResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	history, err := s.settingsUsecase.GetSettingsHistory(ctx, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get settings history: %v", err)
	}
	return &pb.GetRagSettingsHistoryResponse{Entries: history}, nil
}
