package service

import (
	"context"
	"errors"
	"fmt"
	"go-backend-project/internal/models"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
)

type UserRepository interface {
	CreateUser(context.Context, string, string) (int64, error)
	GetByEmail(context.Context, string) (models.User, error)
	GetRole(context.Context, int64) (string, error)
}

type UserService struct{ users UserRepository }

func NewUserService(users UserRepository) *UserService { return &UserService{users: users} }

func (s *UserService) Register(ctx context.Context, email, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, ErrInvalidPassword
	}
	id, err := s.users.CreateUser(ctx, strings.ToLower(strings.TrimSpace(email)), string(hash))
	if err != nil {
		return 0, fmt.Errorf("register: %w", err)
	}
	return id, nil
}
func (s *UserService) Login(ctx context.Context, email, password string) (models.User, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, ErrUserNotFound) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("login: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return models.User{}, ErrInvalidCredentials
	}
	return user, nil
}
func (s *UserService) Role(ctx context.Context, id int64) (string, error) {
	return s.users.GetRole(ctx, id)
}
