package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"go-backend-project/internal/models"
)

var ErrInvalidProduct = errors.New("invalid product")
var ErrCache = errors.New("cache unavailable")

type ProductRepository interface {
	List(context.Context, models.ProductFilter) (models.ProductPage, error)
	Create(context.Context, models.ProductInput) (int64, error)
	Update(context.Context, int64, models.ProductInput) error
	Delete(context.Context, int64) error
}
type CatalogCache interface {
	Version(context.Context) (string, error)
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte) error
	Invalidate(context.Context) error
}
type ProductService struct {
	repo  ProductRepository
	cache CatalogCache
}

func NewProductService(repo ProductRepository, cache CatalogCache) *ProductService {
	return &ProductService{repo: repo, cache: cache}
}
func (s *ProductService) List(ctx context.Context, f models.ProductFilter, key string) (models.ProductPage, error) {
	var page models.ProductPage
	if (f.SortBy != "title" && f.SortBy != "price") || (f.Order != "asc" && f.Order != "desc") || f.Limit < 1 || f.Page < 1 {
		return page, ErrInvalidProduct
	}
	version, err := s.cache.Version(ctx)
	if err != nil {
		return page, ErrCache
	}
	key = "catalog:" + version + ":" + key
	if data, e := s.cache.Get(ctx, key); e == nil && json.Unmarshal(data, &page) == nil {
		return page, nil
	}
	page, err = s.repo.List(ctx, f)
	if err != nil {
		return page, err
	}
	if data, e := json.Marshal(page); e == nil {
		_ = s.cache.Set(ctx, key, data)
	}
	return page, nil
}
func validProduct(v models.ProductInput) bool {
	if strings.TrimSpace(v.Title) == "" || v.Price == nil || v.Rating == nil || v.Stock == nil || len(v.Images) < 3 {
		return false
	}
	for _, value := range v.Images {
		u, err := url.ParseRequestURI(value)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return false
		}
	}
	return true
}
func (s *ProductService) Create(ctx context.Context, v models.ProductInput) (int64, error) {
	if !validProduct(v) {
		return 0, ErrInvalidProduct
	}
	id, err := s.repo.Create(ctx, v)
	if err != nil {
		return 0, err
	}
	if err = s.cache.Invalidate(ctx); err != nil {
		return id, ErrCache
	}
	return id, nil
}
func (s *ProductService) Update(ctx context.Context, id int64, v models.ProductInput) error {
	if !validProduct(v) {
		return ErrInvalidProduct
	}
	if err := s.repo.Update(ctx, id, v); err != nil {
		return err
	}
	if err := s.cache.Invalidate(ctx); err != nil {
		return ErrCache
	}
	return nil
}
func (s *ProductService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.cache.Invalidate(ctx); err != nil {
		return ErrCache
	}
	return nil
}
