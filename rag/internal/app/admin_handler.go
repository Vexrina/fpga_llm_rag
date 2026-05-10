package app

import (
	"context"
	"time"

	"rag/internal/usecases"
	pb "rag/pkg/rag"
)

type AdminUsecase interface {
	Login(ctx context.Context, username, password string) (*usecases.LoginResult, error)
	AddAdmin(ctx context.Context, username, password string) (*usecases.AdminDomain, error)
	RemoveAdmin(ctx context.Context, adminID int) error
	ValidateToken(ctx context.Context, token string) (*usecases.AdminDomain, error)
	Logout(ctx context.Context, token string) error
	ListAdmins(ctx context.Context) ([]*usecases.AdminDomain, error)
}

type AdminHandler struct {
	adminUsecase AdminUsecase
}

func NewAdminHandler(adminUsecase AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: adminUsecase}
}

func (h *AdminHandler) AdminLogin(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	result, err := h.adminUsecase.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return &pb.AdminLoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.AdminLoginResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		Admin: &pb.AdminInfo{
			Id:       int32(result.Admin.ID),
			Username: result.Admin.Username,
			Role:     result.Admin.Role,
		},
		Success: true,
		Message: "Login successful",
	}, nil
}

func (h *AdminHandler) AdminLogout(ctx context.Context, req *pb.AdminLogoutRequest) (*pb.AdminLogoutResponse, error) {
	err := h.adminUsecase.Logout(ctx, req.GetToken())
	if err != nil {
		return &pb.AdminLogoutResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.AdminLogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}

func (h *AdminHandler) AddAdmin(ctx context.Context, req *pb.AddAdminRequest) (*pb.AddAdminResponse, error) {
	_, err := h.adminUsecase.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &pb.AddAdminResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	admin, err := h.adminUsecase.AddAdmin(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return &pb.AddAdminResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.AddAdminResponse{
		Success: true,
		Message: "Admin added successfully",
		Admin: &pb.AdminInfo{
			Id:       int32(admin.ID),
			Username: admin.Username,
			Role:     admin.Role,
		},
	}, nil
}

func (h *AdminHandler) RemoveAdmin(ctx context.Context, req *pb.RemoveAdminRequest) (*pb.RemoveAdminResponse, error) {
	_, err := h.adminUsecase.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &pb.RemoveAdminResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	err = h.adminUsecase.RemoveAdmin(ctx, int(req.GetAdminId()))
	if err != nil {
		return &pb.RemoveAdminResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RemoveAdminResponse{
		Success: true,
		Message: "Admin removed successfully",
	}, nil
}

func (h *AdminHandler) ListAdmins(ctx context.Context, req *pb.ListAdminsRequest) (*pb.ListAdminsResponse, error) {
	_, err := h.adminUsecase.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &pb.ListAdminsResponse{
			Admins: nil,
		}, nil
	}

	admins, err := h.adminUsecase.ListAdmins(ctx)
	if err != nil {
		return &pb.ListAdminsResponse{
			Admins: nil,
		}, nil
	}

	result := make([]*pb.AdminInfo, len(admins))
	for i, admin := range admins {
		result[i] = &pb.AdminInfo{
			Id:       int32(admin.ID),
			Username: admin.Username,
			Role:     admin.Role,
		}
	}

	return &pb.ListAdminsResponse{
		Admins: result,
	}, nil
}

func (h *AdminHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	admin, err := h.adminUsecase.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		Admin: &pb.AdminInfo{
			Id:       int32(admin.ID),
			Username: admin.Username,
			Role:     admin.Role,
		},
	}, nil
}
