package handler

import (
	"github.com/gin-gonic/gin"
	"go-backend-project/internal/service"
)

type Handler struct {
	users    *service.UserService
	products *service.ProductService
	cart     *service.CartService
	secret   []byte
}

func NewHandler(users *service.UserService, products *service.ProductService, cart *service.CartService, secret []byte) *Handler {
	return &Handler{users: users, products: products, cart: cart, secret: secret}
}
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	auth.POST("/register", h.handleRegister)
	auth.POST("/login", h.handleLogin)
	router.GET("/api/products", h.handleListProducts)
	admin := router.Group("/api/products", h.auth, adminOnly)
	admin.POST("", h.handleCreateProduct)
	admin.PUT("/:id", h.handleUpdateProduct)
	admin.DELETE("/:id", h.handleDeleteProduct)
	cart := router.Group("/api/cart", h.auth)
	cart.GET("", h.handleCart)
	cart.POST("/items", h.handleAddItem)
	cart.PATCH("/items/:id", h.handleUpdateItem)
	cart.DELETE("/items/:id", h.handleDeleteItem)
	cart.POST("/apply-promo", h.handleApplyPromo)
}
