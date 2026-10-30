package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go-backend-project/internal/models"
	"go-backend-project/internal/service"

	"github.com/gin-gonic/gin"
)

func positive(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return n, nil
}
func (h *Handler) handleListProducts(c *gin.Context) {
	q := c.Request.URL.Query()
	var f models.ProductFilter
	var err error
	f.Page, err = positive(q.Get("page"), 1)
	if err != nil || f.Page > 1000000 {
		fail(c, 400, "invalid page")
		return
	}
	f.Limit, err = positive(q.Get("limit"), 12)
	if err != nil || f.Limit > 100 {
		fail(c, 400, "invalid limit")
		return
	}
	f.Search = q.Get("search")
	f.SortBy = q.Get("sort_by")
	if f.SortBy == "" {
		f.SortBy = "title"
	}
	f.Order = strings.ToLower(q.Get("order"))
	if f.Order == "" {
		f.Order = "asc"
	}
	if (f.SortBy != "title" && f.SortBy != "price") || (f.Order != "asc" && f.Order != "desc") {
		fail(c, 400, "invalid sort")
		return
	}
	for _, field := range []struct {
		name string
		dest **float64
	}{{"min_price", &f.MinPrice}, {"max_price", &f.MaxPrice}, {"min_rating", &f.MinRating}} {
		if value := q.Get(field.name); value != "" {
			n, e := strconv.ParseFloat(value, 64)
			if e != nil || n < 0 || n != n || n > 1e12 {
				fail(c, 400, "invalid "+field.name)
				return
			}
			*field.dest = &n
		}
	}
	page, err := h.products.List(c.Request.Context(), f, q.Encode())
	if errors.Is(err, service.ErrCache) {
		fail(c, 503, "cache unavailable")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.JSON(200, page)
}
func (h *Handler) handleCreateProduct(c *gin.Context) {
	var req models.ProductInput
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "invalid product")
		return
	}
	id, err := h.products.Create(c.Request.Context(), req)
	if errors.Is(err, service.ErrInvalidProduct) {
		fail(c, 400, "invalid product")
		return
	}
	if errors.Is(err, service.ErrCache) {
		fail(c, 503, "product created; cache invalidation failed")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.JSON(201, gin.H{"id": id})
}
func (h *Handler) handleUpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		fail(c, 400, "invalid id")
		return
	}
	var req models.ProductInput
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "invalid product")
		return
	}
	err = h.products.Update(c.Request.Context(), id, req)
	if errors.Is(err, service.ErrInvalidProduct) {
		fail(c, 400, "invalid product")
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		fail(c, 404, "product not found")
		return
	}
	if errors.Is(err, service.ErrCache) {
		fail(c, 503, "product updated; cache invalidation failed")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.Status(204)
}
func (h *Handler) handleDeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		fail(c, 400, "invalid id")
		return
	}
	err = h.products.Delete(c.Request.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		fail(c, 404, "product not found")
		return
	}
	if errors.Is(err, service.ErrCache) {
		fail(c, 503, "product deleted; cache invalidation failed")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.Status(204)
}
