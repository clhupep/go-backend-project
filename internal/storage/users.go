package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"go-backend-project/internal/service"
)

type UserDB struct{ DB *sql.DB }

func NewUserDB(db *sql.DB) *UserDB { return &UserDB{DB: db} }

func (u *UserDB) CreateUser(ctx context.Context, email, hash string) (int64, error) {
	var id int64
	err := u.DB.QueryRowContext(ctx, "INSERT INTO users(email,password_hash) VALUES ($1,$2) RETURNING id", email, hash).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, service.ErrEmailExists
	}
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}
func (u *UserDB) GetByEmail(ctx context.Context, email string) (service.User, error) {
	var user service.User
	err := u.DB.QueryRowContext(ctx, "SELECT id,password_hash,role FROM users WHERE email=$1", email).Scan(&user.ID, &user.PasswordHash, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return user, service.ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}
func (u *UserDB) GetRole(ctx context.Context, id int64) (string, error) {
	var role string
	err := u.DB.QueryRowContext(ctx, "SELECT role FROM users WHERE id=$1", id).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", service.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get user role: %w", err)
	}
	return role, nil
}
