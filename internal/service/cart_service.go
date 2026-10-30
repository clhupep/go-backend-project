package service

import (
	"context"
	"errors"
	"strings"

	"go-backend-project/internal/models"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnavailable  = errors.New("product unavailable or insufficient stock")
	ErrInvalidPromo = errors.New("invalid promo code")
)

type CartRepository interface {
	Get(context.Context, int64) (models.Cart, error)
	Add(context.Context, int64, int64, int) (int64, error)
	Update(context.Context, int64, int64, int) error
	Delete(context.Context, int64, int64) error
	ApplyPromo(context.Context, int64, string) error
}

type CartService struct{ repo CartRepository }

func NewCartService(repo CartRepository) *CartService { return &CartService{repo: repo} }
func (s *CartService) Get(ctx context.Context, userID int64) (models.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return cart, err
	}
	for _, item := range cart.Items {
		cart.Subtotal += item.Price * float64(item.Quantity)
	}
	cart.Total = cart.Subtotal * float64(100-cart.DiscountPercentage) / 100
	return cart, nil
}
func (s *CartService) Add(ctx context.Context, userID, productID int64, qty int) (int64, error) {
	return s.repo.Add(ctx, userID, productID, qty)
}
func (s *CartService) Update(ctx context.Context, userID, itemID int64, qty int) error {
	return s.repo.Update(ctx, userID, itemID, qty)
}
func (s *CartService) Delete(ctx context.Context, userID, itemID int64) error {
	return s.repo.Delete(ctx, userID, itemID)
}
func (s *CartService) ApplyPromo(ctx context.Context, userID int64, code string) (models.Cart, error) {
	if err := s.repo.ApplyPromo(ctx, userID, strings.ToUpper(strings.TrimSpace(code))); err != nil {
		return models.Cart{}, err
	}
	return s.Get(ctx, userID)
}
