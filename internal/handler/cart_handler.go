package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-backend-project/internal/service"
)

func (h *Handler) handleCart(c *gin.Context) {
	cart, err := h.cart.Get(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		fail(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, cart)
}
func (h *Handler) handleAddItem(c *gin.Context) {
	var req struct {
		ProductID int64 `json:"product_id" binding:"required,gt=0"`
		Quantity  int   `json:"quantity" binding:"required,gt=0"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "invalid item")
		return
	}
	id, err := h.cart.Add(c.Request.Context(), c.GetInt64("userID"), req.ProductID, req.Quantity)
	if errors.Is(err, service.ErrUnavailable) {
		fail(c, 400, "product unavailable or insufficient stock")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.JSON(201, gin.H{"id": id})
}
func (h *Handler) handleUpdateItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Quantity int `json:"quantity" binding:"required,gt=0"`
	}
	if err != nil || id < 1 || c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "invalid item")
		return
	}
	err = h.cart.Update(c.Request.Context(), c.GetInt64("userID"), id, req.Quantity)
	if errors.Is(err, service.ErrNotFound) {
		fail(c, 404, "item unavailable or insufficient stock")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.Status(204)
}
func (h *Handler) handleDeleteItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		fail(c, 400, "invalid id")
		return
	}
	err = h.cart.Delete(c.Request.Context(), c.GetInt64("userID"), id)
	if errors.Is(err, service.ErrNotFound) {
		fail(c, 404, "item not found")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.Status(204)
}
func (h *Handler) handleApplyPromo(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "invalid code")
		return
	}
	cart, err := h.cart.ApplyPromo(c.Request.Context(), c.GetInt64("userID"), req.Code)
	if errors.Is(err, service.ErrInvalidPromo) {
		fail(c, 400, "invalid promo code")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.JSON(200, cart)
}
