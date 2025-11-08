package main

import (
	"log"

	"hushzone/internal/app"
	"hushzone/internal/config"
	"hushzone/internal/db"
)

func main() {
	cfg := config.Load()
	log.Printf("DB URL (debug): %s", cfg.DBUrl)
	pool, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	r := app.Router(app.Deps{
		DB:            pool,
		AccessSecret:  cfg.JWTAccessKey,
		RefreshSecret: cfg.JWTRefreshKey,
		AccessTTL:     cfg.AccessTTL,
		RefreshTTL:    cfg.RefreshTTL,
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
