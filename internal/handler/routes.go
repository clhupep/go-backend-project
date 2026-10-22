package handler

import (
	"github.com/gin-gonic/gin"
	"go-backend-project/internal/service"
)

type Handler struct {
	users    *service.UserService
	products *service.ProductService
	secret   []byte
}

func NewHandler(users *service.UserService, products *service.ProductService, secret []byte) *Handler {
	return &Handler{users: users, products: products, secret: secret}
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
}
