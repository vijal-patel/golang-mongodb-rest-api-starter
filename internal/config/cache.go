package config

import (
	"log"
	"os"
	"strconv"
)

type CacheConfig struct {
	User     string
	Password string
	DB       int
	Host     string
}

func LoadCacheConfig() CacheConfig {
	db, err := strconv.Atoi(os.Getenv("CACHE_NAME"))
	if err != nil {
		log.Fatal("Invalid cache config")
	}
	return CacheConfig{
		User:     os.Getenv("CACHE_USER"),
		Password: os.Getenv("CACHE_PASSWORD"),
		DB:       db,
		Host:     os.Getenv("CACHE_HOST"),
	}
}
