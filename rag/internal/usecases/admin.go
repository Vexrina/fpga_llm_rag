package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rag/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAdminNotFound      = errors.New("admin not found")
	ErrAdminExists        = errors.New("admin already exists")
	ErrSessionExpired     = errors.New("session expired")
	ErrUnauthorized       = errors.New("unauthorized")
)

type AdminDomain struct {
	ID       int
	Username string
	Password string
	Role     string
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	Admin     *AdminDomain
}

type AdminUsecase struct {
	repo        *repository.AdminRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

func NewAdminUsecase(repo *repository.AdminRepository, jwtSecret string, tokenExpiry time.Duration) *AdminUsecase {
	return &AdminUsecase{
		repo:        repo,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: tokenExpiry,
	}
}

func (u *AdminUsecase) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	admin, err := u.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := u.repo.UpdateLastLogin(ctx, admin.ID); err != nil {
		return nil, err
	}

	token, expiresAt, err := u.generateToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		return nil, err
	}

	if err := u.repo.CreateSession(ctx, admin.ID, token, expiresAt); err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin: &AdminDomain{
			ID:       admin.ID,
			Username: admin.Username,
			Role:     admin.Role,
		},
	}, nil
}

func (u *AdminUsecase) AddAdmin(ctx context.Context, username, password string) (*AdminDomain, error) {
	existing, err := u.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAdminExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin, err := u.repo.Create(ctx, username, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return &AdminDomain{
		ID:       admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
	}, nil
}

func (u *AdminUsecase) RemoveAdmin(ctx context.Context, adminID int) error {
	admin, err := u.repo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return ErrAdminNotFound
	}

	return u.repo.Delete(ctx, adminID)
}

func (u *AdminUsecase) ValidateToken(ctx context.Context, tokenString string) (*AdminDomain, error) {
	session, err := u.repo.GetSession(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionExpired
	}

	admin, err := u.repo.GetByID(ctx, session.AdminID)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, ErrAdminNotFound
	}

	return &AdminDomain{
		ID:       admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
	}, nil
}

func (u *AdminUsecase) Logout(ctx context.Context, tokenString string) error {
	return u.repo.DeleteSession(ctx, tokenString)
}

func (u *AdminUsecase) ListAdmins(ctx context.Context) ([]*AdminDomain, error) {
	admins, err := u.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*AdminDomain, len(admins))
	for i, admin := range admins {
		result[i] = &AdminDomain{
			ID:       admin.ID,
			Username: admin.Username,
			Role:     admin.Role,
		}
	}
	return result, nil
}

func (u *AdminUsecase) LogoutAll(ctx context.Context, adminID int) error {
	sessions, err := u.repo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, admin := range sessions {
		if admin.ID == adminID {
			return nil
		}
	}
	return nil
}

func (u *AdminUsecase) generateToken(adminID int, username, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(u.tokenExpiry)

	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", adminID),
		"username": username,
		"role":     role,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
