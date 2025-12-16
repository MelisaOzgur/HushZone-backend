package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"hushzone/internal/auth"
	"hushzone/internal/measurements"
	"hushzone/internal/middleware"
	"hushzone/internal/venues"
)

type Deps struct {
	DB            *pgxpool.Pool
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func Router(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	r.POST("/v1/auth/signup",
		auth.SignUp(d.DB, d.AccessSecret, d.RefreshSecret, d.AccessTTL, d.RefreshTTL))
	r.POST("/v1/auth/signin",
		auth.SignIn(d.DB, d.AccessSecret, d.RefreshSecret, d.AccessTTL, d.RefreshTTL))
	r.POST("/v1/auth/refresh",
		auth.Refresh(d.DB, d.AccessSecret, d.RefreshSecret, d.AccessTTL, d.RefreshTTL))
	r.POST("/v1/auth/logout",
		auth.Logout(d.DB))
	r.POST("/v1/auth/google",
		auth.GoogleSignIn(d.DB, d.AccessSecret, d.RefreshSecret, d.AccessTTL, d.RefreshTTL))

	api := r.Group("/v1")
	api.Use(middleware.RequireAuth(d.AccessSecret))

	api.GET("/me", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	api.GET("/venues", venues.List(d.DB))
	api.POST("/venues", venues.Create(d.DB))

	api.POST("/measurements", measurements.Create(d.DB))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}