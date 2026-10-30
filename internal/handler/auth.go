package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go-backend-project/internal/service"
)

type claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
type credentials struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func (h *Handler) handleRegister(c *gin.Context) {
	var in credentials
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid email or password (8вЂ“72 bytes required)")
		return
	}
	id, err := h.users.Register(c.Request.Context(), in.Email, in.Password)
	if errors.Is(err, service.ErrInvalidPassword) {
		fail(c, 400, "invalid password")
		return
	}
	if errors.Is(err, service.ErrEmailExists) {
		fail(c, 409, "email already exists")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}
func (h *Handler) handleLogin(c *gin.Context) {
	var in credentials
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, "invalid credentials")
		return
	}
	user, err := h.users.Login(c.Request.Context(), in.Email, in.Password)
	if errors.Is(err, service.ErrUserNotFound) {
		fail(c, 404, "user not found")
		return
	}
	if errors.Is(err, service.ErrInvalidCredentials) {
		fail(c, 401, "invalid credentials")
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{UserID: user.ID, Role: user.Role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now())}})
	signed, err := token.SignedString(h.secret)
	if err != nil {
		fail(c, 500, "token error")
		return
	}
	c.JSON(200, gin.H{"token": signed})
}
func (h *Handler) auth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		fail(c, 401, "authorization required")
		c.Abort()
		return
	}
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(header, "Bearer "), &claims{}, func(t *jwt.Token) (any, error) { return h.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		fail(c, 401, "invalid token")
		c.Abort()
		return
	}
	cl, ok := token.Claims.(*claims)
	if !ok || cl.UserID <= 0 {
		fail(c, 401, "invalid token")
		c.Abort()
		return
	}
	role, err := h.users.Role(c.Request.Context(), cl.UserID)
	if errors.Is(err, service.ErrUserNotFound) {
		fail(c, 401, "invalid user")
		c.Abort()
		return
	}
	if err != nil {
		fail(c, 500, "database error")
		c.Abort()
		return
	}
	c.Set("userID", cl.UserID)
	c.Set("role", role)
	c.Next()
}
func adminOnly(c *gin.Context) {
	if c.GetString("role") != "admin" {
		fail(c, 403, "admin access required")
		c.Abort()
		return
	}
	c.Next()
}
