package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Admin struct {
	ID        int
	Username  string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin *time.Time
}

type AdminSession struct {
	ID        int
	AdminID   int
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AdminRepository struct {
	conn *pgxpool.Pool
}

func NewAdminRepository(conn *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{conn: conn}
}

func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (*Admin, error) {
	var admin Admin
	err := r.conn.QueryRow(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at, last_login
		FROM admins
		WHERE username = $1
	`, username).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.Role, &admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) GetByID(ctx context.Context, id int) (*Admin, error) {
	var admin Admin
	err := r.conn.QueryRow(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at, last_login
		FROM admins
		WHERE id = $1
	`, id).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.Role, &admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) Create(ctx context.Context, username, passwordHash string) (*Admin, error) {
	var admin Admin
	err := r.conn.QueryRow(ctx, `
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, role, created_at, updated_at, last_login
	`, username, passwordHash).Scan(&admin.ID, &admin.Username, &admin.Password, &admin.Role, &admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) Delete(ctx context.Context, id int) error {
	_, err := r.conn.Exec(ctx, `DELETE FROM admins WHERE id = $1`, id)
	return err
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, id int) error {
	_, err := r.conn.Exec(ctx, `UPDATE admins SET last_login = NOW() WHERE id = $1`, id)
	return err
}

func (r *AdminRepository) CreateSession(ctx context.Context, adminID int, token string, expiresAt time.Time) error {
	_, err := r.conn.Exec(ctx, `
		INSERT INTO admin_sessions (admin_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, adminID, token, expiresAt)
	return err
}

func (r *AdminRepository) GetSession(ctx context.Context, token string) (*AdminSession, error) {
	var session AdminSession
	err := r.conn.QueryRow(ctx, `
		SELECT id, admin_id, token, expires_at, created_at
		FROM admin_sessions
		WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&session.ID, &session.AdminID, &session.Token, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *AdminRepository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.conn.Exec(ctx, `DELETE FROM admin_sessions WHERE token = $1`, token)
	return err
}

func (r *AdminRepository) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.conn.Exec(ctx, `DELETE FROM admin_sessions WHERE expires_at < NOW()`)
	return err
}

func (r *AdminRepository) ListAll(ctx context.Context) ([]Admin, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at, last_login
		FROM admins
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []Admin
	for rows.Next() {
		var admin Admin
		if err := rows.Scan(&admin.ID, &admin.Username, &admin.Password, &admin.Role, &admin.CreatedAt, &admin.UpdatedAt, &admin.LastLogin); err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}
	return admins, nil
}
