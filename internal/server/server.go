package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go-backend-project/internal/handler"
	"go-backend-project/internal/service"
	"go-backend-project/internal/storage"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func Run() error {
	dsn, redisURL := os.Getenv("DATABASE_URL"), os.Getenv("REDIS_URL")
	if dsn == "" || redisURL == "" {
		return errors.New("DATABASE_URL and REDIS_URL are required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("parse REDIS_URL: %w", err)
	}
	cache := redis.NewClient(opts)
	defer cache.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err = cache.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := os.Getenv("CORS_ORIGINS")
		if allowed == "" {
			allowed = "http://localhost:3000,http://localhost:5173,http://localhost:5500,http://127.0.0.1:5500"
		}
		c.Header("Vary", "Origin")
		for _, value := range strings.Split(allowed, ",") {
			if origin != "" && strings.TrimSpace(value) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	users := service.NewUserService(storage.NewUserDB(db))
	products := service.NewProductService(storage.NewProductDB(db), storage.NewProductCache(cache))
	cart := service.NewCartService(storage.NewCartDB(db))
	handler.NewHandler(users, products, cart, []byte(secret)).RegisterRoutes(router)
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if db.PingContext(ctx) != nil || cache.Ping(ctx).Err() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	return srv.ListenAndServe()
}
