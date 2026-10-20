package handler

import (
	"github.com/gin-gonic/gin"
	"go-backend-project/internal/service"
)

type Handler struct {
	users  *service.UserService
	secret []byte
}

func NewHandler(users *service.UserService, secret []byte) *Handler {
	return &Handler{users: users, secret: secret}
}
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	auth.POST("/register", h.handleRegister)
	auth.POST("/login", h.handleLogin)
}
