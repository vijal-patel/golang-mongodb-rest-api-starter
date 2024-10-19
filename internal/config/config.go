package config

import (
	"fmt"
	"golang-mongodb-rest-api-starter/internal/constants"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Auth    AuthConfig
	Captcha CaptchaConfig
	Cache   CacheConfig
	DB      DBConfig
	Email   EmailConfig
	Env     string
	HTTP    HTTPConfig
}

func NewConfig() *Config {
	env := os.Getenv("ENV")
	if env == constants.EmptyString {
		env = "dev"
	}

	// Load .env file for env
	if err := godotenv.Load(fmt.Sprintf("%s/.env.%s", os.Getenv("KO_DATA_PATH"), env)); err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		Auth:    LoadAuthConfig(),
		Captcha: LoadCaptchaConfig(),
		Cache:   LoadCacheConfig(),
		DB:      LoadDBConfig(),
		Email:   LoadEmailConfig(),
		Env:     env,
		HTTP:    LoadHTTPConfig(),
	}
}
