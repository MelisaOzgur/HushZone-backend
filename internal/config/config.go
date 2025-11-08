package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl         string
	JWTAccessKey  string
	JWTRefreshKey string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func Load() Config {
	_ = godotenv.Load() 

	return Config{
		DBUrl:         mustEnv("DATABASE_URL"),
		JWTAccessKey:  mustEnv("JWT_ACCESS_SECRET"),
		JWTRefreshKey: mustEnv("JWT_REFRESH_SECRET"),
		AccessTTL:     minutesEnv("ACCESS_TTL_MINUTES", 15),
		RefreshTTL: minutesEnv("REFRESH_TTL_DAYS", 30*24*60),
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}

func minutesEnv(k string, def int) time.Duration {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Minute
		}
	}
	return time.Duration(def) * time.Minute
}
