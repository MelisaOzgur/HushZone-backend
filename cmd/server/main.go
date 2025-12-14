package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"hushzone/internal/app"
	"hushzone/internal/config"
	"hushzone/internal/db"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := app.Router(app.Deps{
		DB:            pool,
		AccessSecret:  cfg.JWTAccessKey,
		RefreshSecret: cfg.JWTRefreshKey,
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}