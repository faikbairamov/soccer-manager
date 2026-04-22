// @title           Soccer Manager API
// @version         1.0
// @description     RESTful API for managing fantasy football teams.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/faikbairamov/soccer-manager/docs"
	"github.com/faikbairamov/soccer-manager/internal/config"
	"github.com/faikbairamov/soccer-manager/internal/db"
	"github.com/faikbairamov/soccer-manager/internal/handler"
	"github.com/faikbairamov/soccer-manager/internal/i18n"
	"github.com/faikbairamov/soccer-manager/internal/middleware"
	"github.com/faikbairamov/soccer-manager/internal/repository"
	"github.com/faikbairamov/soccer-manager/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	i18n.Init("locales")

	pool, err := db.Connect(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connection established")

	store := repository.NewStore(pool)

	authSvc := service.NewAuthService(store, cfg)
	teamSvc := service.NewTeamService(store)
	playerSvc := service.NewPlayerService(store)
	transferSvc := service.NewTransferService(store)

	authH := handler.NewAuthHandler(authSvc)
	teamH := handler.NewTeamHandler(teamSvc)
	playerH := handler.NewPlayerHandler(playerSvc)
	transferH := handler.NewTransferHandler(transferSvc)

	router := gin.New()

	router.Use(requestIDMiddleware())
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
	}))
	router.Use(i18n.Middleware())

	router.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authRoutes := router.Group("/api/v1/auth")
	authRoutes.Use(middleware.RateLimitMiddleware(rate.Limit(5), 10))
	{
		authRoutes.POST("/register", authH.Register)
		authRoutes.POST("/login", authH.Login)
	}

	api := router.Group("/api/v1")
	api.Use(middleware.Middleware(cfg.JWTSecret))
	{
		api.GET("/teams/me", teamH.GetTeam)
		api.PATCH("/teams/me", teamH.UpdateTeam)
		api.GET("/players/:id", playerH.GetPlayer)
		api.PATCH("/players/:id", playerH.UpdatePlayer)
		api.GET("/transfers", transferH.GetTransfers)
		api.POST("/transfers", transferH.ListTransfer)
		api.DELETE("/transfers/:id", transferH.DelistTransfer)
		api.POST("/transfers/:id/buy", transferH.BuyPlayer)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}
	slog.Info("server stopped")
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx.Set("request_id", id)
		ctx.Header("X-Request-ID", id)
		ctx.Next()
	}
}

func loggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		slog.Info("request",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", ctx.GetString("request_id"),
		)
	}
}
